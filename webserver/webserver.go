// Teal.Finance/Garcon is an opinionated boilerplate API and website server.
// Copyright (C) 2021 Teal.Finance contributors
//
// This file is part of Teal.Finance/Garcon, licensed under LGPL-3.0-or-later.
//
// Teal.Finance/Garcon is free software: you can redistribute it
// and/or modify it under the terms of the GNU Lesser General Public License
// either version 3 of the License, or (at your option) any later version.
//
// Teal.Finance/Garcon is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty
// of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.

package webserver

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/teal-finance/garcon/reserr"
)

type WebServer struct {
	Dir    string
	ResErr reserr.ResErr
}

// ServeFile handles one specific file (and its specific Content-Type).
func (ws WebServer) ServeFile(urlPath, contentType string) func(w http.ResponseWriter, r *http.Request) {
	absPath := path.Join(ws.Dir, urlPath)

	return func(w http.ResponseWriter, r *http.Request) {
		// Set aggressive "Cache-Control" because ServeFile() is often used
		// to serve "favicon.ico" and other assets that do not change often
		w.Header().Set("Cache-Control", "public,max-age=31536000,immutable")
		w.Header().Set("Content-Type", contentType)

		ws.send(w, r, absPath)
	}
}

// ServeDir handles the static files using the same Content-Type.
func (ws WebServer) ServeDir(contentType string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if validPath(w, r) {
			// JS and CSS files should contain a [hash].
			// Thus the path changes when content changes,
			// enabling aggressive Cache-Control parameters:
			// public            Can be cached by proxy (reverse-proxy. CDN…) and by browser
			// max-age=31536000  Store it up to 1 year (browser stores it some days due to limited cache size)
			// immutable         Only supported by Firefox and Safari
			w.Header().Set("Cache-Control", "public,max-age=31536000,immutable")
			w.Header().Set("Content-Type", contentType)

			absPath := path.Join(ws.Dir, r.URL.Path)
			ws.send(w, r, absPath)
		}
	}
}

// ServeImages detects the Content-Type depending on the image extension.
func (ws WebServer) ServeImages() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if validPath(w, r) {
			// Images are supposed never change, else better to create a new image
			// (or to wait some days the browser clears out data based on LRU).
			w.Header().Set("Cache-Control", "public,max-age=31536000,immutable")

			absPath, contentType := ws.imagePathAndType(r)
			if contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}

			ws.send(w, r, absPath)
		}
	}
}

// validPath returns a HTTP error if the path is invalid.
func validPath(w http.ResponseWriter, r *http.Request) bool {
	if strings.Contains(r.URL.Path, "..") {
		reserr.Write(w, r, http.StatusBadRequest, "Invalid URL Path Containing '..'")
		log.Print("WRN WebServer: reject path with '..' ", r.URL.Path)

		return false
	}

	return true
}

func (ws WebServer) send(w http.ResponseWriter, r *http.Request, absPath string) {
	var (
		file *os.File
		err  error
	)

	// if client (browser) supports Brotli and the *.br file is present
	// => send the *.br file
	if strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
		brotli := absPath + ".br"

		file, err = os.Open(brotli)
		if err == nil {
			w.Header().Set("Content-Encoding", "br")

			absPath = brotli
		}
	}

	if file == nil {
		file, err = os.Open(absPath)
		if err != nil {
			ws.ResErr.Write(w, r, http.StatusNotFound, "Page not found")
			log.Print("WRN WebServer: ", err)

			return
		}
	}

	defer func() {
		if e := file.Close(); e != nil {
			log.Print("WRN WebServer: Close() ", e)
		}
	}()

	fi, err := file.Stat()
	if err != nil {
		ws.ResErr.Write(w, r, http.StatusInternalServerError, "Internal Server Error")
		log.Print("WRN WebServer: Stat(", absPath, ") ", err)

		return
	}

	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	w.Header().Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))
	// We do not manage PartialContent because too much stuff
	// to handle the headers Range If-Range Etag and Content-Range.

	if n, err := io.Copy(w, file); err != nil {
		log.Print("WRN WebServer: Copy(", absPath, ") ", err)
	} else {
		log.Print("WebServer sent ", absPath, " ", IEC64(n))
	}
}

// IEC64 converts bytes into KiB (1024 bytes), MiB, GiB…
// as defined within the International System of Quantities (ISQ)
// standardized by the ISO/IEC 80000 and published in 2008.
func IEC64(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// imagePathAndType returns the path/filename and the Content-Type of the image.
// If the client (browser) supports AVIF, imagePathAndType replaces the requested image by the AVIF one.
func (ws WebServer) imagePathAndType(r *http.Request) (absPath, contentType string) {
	extPos := extIndex(r.URL.Path)

	// We only check the first Header "Accept":
	// We do not care to miss an "image/avif" within the second Header "Accept",
	// because we do not break anything: we send the image requested by the client.
	scheme := r.Header.Get("Accept")

	// We perform a stupid search to be fast,
	// but we hope there is no Content-Type such as "image/avifuck"
	const avifContentType = "image/avif"
	if strings.Contains(scheme, avifContentType) {
		avifPath := r.URL.Path[:extPos] + "avif"
		absPath = path.Join(ws.Dir, avifPath)

		_, err := os.Stat(absPath)
		if err == nil {
			return absPath, avifContentType
		}

		log.Printf("WRN WebServer supports Content-Type=%q "+
			"but cannot access %q %v", avifContentType, absPath, err)
	}

	absPath = path.Join(ws.Dir, r.URL.Path)

	ext := r.URL.Path[extPos:]
	contentType = imageContentType(ext)

	return absPath, contentType
}

// extIndex returns the position of the extension within the the urlPath.
// If no dot, returns the ending position.
func extIndex(urlPath string) int {
	for i := len(urlPath) - 1; i >= 0 && urlPath[i] != '/'; i-- {
		if urlPath[i] == '.' {
			return i + 1
		}
	}

	return len(urlPath)
}

// imageContentType determines the Content-Type depending on the file extension.
func imageContentType(ext string) (contentType string) {
	// Only the most popular image extensions
	switch ext {
	case "png":
		return "image/png"
	case "jpg":
		return "image/jpeg"
	case "svg":
		return "image/svg+xml"
	default:
		log.Print("WRN WebServer does not support image extension: ", ext)

		return ""
	}
}

// Extension  MIME type
// ---------  --------------------------------
//  .html     text/html; charset=utf-8
//  .css      text/css; charset=utf-8
//  .csv      text/csv; charset=utf-8
//  .xml      text/xml; charset=utf-8
//  .js       text/javascript; charset=utf-8
//  .md       text/markdown; charset=utf-8
//  .yaml     text/x-yaml; charset=utf-8
//  .json     application/json; charset=utf-8
//  .pdf      application/pdf
//  .woff2    font/woff2
//  .avif     image/avif
//  .gif      image/gif
//  .ico      image/x-icon
//  .jpg      image/jpeg
//  .png      image/png
//  .svg      image/svg+xml
//  .webp     image/webp
// #region <editor-fold desc="Preamble">
// Copyright (c) 2021-2022 Teal.Finance contributors
//
// This file is part of Teal.Finance/Garcon, an API and website server.
// Teal.Finance/Garcon is free software: you can redistribute it
// and/or modify it under the terms of the GNU Lesser General Public License
// either version 3 or any later version, at the licensee’s option.
// SPDX-License-Identifier: LGPL-3.0-or-later
//
// Teal.Finance/Garcon is distributed WITHOUT ANY WARRANTY.
// For more details, see the LICENSE file (alongside the source files)
// or online at <https://www.gnu.org/licenses/lgpl-3.0.html>
// #endregion </editor-fold>

package cors

import (
	"log"
	"net/http"
	"strings"

	"github.com/rs/cors"
)

// Handler uses restrictive CORS values.
func Handler(origins []string, debug bool) func(next http.Handler) http.Handler {
	options := cors.Options{
		AllowedOrigins:         []string{},
		AllowOriginFunc:        nil,
		AllowOriginRequestFunc: nil,
		AllowedMethods:         []string{http.MethodGet, http.MethodPost},
		AllowedHeaders:         []string{"Origin", "Accept", "Content-Type", "Authorization", "Cookie"},
		ExposedHeaders:         []string{},
		MaxAge:                 24 * 3600, // https://developer.mozilla.org/docs/Web/HTTP/Headers/Access-Control-Max-Age
		AllowCredentials:       true,
		OptionsPassthrough:     false,
		OptionsSuccessStatus:   http.StatusNoContent,
		Debug:                  debug, // verbose logs
	}

	InsertSchema(origins)

	if len(origins) == 1 {
		options.AllowOriginFunc = oneOrigin(origins[0])
	} else {
		options.AllowOriginFunc = multipleOriginPrefixes(origins)
	}

	log.Printf("CORS: Methods=%v Headers=%v Credentials=%v MaxAge=%v",
		options.AllowedMethods, options.AllowedHeaders, options.AllowCredentials, options.MaxAge)

	return cors.New(options).Handler
}

func InsertSchema(origins []string) {
	for i, o := range origins {
		if !strings.HasPrefix(o, "https://") &&
			!strings.HasPrefix(o, "http://") {
			origins[i] = "http://" + o
		}
	}
}

func oneOrigin(addr string) func(string) bool {
	log.Print("CORS: Set one origin: ", addr)

	return func(origin string) bool {
		return origin == addr
	}
}

func multipleOriginPrefixes(addrPrefixes []string) func(origin string) bool {
	log.Print("CORS: Set origin prefixes: ", addrPrefixes)

	return func(origin string) bool {
		for _, prefix := range addrPrefixes {
			if strings.HasPrefix(origin, prefix) {
				return true
			}
		}

		log.Print("CORS: Refuse ", origin, " without prefixes ", addrPrefixes)

		return false
	}
}

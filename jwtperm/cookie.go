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

// Package jwtperm delivers and checks the JWT permissions
package jwtperm

import (
	"context"
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/teal-finance/garcon/reserr"
	"github.com/teal-finance/garcon/security"
	"github.com/teal-finance/quid/quidlib/tokens"
)

const (
	// authScheme is part of the HTTP "Authorization" header
	// conveying the "Bearer Token" definitioned by RFC 6750 as
	// as security token with the property that any party in possession of
	// the token (a "bearer") can use the token in any way that any other
	// party in possession of it can.  Using a bearer token does not
	// require a bearer to prove possession of cryptographic key material
	// (proof-of-possession).
	authScheme = "Bearer "

	defaultCookieName = "g" // g as in garcon
	defaultPlanName   = "DefaultPlan"
	defaultPermValue  = 3600     // one hour
	oneYearInSeconds  = 31556952 // average including leap years
	oneYearInNS       = oneYearInSeconds * 1_000_000_000
)

var (
	ErrUnauthorized  = errors.New("JWT not authorized")
	ErrNoTokenFound  = errors.New("no JWT found")
	ErrInvalidCookie = errors.New("invalid cookie")
	ErrExpiredToken  = errors.New("expired or invalid refresh token")
)

type Perm struct {
	Value int
}

type Checker struct {
	resErr      reserr.ResErr
	b64encoding *base64.Encoding
	secretKey   []byte
	perms       []Perm
	plans       []string
	cookies     []http.Cookie
	devOrigins  []string
}

func New(urls []*url.URL, resErr reserr.ResErr, secretKey []byte, permissions ...interface{}) *Checker {
	n := len(permissions) / 2
	if n == 0 {
		n = 1
	}

	names := make([]string, n)
	values := make([]int, n)

	names[0] = defaultPlanName
	values[0] = defaultPermValue

	for i, p := range permissions {
		var ok bool
		if i%2 == 0 {
			names[i/2], ok = p.(string)
		} else {
			values[i/2], ok = p.(int)
		}

		if !ok {
			log.Panic("Wrong type for the parametric arguments in jwtperm.New(), " +
				"must alternate string and int: plan1, perm1, plan2, perm2...")
		}
	}

	secure, dns, path := extractMainDomain(urls)
	perms := make([]Perm, n)
	cookies := make([]http.Cookie, n)

	for i, v := range values {
		perms[i] = Perm{Value: v}
		cookies[i] = createCookie(names[i], secure, dns, path, secretKey)
	}

	return &Checker{
		resErr:      resErr,
		b64encoding: base64.RawURLEncoding,
		secretKey:   secretKey,
		plans:       names,
		perms:       perms,
		cookies:     cookies,
		devOrigins:  extractDevOrigins(urls),
	}
}

const (
	HTTP  = "http"
	HTTPS = "https"
)

func extractMainDomain(urls []*url.URL) (secure bool, dns, path string) {
	if len(urls) == 0 {
		log.Panic("No urls => Cannot set Cookie domain")
	}

	u := urls[0]
	if u == nil {
		log.Panic("Unexpected nil in URL slide: ", urls)
	}

	switch {
	case u.Scheme == HTTP:
		secure = false

	case u.Scheme == HTTPS:
		secure = true

	default:
		log.Panic("Unexpected scheme in ", u)
	}

	return secure, u.Hostname(), u.Path
}

func extractDevURLs(urls []*url.URL) (devURLs []*url.URL) {
	if len(urls) == 1 {
		log.Print("JWT required for single domain: ", urls)
		return nil
	}

	for i, u := range urls {
		if u == nil {
			log.Panic("Unexpected nil in URL slide: ", urls)
		}
		if u.Scheme == HTTP {
			return urls[i:]
		}
	}

	return nil
}

func extractDevOrigins(urls []*url.URL) (devOrigins []string) {
	if len(urls) > 0 && urls[0].Scheme == "http" {
		host, _, _ := net.SplitHostPort(urls[0].Host)
		if host == "localhost" {
			return []string{"*"} // Accept absence of cookie for http://localhost
		}
	}

	devURLS := extractDevURLs(urls)

	if len(devURLS) == 0 {
		return nil
	}

	devOrigins = make([]string, 0, len(urls))

	for _, u := range urls {
		o := u.Scheme + "://" + u.Host
		devOrigins = append(devOrigins, o)
	}

	log.Print("JWT not required for dev. origins: ", devOrigins)
	return devOrigins
}

func createCookie(plan string, secure bool, dns, path string, secretKey []byte) http.Cookie {
	if len(secretKey) < 32 {
		log.Panic("Want HMAC-SHA256 key containing 32 bytes (or more), but got ", len(secretKey))
	}

	jwt, err := tokens.GenRefreshToken("1y", "1y", plan, "", secretKey)
	if err != nil || jwt == "" {
		log.Panic("Cannot create JWT: ", err)
	}

	name := defaultCookieName

	if path != "" {
		// remove trailing slash
		if path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}

		for i := len(path) - 1; i >= 0; i-- {
			if path[i] == byte('/') {
				name = path[i+1:]
				break
			}
		}
	}

	log.Print("Create cookie plan=", plan, " domain=", dns, " secure=", secure, " ", name, "=", jwt)

	return http.Cookie{
		Name:       name,
		Value:      jwt,
		Path:       path,
		Domain:     dns,
		Expires:    time.Time{},
		RawExpires: "",
		MaxAge:     oneYearInSeconds,
		Secure:     secure,
		HttpOnly:   true,
		SameSite:   http.SameSiteStrictMode,
		Raw:        "",
		Unparsed:   nil,
	}
}

// Set puts a HttpOnly cookie when no valid cookie is present in the HTTP response header.
// The permission conveyied by te cookie is also put in the request context.
func (ck *Checker) Set(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		perm, ok := ck.getPerm(r)
		if !ok {
			perm = ck.perms[0]
			ck.cookies[0].Expires = time.Now().Add(oneYearInNS)
			http.SetCookie(w, &ck.cookies[0])
		}

		next.ServeHTTP(w, perm.putInCtx(r))
	})
}

// Chk accepts the HTTP request only if it contains a valid cookie.
func (ck *Checker) Chk(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		perm, err := ck.permFromCookie(r)
		if err != nil {
			if ck.isDevOrigin(r) {
				perm = ck.perms[0]
			} else {
				ck.resErr.Write(w, r, http.StatusUnauthorized, err.Error())
				return
			}
		}

		next.ServeHTTP(w, perm.putInCtx(r))
	})
}

// Vet accepts the HTTP request only if a valid JWT
// is in the cookie or in the first "Authorization" header.
func (ck *Checker) Vet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		perm, err := ck.permFromBearerOrCookie(r)
		if err != nil {
			if ck.isDevOrigin(r) {
				perm = ck.perms[0]
			} else {
				ck.resErr.Write(w, r, http.StatusUnauthorized, err.Error())
				return
			}
		}

		next.ServeHTTP(w, perm.putInCtx(r))
	})
}

func (ck *Checker) isDevOrigin(r *http.Request) bool {
	if len(ck.devOrigins) == 0 {
		return false
	}

	if len(ck.devOrigins) > 0 {
		origin := r.Header.Get("Origin")
		for _, prefix := range ck.devOrigins {
			if prefix == "*" {
				return true
			}
			if strings.HasPrefix(origin, prefix) {
				return true
			}
		}
	}

	return false
}

func (ck *Checker) getPerm(r *http.Request) (perm Perm, ok bool) {
	cookie, err := r.Cookie(ck.cookies[0].Name)
	if err != nil {
		return perm, false
	}

	for i, c := range ck.cookies {
		if cookie.Value == c.Value {
			return ck.perms[i], true
		}
	}

	perm, err = ck.permFromJWT(cookie.Value)
	return perm, (err == nil)
}

func (ck *Checker) permFromBearerOrCookie(r *http.Request) (perm Perm, err error) {
	jwt, errBearer := ck.jwtFromBearer(r)
	if errBearer != nil {
		jwt, err = ck.jwtFromCookie(r)
		if err != nil {
			err = fmt.Errorf("cannot find a valid JWT in either "+
				"the first 'Authorization' HTTP header or "+
				"in the cookie %q because: %w and %v",
				ck.cookies[0].Name, errBearer, err.Error())
			return perm, err
		}
	}
	return ck.permFromJWT(jwt)
}

func (ck *Checker) permFromCookie(r *http.Request) (perm Perm, err error) {
	jwt, err := ck.jwtFromCookie(r)
	if err != nil {
		return perm, err
	}
	return ck.permFromJWT(jwt)
}

func (ck *Checker) jwtFromBearer(r *http.Request) (jwt string, err error) {
	auth := r.Header.Get("Authorization")

	n := len(authScheme)
	if len(auth) > n && auth[:n] == authScheme {
		return auth[n:], nil
	}

	if auth == "" {
		return "", errors.New("Provide your JWT within the 'Authorization Bearer' HTTP header")
	}

	return "", ErrInvalidCookie
}

func (ck *Checker) jwtFromCookie(r *http.Request) (jwt string, err error) {
	c, err := r.Cookie(ck.cookies[0].Name)
	if err != nil {
		return "", errors.New("visit the official " +
			ck.cookies[0].Domain + " web site to get a valid Cookie")
	}
	return c.Value, nil
}

func (ck *Checker) permFromJWT(jwt string) (perm Perm, err error) {
	for i, c := range ck.cookies {
		if c.Value == jwt {
			return ck.perms[i], nil
		}
	}

	parts, err := ck.partsFromJWT(jwt)
	if err != nil {
		return perm, err
	}

	perm, err = ck.permFromRefreshBytes(parts)
	if err != nil {
		return perm, ErrExpiredToken
	}

	return perm, nil
}

func (ck *Checker) permFromRefreshClaims(claims *tokens.RefreshClaims) Perm {
	for i, p := range ck.plans {
		if p == claims.Namespace {
			return ck.perms[i]
		}
	}

	return ck.perms[0]
}

func (ck *Checker) decomposeJWT(jwt string) (parts []string, err error) {
	parts = strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, errors.New("JWT is not composed by three segments (separated by dots)")
	}

	if err = ck.verifySignature(parts); err != nil {
		return nil, err
	}

	return parts, nil
}

func (ck *Checker) partsFromJWT(jwt string) (claimsJSON []byte, err error) {
	parts, err := ck.decomposeJWT(jwt)
	if err != nil {
		return nil, err
	}

	claimsJSON, err = ck.b64encoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("The token claims (second part of the JWT) is not base64-valid")
	}

	return claimsJSON, nil
}

// verifySignature of HS256 tokens.
func (ck *Checker) verifySignature(parts []string) (err error) {
	signingString := strings.Join(parts[0:2], ".")
	signedString := ck.sign(signingString)

	if signature := parts[2]; signature != signedString {
		return errors.New("JWT signature mismatch")
	}

	return nil
}

func (ck *Checker) permFromRefreshBytes(claimsJSON []byte) (perm Perm, err error) {
	claims := &tokens.RefreshClaims{
		Namespace: "",
		UserName:  "",
		StandardClaims: jwt.StandardClaims{
			Audience:  "",
			ExpiresAt: 0,
			Id:        ErrInvalidCookie.Error(),
			IssuedAt:  0,
			Issuer:    "",
			NotBefore: 0,
			Subject:   "",
		},
	}

	if err := json.Unmarshal(claimsJSON, claims); err != nil {
		return perm, fmt.Errorf("%w while unmarshaling RefreshClaims: "+
			security.Sanitize(string(claimsJSON)), err)
	}

	if err := claims.Valid(); err != nil {
		return perm, fmt.Errorf("%w in RefreshClaims: "+
			security.Sanitize(string(claimsJSON)), err)
	}

	perm = ck.permFromRefreshClaims(claims)
	return perm, nil
}

// sign allocates the hasher each time to avoid race condition.
func (ck *Checker) sign(signingString string) (signature string) {
	hasher := hmac.New(crypto.SHA256.New, ck.secretKey)
	_, _ = hasher.Write([]byte(signingString))
	return ck.b64encoding.EncodeToString(hasher.Sum(nil))
}

// --------------------------------------
// Read/write permissions to/from context

var permKey struct{}

// FromCtx gets the permission information from the request context.
func FromCtx(r *http.Request) Perm {
	perm, ok := r.Context().Value(permKey).(Perm)
	if !ok {
		log.Print("WRN JWT No permissions within the context ", r.URL.Path)
	}
	return perm
}

// putInCtx stores the permission info within the request context.
func (perm Perm) putInCtx(r *http.Request) *http.Request {
	parent := r.Context()
	child := context.WithValue(parent, permKey, perm)
	return r.WithContext(child)
}

// CC0-1.0: Creative Commons Zero v1.0 Universal
// No Rights Reserved - (CC) ZERO - (0) PUBLIC DOMAIN
//
// To the extent possible under law, the Teal.Finance contributors
// have waived all copyright and related or neighboring rights
// to this file "full-example_test.go" to be copied without restrictions.
// Refer to https://creativecommons.org/publicdomain/zero/1.0

package server_test

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/teal-finance/server"
	"github.com/teal-finance/server/chain"
	"github.com/teal-finance/server/cors"
	"github.com/teal-finance/server/limiter"
	"github.com/teal-finance/server/metrics"
	"github.com/teal-finance/server/opa"
	"github.com/teal-finance/server/reserr"
)

func Example() {
	// Uniformize error responses with API doc
	resErr := reserr.New("https://my-dns.com/doc")

	middlewares, connState := setMiddlewares(resErr)

	// Handles both REST API and static web files
	h := handler(resErr)
	h = middlewares.Then(h)

	runServer(h, connState)
}

func setMiddlewares(resErr reserr.ResErr) (middlewares chain.Chain, connState func(net.Conn, http.ConnState)) {
	// Start a metrics server in background if export port > 0.
	// The metrics server is for use with Prometheus or another compatible monitoring tool.
	metrics := metrics.Metrics{}
	middlewares, connState = metrics.StartServer(9093, true)

	// Limit the input request rate per IP
	reqLimiter := limiter.New(10, 20, true, resErr)
	middlewares = middlewares.Append()

	// Endpoint authentication rules (Open Policy Agent)
	policy, err := opa.New(resErr, []string{"rego.json"})
	if err != nil {
		log.Fatal(err)
	}

	// CORS
	allowedOrigins := []string{"http://my-dns.com"}

	middlewares = middlewares.Append(
		server.LogRequests,
		reqLimiter.Limit,
		server.Header("MyServerName-1.2.3"),
		policy.Auth,
		cors.HandleCORS(allowedOrigins),
	)

	return middlewares, connState
}

// runServer runs in foreground the main server.
func runServer(h http.Handler, connState func(net.Conn, http.ConnState)) {
	server := http.Server{
		Addr:              ":8080",
		Handler:           h,
		TLSConfig:         nil,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       1 * time.Second,
		MaxHeaderBytes:    222,
		TLSNextProto:      nil,
		ConnState:         connState,
		ErrorLog:          log.Default(),
		BaseContext:       nil,
		ConnContext:       nil,
	}

	log.Print("Server listening on http://localhost", server.Addr)

	log.Fatal(server.ListenAndServe())
}

package sdk

import (
	"net/http"

	"sumeru/core/server/router"
)

// RegisterMiddleware adds HTTP middleware applied after security enrichment.
func RegisterMiddleware(fn func(http.Handler) http.Handler) {
	router.RegisterMiddleware(fn)
}

// RegisterRoute adds an HTTP route (session, API key, or public auth).
func RegisterRoute(method, path string, auth router.AuthMode, handler http.HandlerFunc) {
	router.Register(method, path, auth, handler)
}

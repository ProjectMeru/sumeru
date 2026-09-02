package router

import (
	"net/http"
	"sync"
)

var (
	middlewareMu sync.RWMutex
	middlewares  []func(http.Handler) http.Handler
)

// RegisterMiddleware adds an HTTP middleware wrapper applied after security enrichment.
// Later registrations wrap outside earlier ones (first registered = innermost).
func RegisterMiddleware(fn func(http.Handler) http.Handler) {
	if fn == nil {
		return
	}
	middlewareMu.Lock()
	defer middlewareMu.Unlock()
	middlewares = append(middlewares, fn)
}

// ClearMiddleware removes registered middleware (tests).
func ClearMiddleware() {
	middlewareMu.Lock()
	defer middlewareMu.Unlock()
	middlewares = nil
}

// ApplyMiddleware wraps handler with all registered middleware (outermost last).
func ApplyMiddleware(handler http.Handler) http.Handler {
	middlewareMu.RLock()
	list := append([]func(http.Handler) http.Handler(nil), middlewares...)
	middlewareMu.RUnlock()
	for i := len(list) - 1; i >= 0; i-- {
		handler = list[i](handler)
	}
	return handler
}

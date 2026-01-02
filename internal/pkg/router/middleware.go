package router

import "net/http"

// Middleware wraps an http.Handler, typically to add cross-cutting behavior.
type Middleware func(http.Handler) http.Handler

// Chain applies middleware in order, returning the final wrapped handler.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Package router wraps HTTP routing and common middleware used by the API.
//
// It provides a small router abstraction over httprouter plus shared concerns
// like JSON encoding, error mapping, logging, recovery, authentication, and
// correlation ID propagation.
package router

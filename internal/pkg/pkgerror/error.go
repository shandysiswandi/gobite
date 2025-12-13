// Package pkgerror defines shared sentinel errors used across the application.
// These errors are intended to be compared with errors.Is and mapped by
// handlers to consistent HTTP or RPC responses.
package pkgerror

import "errors"

var (
	// ErrNotFound indicates that the requested resource could not be found.
	ErrNotFound = errors.New("resource not found")

	// ErrMethodNotAllowed indicates that the requested operation or HTTP method
	// is not allowed for the target resource.
	ErrMethodNotAllowed = errors.New("method not allowed")

	// ErrUnauthenticated indicates that authentication is required or that the
	// provided credentials are missing or invalid.
	ErrUnauthenticated = errors.New("authentication required")

	// ErrUnauthorized indicates that the authenticated user does not have
	// sufficient permission to perform the requested action.
	ErrUnauthorized = errors.New("permission denied")
)

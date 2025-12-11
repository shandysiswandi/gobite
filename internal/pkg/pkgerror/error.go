package pkgerror

import "errors"

var (
	// ErrNotFound indicates that the requested resource does not exist.
	ErrNotFound = errors.New("resource not found")

	// ErrUnauthenticated indicates that authentication is required
	// or the provided credentials are missing or invalid.
	ErrUnauthenticated = errors.New("authentication required")

	// ErrUnauthorized indicates that the authenticated user
	// does not have permission to perform the requested action.
	ErrUnauthorized = errors.New("permission denied")
)

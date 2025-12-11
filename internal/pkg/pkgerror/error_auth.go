package pkgerror

import "errors"

var (
	// ErrAuthNotVerified indicates that the user's account exists
	// but has not yet completed the verification process.
	ErrAuthNotVerified = errors.New("user account is not verified")

	// ErrAuthBanned indicates that the user's account is banned
	// and access is permanently or temporarily restricted.
	ErrAuthBanned = errors.New("user account is banned")

	// ErrAuthUnauthenticated indicates that authentication is required
	// or the provided credentials are missing or invalid.
	ErrAuthUnauthenticated = errors.New("invalid email or password")
)

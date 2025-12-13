package pkgerror

import "errors"

// Authentication-specific sentinel errors.

var (
	// ErrAuthNotVerified indicates that the user account exists but has not
	// completed the required verification process (for example, email verification).
	ErrAuthNotVerified = errors.New("user account is not verified")

	// ErrAuthBanned indicates that the user account is banned and access is
	// temporarily or permanently restricted.
	ErrAuthBanned = errors.New("user account is banned")

	// ErrAuthUnauthenticated indicates that authentication failed due to missing
	// or invalid credentials.
	//
	// Note: The error message is intentionally generic to avoid leaking
	// information about which credential was incorrect.
	ErrAuthUnauthenticated = errors.New("invalid email or password")

	// ErrAuthEmailUsed indicates that the provided email address is already
	// associated with an existing user account.
	ErrAuthEmailUsed = errors.New("email is already registered")
)

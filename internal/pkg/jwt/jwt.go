package jwt

import (
	"errors"
)

var (
	// ErrInvalidSigningMethod is returned when the JWT signing method is not supported.
	ErrInvalidSigningMethod = errors.New("invalid JWT signing method")

	// ErrSigningKeyTooShort is returned when the HS512 signing key is less than 64 bytes.
	ErrSigningKeyTooShort = errors.New("HS512 signing key must be at least 64 bytes (512 bits)")

	// ErrTokenExpired is returned when the JWT token has expired.
	ErrTokenExpired = errors.New("JWT token has expired")

	// ErrInvalidToken is returned when the token is malformed or fails validation.
	ErrInvalidToken = errors.New("invalid token")
)

// JWT defines the minimal operations needed by the app: generate and verify a token.
type JWT interface {
	Generate(...Option) (token string, err error)
	Verify(token string, opts ...Option) (Claims, error)
}

package pkgjwt

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

	// ErrTokenExpired is returned when the JWT token has expired.
	ErrInvalidToken = errors.New("invalid token")
)

type JWT[T any] interface {
	Generate(subject string, payload T) (token string, jti string, err error)
	Verify(token string) (Claims[T], error)
}

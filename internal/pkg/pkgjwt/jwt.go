package pkgjwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
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
	Verify(token string) (subject string, payload T, err error)
}

// Claims is a helper for wrapping registered claims with a typed payload.
type Claims[T any] struct {
	jwt.RegisteredClaims
	Payload T `json:"payload,omitempty"`
}

type Config struct {
	Secret   []byte
	Issuer   string
	Audience string
	TTL      time.Duration
	// ---
	Clock pkgclock.Clocker
	UUID  pkguid.StringID
}

// Access token payload
type AccessTokenPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// Refresh token payload
type RefreshTokenPayload struct {
	Message string `json:"message"`
}

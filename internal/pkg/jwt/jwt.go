package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	// Generate creates a signed token for the user.
	Generate(uid int64, email string) (string, error)
	// Verify parses and validates the token and returns claims.
	Verify(tokenStr string) (Claims, error)
}

type clocker interface {
	Now() time.Time
}

type generator interface {
	Generate() string
}

type jwtContextKey struct{}

// Config defines the inputs for building a JWT implementation.
type Config struct {
	// Secret is the HMAC signing key.
	Secret []byte
	// Issuer is the token issuer value.
	Issuer string
	// Audiences are the accepted token audiences.
	Audiences []string
	// TTLMinutes is the token time-to-live.
	TTLMinutes time.Duration
	// Clock provides the current time source.
	Clock clocker
	// UUID generates token IDs.
	UUID generator
}

// Claims is a helper for wrapping registered claims with a payload.
type Claims struct {
	// RegisteredClaims holds the standard JWT claims.
	jwt.RegisteredClaims
	// UserID is the authenticated user identifier.
	UserID int64 `json:"user_id,string"`
	// UserEmail is the authenticated user email.
	UserEmail string `json:"user_email"`
}

// GetAuth returns the JWT claims stored in the context, if any.
func GetAuth(ctx context.Context) *Claims {
	clm, ok := ctx.Value(jwtContextKey{}).(Claims)
	if !ok {
		return nil
	}

	return &clm
}

// SetAuth stores JWT claims in the context.
func SetAuth(ctx context.Context, clm Claims) context.Context {
	return context.WithValue(ctx, jwtContextKey{}, clm)
}

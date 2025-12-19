package jwt

import (
	"time"

	libJWT "github.com/golang-jwt/jwt/v5"
)

type jwtOptions struct {
	registeredClaims libJWT.RegisteredClaims
	payload          map[string]any
	secret           []byte
}

// Option configures token generation.
type Option func(*jwtOptions)

// WithSecret overrides the signing secret for this call.
func WithSecret(secret []byte) Option {
	return func(o *jwtOptions) {
		o.secret = secret
	}
}

// WithSubject sets the subject ("sub") claim.
func WithSubject(subject string) Option {
	return func(o *jwtOptions) {
		o.registeredClaims.Subject = subject
	}
}

// WithAudience sets the audience ("aud") claim.
func WithAudience(audience ...string) Option {
	return func(o *jwtOptions) {
		if len(audience) == 0 {
			return
		}
		o.registeredClaims.Audience = append([]string(nil), audience...)
	}
}

// WithID sets the JWT ID ("jti") claim.
func WithID(id string) Option {
	return func(o *jwtOptions) {
		o.registeredClaims.ID = id
	}
}

// WithExpiresAt sets the expiration ("exp") claim.
func WithExpiresAt(expires time.Duration) Option {
	return func(o *jwtOptions) {
		o.registeredClaims.ExpiresAt = libJWT.NewNumericDate(time.Now().Add(expires))
	}
}

// WithPayloadValue sets a single payload key.
func WithPayloadValue(key string, value any) Option {
	return func(o *jwtOptions) {
		if key == "" {
			return
		}
		if o.payload == nil {
			o.payload = make(map[string]any, 1)
		}
		o.payload[key] = value
	}
}

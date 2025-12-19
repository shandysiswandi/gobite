package jwt

import (
	"time"
)

type clocker interface {
	Now() time.Time
}

// Config defines the inputs for building a JWT implementation.
type Config struct {
	Secret    []byte
	Issuer    string
	Audiences []string
	TTL       time.Duration
	Clock     clocker
}

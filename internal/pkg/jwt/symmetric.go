package jwt

import (
	"errors"
	"time"

	libJWT "github.com/golang-jwt/jwt/v5"
)

// Symmetric implements JWT signing and verification using an HMAC secret.
type Symmetric struct {
	secret    []byte
	issuer    string
	audiences []string
	ttl       time.Duration
	// ---
	clock clocker
}

// NewHS512 constructs a Symmetric JWT implementation using HS512.
func NewHS512(cfg Config) (*Symmetric, error) {
	if len(cfg.Secret) < 64 {
		return nil, ErrSigningKeyTooShort
	}

	return &Symmetric{
		secret:    cfg.Secret,
		issuer:    cfg.Issuer,
		audiences: cfg.Audiences,
		ttl:       cfg.TTL,
		clock:     cfg.Clock,
	}, nil
}

func (s *Symmetric) Generate(opts ...Option) (string, error) {
	now := s.clock.Now()

	jwtOpt := &jwtOptions{
		registeredClaims: libJWT.RegisteredClaims{
			Issuer:    s.issuer,
			Audience:  s.audiences,
			IssuedAt:  libJWT.NewNumericDate(now),
			NotBefore: libJWT.NewNumericDate(now),
			ExpiresAt: libJWT.NewNumericDate(now.Add(s.ttl)),
		},
		payload: map[string]any{},
		secret:  s.secret,
	}

	for _, opt := range opts {
		opt(jwtOpt)
	}

	if len(jwtOpt.secret) < 64 {
		return "", ErrSigningKeyTooShort
	}

	return libJWT.
		NewWithClaims(libJWT.SigningMethodHS512, Claims{
			registeredClaims: jwtOpt.registeredClaims,
			payload:          jwtOpt.payload,
		}).
		SignedString(jwtOpt.secret)
}

func (s *Symmetric) Verify(tokenStr string, opts ...Option) (Claims, error) {
	var claims Claims

	jwtOpt := &jwtOptions{
		registeredClaims: libJWT.RegisteredClaims{
			Issuer:   s.issuer,
			Audience: s.audiences,
		},
		secret: s.secret,
	}

	for _, opt := range opts {
		opt(jwtOpt)
	}

	if len(jwtOpt.secret) < 64 {
		return Claims{}, ErrSigningKeyTooShort
	}

	token, err := libJWT.ParseWithClaims(tokenStr, &claims,
		func(t *libJWT.Token) (any, error) {
			if t.Method != libJWT.SigningMethodHS512 {
				return nil, ErrInvalidSigningMethod
			}
			return jwtOpt.secret, nil
		},
		libJWT.WithIssuer(jwtOpt.registeredClaims.Issuer),
		libJWT.WithAudience(jwtOpt.registeredClaims.Audience...),
		libJWT.WithValidMethods([]string{libJWT.SigningMethodHS512.Alg()}),
		libJWT.WithIssuedAt(),
		libJWT.WithExpirationRequired(),
	)

	if err != nil {
		if errors.Is(err, libJWT.ErrTokenExpired) {
			return Claims{}, ErrTokenExpired
		}
		return Claims{}, err
	}

	if !token.Valid {
		return Claims{}, ErrInvalidToken
	}

	return claims, nil
}

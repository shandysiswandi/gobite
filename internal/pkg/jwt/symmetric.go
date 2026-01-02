package jwt

import (
	"errors"
	"strconv"
	"time"

	libJWT "github.com/golang-jwt/jwt/v5"
)

// Symmetric implements JWT signing and verification using an HMAC secret.
type Symmetric struct {
	secret    []byte
	issuer    string
	audiences []string
	ttl       time.Duration
	clock     clocker
	uuid      generator
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
		ttl:       cfg.TTLMinutes,
		clock:     cfg.Clock,
		uuid:      cfg.UUID,
	}, nil
}

// Generate creates a signed JWT for the user.
func (s *Symmetric) Generate(uid int64, email string) (string, error) {
	now := s.clock.Now()

	if len(s.secret) < 64 {
		return "", ErrSigningKeyTooShort
	}

	return libJWT.
		NewWithClaims(libJWT.SigningMethodHS512, Claims{
			RegisteredClaims: libJWT.RegisteredClaims{
				ID:        s.uuid.Generate(),
				Subject:   strconv.FormatInt(uid, 10),
				Issuer:    s.issuer,
				Audience:  s.audiences,
				IssuedAt:  libJWT.NewNumericDate(now),
				NotBefore: libJWT.NewNumericDate(now),
				ExpiresAt: libJWT.NewNumericDate(now.Add(s.ttl)),
			},
			UserID:    uid,
			UserEmail: email,
		}).
		SignedString(s.secret)
}

// Verify parses and validates a JWT string.
func (s *Symmetric) Verify(tokenStr string) (Claims, error) {
	var claims Claims

	if len(s.secret) < 64 {
		return Claims{}, ErrSigningKeyTooShort
	}

	token, err := libJWT.ParseWithClaims(tokenStr, &claims,
		func(t *libJWT.Token) (any, error) {
			if t.Method != libJWT.SigningMethodHS512 {
				return nil, ErrInvalidSigningMethod
			}
			return s.secret, nil
		},
		libJWT.WithIssuer(s.issuer),
		libJWT.WithAudience(s.audiences...),
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

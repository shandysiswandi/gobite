package pkgjwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
)

type Symmetric[T any] struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
	// ---
	clock pkgclock.Clocker
	uuid  pkguid.StringID
}

func NewHS512[T any](cfg Config) (*Symmetric[T], error) {
	if len(cfg.Secret) < 64 {
		return nil, ErrSigningKeyTooShort
	}

	return &Symmetric[T]{
		secret:   cfg.Secret,
		issuer:   cfg.Issuer,
		audience: cfg.Audience,
		ttl:      cfg.TTL,
		clock:    cfg.Clock,
		uuid:     cfg.UUID,
	}, nil
}

func (s *Symmetric[T]) Generate(subject string, payload T) (string, string, error) {
	now := s.clock.Now()
	jti := s.uuid.Generate()

	claims := Claims[T]{}
	claims.registeredClaims = jwt.RegisteredClaims{
		Issuer:    s.issuer,
		Audience:  []string{s.audience},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		Subject:   subject,
		ID:        jti,
	}
	claims.payload = payload

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signed, err := token.SignedString(s.secret)
	return signed, jti, err
}

func (s *Symmetric[T]) Verify(tokenStr string) (Claims[T], error) {
	var claims Claims[T]
	var zero Claims[T]

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&claims,
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS512 {
				return nil, ErrInvalidSigningMethod
			}
			return s.secret, nil
		},
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Alg()}),
		jwt.WithIssuedAt(),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return zero, ErrTokenExpired
		}
		return zero, err
	}

	if !token.Valid {
		return zero, ErrInvalidToken
	}

	return claims, nil
}

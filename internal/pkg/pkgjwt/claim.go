package pkgjwt

import (
	"encoding/json"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is a helper for wrapping registered claims with a typed payload.
type Claims[T any] struct {
	registeredClaims jwt.RegisteredClaims
	payload          T
}

func (c Claims[T]) ID() string {
	return c.registeredClaims.ID
}

func (c Claims[T]) Subject() string {
	return c.registeredClaims.Subject
}

func (c Claims[T]) Payload() T {
	return c.payload
}

func (c Claims[T]) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.registeredClaims.GetExpirationTime()
}

func (c Claims[T]) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.registeredClaims.GetIssuedAt()
}

func (c Claims[T]) GetNotBefore() (*jwt.NumericDate, error) {
	return c.registeredClaims.GetNotBefore()
}

func (c Claims[T]) GetIssuer() (string, error) {
	return c.registeredClaims.GetIssuer()
}

func (c Claims[T]) GetSubject() (string, error) {
	return c.registeredClaims.GetSubject()
}

func (c Claims[T]) GetAudience() (jwt.ClaimStrings, error) {
	return c.registeredClaims.GetAudience()
}

// MarshalJSON ensures private fields are encoded with proper keys.
func (c Claims[T]) MarshalJSON() ([]byte, error) {
	type alias struct {
		jwt.RegisteredClaims
		Payload T `json:"payload,omitempty"`
	}

	return json.Marshal(alias{
		RegisteredClaims: c.registeredClaims,
		Payload:          c.payload,
	})
}

// UnmarshalJSON ensures private fields are decoded correctly.
func (c *Claims[T]) UnmarshalJSON(data []byte) error {
	type alias struct {
		jwt.RegisteredClaims
		Payload T `json:"payload,omitempty"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	c.registeredClaims = a.RegisteredClaims
	c.payload = a.Payload
	return nil
}

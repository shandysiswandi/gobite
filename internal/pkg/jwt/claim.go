package jwt

import (
	"encoding/json"

	libJWT "github.com/golang-jwt/jwt/v5"
)

// Claims is a helper for wrapping registered claims with a payload.
type Claims struct {
	registeredClaims libJWT.RegisteredClaims
	payload          map[string]any
}

func (c Claims) ID() string {
	return c.registeredClaims.ID
}

func (c Claims) Subject() string {
	return c.registeredClaims.Subject
}

func (c Claims) Payload() map[string]any {
	return c.payload
}

func (c Claims) PayloadValue(key string) (any, bool) {
	if c.payload == nil {
		return nil, false
	}

	value, ok := c.payload[key]
	return value, ok
}

func (c Claims) GetString(key string) string {
	value, ok := c.PayloadValue(key)
	if !ok {
		return ""
	}

	if v, ok := value.(string); ok {
		return v
	}

	return ""
}

func (c Claims) GetInt64(key string) int64 {
	value, ok := c.PayloadValue(key)
	if !ok {
		return 0
	}

	switch v := value.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case json.Number:
		if n, err := v.Int64(); err == nil {
			return n
		}
	}

	return 0
}

func (c Claims) GetBool(key string) bool {
	value, ok := c.PayloadValue(key)
	if !ok {
		return false
	}

	if v, ok := value.(bool); ok {
		return v
	}

	return false
}

func (c Claims) GetExpirationTime() (*libJWT.NumericDate, error) {
	return c.registeredClaims.GetExpirationTime()
}

func (c Claims) GetIssuedAt() (*libJWT.NumericDate, error) {
	return c.registeredClaims.GetIssuedAt()
}

func (c Claims) GetNotBefore() (*libJWT.NumericDate, error) {
	return c.registeredClaims.GetNotBefore()
}

func (c Claims) GetIssuer() (string, error) {
	return c.registeredClaims.GetIssuer()
}

func (c Claims) GetSubject() (string, error) {
	return c.registeredClaims.GetSubject()
}

func (c Claims) GetAudience() (libJWT.ClaimStrings, error) {
	return c.registeredClaims.GetAudience()
}

// MarshalJSON ensures private fields are encoded with proper keys.
func (c Claims) MarshalJSON() ([]byte, error) {
	type alias struct {
		libJWT.RegisteredClaims
		Payload map[string]any `json:"payload,omitempty"`
	}

	return json.Marshal(alias{
		RegisteredClaims: c.registeredClaims,
		Payload:          c.payload,
	})
}

// UnmarshalJSON ensures private fields are decoded correctly.
func (c *Claims) UnmarshalJSON(data []byte) error {
	type alias struct {
		libJWT.RegisteredClaims
		Payload map[string]any `json:"payload,omitempty"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	c.registeredClaims = a.RegisteredClaims
	c.payload = a.Payload
	return nil
}

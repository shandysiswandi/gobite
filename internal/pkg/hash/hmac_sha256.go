package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// HMACSHA256 implements the Hash interface using SHA-256.
type HMACSHA256 struct {
	secret []byte
}

// NewHMACSHA256 creates a new hasher with a secret.
func NewHMACSHA256(secret string) *HMACSHA256 {
	return &HMACSHA256{secret: []byte(secret)}
}

// Hash returns the HMAC SHA-256 hash of the input string (hex-encoded).
func (s *HMACSHA256) Hash(str string) ([]byte, error) {
	return s.gen(str), nil
}

// Verify checks whether the plaintext string matches the given hash.
func (s *HMACSHA256) Verify(hashed, str string) bool {
	expected := s.gen(str)
	return subtle.ConstantTimeCompare([]byte(hashed), expected) == 1
}

func (s *HMACSHA256) gen(str string) []byte {
	h := hmac.New(sha256.New, s.secret)
	h.Write([]byte(str))
	sum := h.Sum(nil)
	result := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(result, sum)
	return result
}

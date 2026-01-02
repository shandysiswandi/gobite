package hash

import (
	"golang.org/x/crypto/bcrypt"
)

// Bcrypt implements Hash using bcrypt.
//
// Pepper is appended to the plaintext before hashing/verifying. Keep the pepper
// secret and store it in configuration (not in the database).
type Bcrypt struct {
	cost   int
	pepper string
}

// NewBcrypt returns a bcrypt-based hasher.
//
// cost controls the hashing work factor (see bcrypt.DefaultCost). pepper is
// optional but recommended as an extra secret.
func NewBcrypt(cost int, pepper string) *Bcrypt {
	return &Bcrypt{cost: cost, pepper: pepper}
}

// Hash hashes plaintext using bcrypt.
func (h *Bcrypt) Hash(plaintext string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(plaintext+h.pepper), h.cost)
}

// Verify returns true when plaintext matches the hashed value.
func (h *Bcrypt) Verify(hashed, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext+h.pepper)) == nil
}

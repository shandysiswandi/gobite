package pkghash

import (
	"golang.org/x/crypto/bcrypt"
)

type Bcrypt struct {
	cost   int
	pepper string
}

func NewBcrypt(cost int, pepper string) *Bcrypt {
	return &Bcrypt{cost: cost, pepper: pepper}
}

func (h *Bcrypt) Hash(plaintext string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(plaintext+h.pepper), h.cost)
}

func (h *Bcrypt) Verify(hashed, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext+h.pepper)) == nil
}

package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id implements the Hash interface using Argon2id.
type Argon2id struct {
	memory        uint32
	iterations    uint32
	parallelism   uint8
	saltLength    uint32
	keyLength     uint32
	maxConcurrent int
	pepper        string
}

// NewArgon2id returns a Argon2id hasher with recommended defaults.
func NewArgon2id(pepper string) *Argon2id {
	return &Argon2id{
		memory:        32 * 1024, // e.g. 32MB, 64MB, 128MB
		iterations:    3,         // time cost
		parallelism:   2,         // threads
		saltLength:    16,
		keyLength:     32,
		maxConcurrent: 2, // semaphore size; 0 disables limiter
		pepper:        pepper,
	}
}

// Hash takes a plaintext string and returns its hashed representation.
func (a *Argon2id) Hash(str string) ([]byte, error) {
	salt := make([]byte, a.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(str+a.pepper), salt, a.iterations, a.memory, a.parallelism, a.keyLength)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		a.memory,
		a.iterations,
		a.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return []byte(encoded), nil
}

// Verify checks if the given plaintext string matches the hashed value.
func (a *Argon2id) Verify(hashed, str string) bool {
	if len(hashed) == 0 || str == "" {
		return false
	}

	// Parse encoded hash
	parts := strings.Split(hashed, "$")
	if len(parts) != 6 {
		return false
	}

	if parts[1] != "argon2id" {
		return false
	}

	// Parse version
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false
	}

	// Parse parameters
	var memory uint32
	var iterations uint32
	var parallelism uint8

	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false
	}

	// Decode salt + expected hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	// Compute hash with extracted params
	computedHash := argon2.IDKey([]byte(str+a.pepper), salt, iterations, memory, parallelism, uint32(len(expectedHash)))

	// Constant-time compare
	return subtle.ConstantTimeCompare(expectedHash, computedHash) == 1
}

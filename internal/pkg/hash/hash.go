package hash

// Hash defines methods for hashing and verifying strings.
type Hash interface {
	// Hash takes a plaintext string and returns its hashed representation.
	// It may return an error if the hashing process fails.
	Hash(str string) ([]byte, error)

	// Verify checks if the given plaintext string matches the hashed value.
	// It returns true if the plaintext matches the hash, otherwise false.
	Verify(hashed, str string) bool
}

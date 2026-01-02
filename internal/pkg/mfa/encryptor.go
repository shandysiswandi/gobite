package mfa

// Encryptor defines the interface for encrypting/decrypting.
type Encryptor interface {
	// Encrypt returns ciphertext for the given plaintext and scope.
	Encrypt(plaintext []byte, scope Scope) (ciphertext []byte, err error)
	// Decrypt returns plaintext for the given ciphertext and scope.
	Decrypt(ciphertext []byte, scope Scope) (plaintext []byte, err error)
}

// KeyProvider provides raw AES keys.
// For AES-256-GCM, keys must be 32 bytes.
type KeyProvider interface {
	// Key returns the raw AES key to use for this scope.
	// You may choose to return per-tenant keys, per-environment keys, etc.
	Key(scope Scope) ([]byte, error)
}

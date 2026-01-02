package mfa

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// AESGCMEncryptor implements Encryptor using AES-GCM.
type AESGCMEncryptor struct {
	keys KeyProvider
}

// NewAESGCMEncryptor constructs an AES-GCM encryptor.
func NewAESGCMEncryptor(keys KeyProvider) *AESGCMEncryptor {
	return &AESGCMEncryptor{keys: keys}
}

// Ciphertext format (binary):
// [0..1]   uint16 version (currently 1)
// [2..13]  12-byte nonce
// [14..]   gcm.Seal output (ciphertext + tag)
const aesGCMVersion uint16 = 1

const (
	gcmNonceSize = 12
	aesKeyLen    = 32
)

var (
	// ErrEncryptorNotConfigured indicates a missing encryptor key provider.
	ErrEncryptorNotConfigured = errors.New("mfacrypto: encryptor not configured")
	// ErrPlaintextEmpty indicates an empty plaintext input.
	ErrPlaintextEmpty = errors.New("mfacrypto: plaintext is empty")
	// ErrInvalidKeyLength indicates the key length is invalid.
	ErrInvalidKeyLength = errors.New("mfacrypto: invalid key length")
	// ErrUnexpectedNonceSize indicates a nonce size mismatch.
	ErrUnexpectedNonceSize = errors.New("mfacrypto: unexpected nonce size")
	// ErrCiphertextTooShort indicates a truncated ciphertext.
	ErrCiphertextTooShort = errors.New("mfacrypto: ciphertext too short")
	// ErrUnsupportedCiphertextVersion indicates an unsupported ciphertext version.
	ErrUnsupportedCiphertextVersion = errors.New("mfacrypto: unsupported ciphertext version")
	// ErrDecryptFailed indicates decryption failure.
	ErrDecryptFailed = errors.New("mfacrypto: decrypt failed")
	// ErrMissingStaticKey indicates a missing static key.
	ErrMissingStaticKey = errors.New("mfacrypto: missing static key")
)

// Encrypt encrypts plaintext with AES-256-GCM, binding the result to scope via AAD.
func (e *AESGCMEncryptor) Encrypt(plaintext []byte, scope Scope) ([]byte, error) {
	if e == nil || e.keys == nil {
		return nil, ErrEncryptorNotConfigured
	}
	if len(plaintext) == 0 {
		return nil, ErrPlaintextEmpty
	}

	key, err := e.keys.Key(scope)
	if err != nil {
		return nil, fmt.Errorf("mfacrypto: key provider error: %w", err)
	}
	if len(key) != aesKeyLen {
		return nil, fmt.Errorf("mfacrypto: invalid key length %d (want %d for AES-256): %w", len(key), aesKeyLen, ErrInvalidKeyLength)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("mfacrypto: aes init failed: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("mfacrypto: gcm init failed: %w", err)
	}
	if gcm.NonceSize() != gcmNonceSize {
		return nil, fmt.Errorf("mfacrypto: unexpected nonce size %d (want %d): %w", gcm.NonceSize(), gcmNonceSize, ErrUnexpectedNonceSize)
	}

	nonce := make([]byte, gcmNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("mfacrypto: nonce generation failed: %w", err)
	}

	aad := scopeAAD(scope)

	// Seal appends ciphertext+tag to the first arg; we pass nil to allocate a fresh slice.
	sealed := gcm.Seal(nil, nonce, plaintext, aad)

	out := make([]byte, 2+gcmNonceSize+len(sealed))
	binary.BigEndian.PutUint16(out[0:2], aesGCMVersion)
	copy(out[2:2+gcmNonceSize], nonce)
	copy(out[2+gcmNonceSize:], sealed)

	return out, nil
}

// Decrypt decrypts ciphertext with AES-256-GCM, requiring the same scope AAD.
func (e *AESGCMEncryptor) Decrypt(ciphertext []byte, scope Scope) ([]byte, error) {
	if e == nil || e.keys == nil {
		return nil, ErrEncryptorNotConfigured
	}
	if len(ciphertext) < 2+gcmNonceSize+1 {
		return nil, ErrCiphertextTooShort
	}

	version := binary.BigEndian.Uint16(ciphertext[0:2])
	if version != aesGCMVersion {
		return nil, fmt.Errorf("mfacrypto: unsupported ciphertext version %d: %w", version, ErrUnsupportedCiphertextVersion)
	}

	nonce := ciphertext[2 : 2+gcmNonceSize]
	sealed := ciphertext[2+gcmNonceSize:]

	key, err := e.keys.Key(scope)
	if err != nil {
		return nil, fmt.Errorf("mfacrypto: key provider error: %w", err)
	}
	if len(key) != aesKeyLen {
		return nil, fmt.Errorf("mfacrypto: invalid key length %d (want %d for AES-256): %w", len(key), aesKeyLen, ErrInvalidKeyLength)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("mfacrypto: aes init failed: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("mfacrypto: gcm init failed: %w", err)
	}
	if gcm.NonceSize() != gcmNonceSize {
		return nil, fmt.Errorf("mfacrypto: unexpected nonce size %d (want %d): %w", gcm.NonceSize(), gcmNonceSize, ErrUnexpectedNonceSize)
	}

	aad := scopeAAD(scope)

	plain, err := gcm.Open(nil, nonce, sealed, aad)
	if err != nil {
		// Important: do not leak whether it was "wrong scope" vs "wrong key" vs "tampered".
		return nil, ErrDecryptFailed
	}
	return plain, nil
}

// scopeAAD encodes the scope into a stable, tamper-evident byte slice for GCM AAD.
//
// We hash a canonical string to:
// - keep AAD length fixed
// - avoid separator ambiguity
// - avoid leaking raw IDs in logs if AAD is ever logged (still: don't log it)
func scopeAAD(s Scope) []byte {
	// Canonical form; include field labels to avoid accidental collisions.
	// Purpose is included to prevent cross-purpose decrypt (OTP seed vs recovery key).
	canonical := fmt.Sprintf("uid=%d\npurpose=%s\n", s.UserID, s.Purpose)
	sum := sha256.Sum256([]byte(canonical))
	return sum[:]
}

// StaticKeyProvider returns the same key for every scope.
// Good for local dev only. In production, prefer a KMS-backed provider and key rotation.
type StaticKeyProvider struct {
	// KeyBytes is the raw AES key material.
	KeyBytes []byte
}

// Key returns the static key for the provided scope.
func (p StaticKeyProvider) Key(_ Scope) ([]byte, error) {
	if len(p.KeyBytes) == 0 {
		return nil, ErrMissingStaticKey
	}
	// Defensive copy so callers can't mutate the provider's key.
	k := make([]byte, len(p.KeyBytes))
	copy(k, p.KeyBytes)
	return k, nil
}

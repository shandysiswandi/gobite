// Package hash provides helpers for hashing and verifying secrets.
//
// Typical usage is for password hashing: store only the hash, then verify user
// input by comparing the plaintext against the stored hash. Implementations
// (like bcrypt) live in this package behind a small interface.
package hash

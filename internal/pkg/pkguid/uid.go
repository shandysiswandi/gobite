// Package pkguid provides interfaces and implementations for generating unique identifiers (UIDs).
package pkguid

// StringID defines an interface for generating unique string-based UIDs.
type StringID interface {
	// Generate generates a unique identifier as a string.
	Generate() string
}

// NumberID defines an interface for generating unique numeric UIDs.
type NumberID interface {
	// Generate generates a unique identifier as a uint64 number.
	Generate() uint64
}

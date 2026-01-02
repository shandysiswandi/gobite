package uid

// StringID generates unique string identifiers.
type StringID interface {
	// Generate generates a unique identifier as a string.
	Generate() string
}

// NumberID generates unique numeric identifiers.
type NumberID interface {
	// Generate generates a unique identifier as a uint64 number.
	Generate() int64
}

package pkguid

import "github.com/google/uuid"

// UUID generates RFC 4122 UUID strings.
type UUID struct{}

// NewUUID returns a UUID generator.
func NewUUID() *UUID {
	return &UUID{}
}

// Generate returns a new UUID string.
func (u *UUID) Generate() string {
	return uuid.Must(uuid.NewV7()).String()
}

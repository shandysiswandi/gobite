package pkguid

import "github.com/google/uuid"

type UUID struct{}

func NewUUID() *UUID {
	return &UUID{}
}

func (u *UUID) Generate() string {
	return uuid.Must(uuid.NewV7()).String()
}

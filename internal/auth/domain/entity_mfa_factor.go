package domain

import (
	"time"
)

type MfaFactor struct {
	ID           int64
	UserID       int64
	Type         int16
	FriendlyName string
	Secret       []byte
	KeyVersion   int16 // key rotation version
	IsVerified   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

package domain

import (
	"time"
)

type MfaFactor struct {
	ID                   int64
	UserID               int64
	Type                 int16
	FriendlyName         string
	EncryptedSecret      []byte
	EncryptionKeyVersion int16 // key rotation version
	IsVerified           bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

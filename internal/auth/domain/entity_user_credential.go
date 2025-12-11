package domain

import (
	"time"
)

type UserCredential struct {
	UserID        int64
	Password      string // hashed
	LastChangedAt time.Time
}

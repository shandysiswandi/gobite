package entity

import (
	"time"
)

type UserCredential struct {
	UserID    int64
	Password  string // hashed
	UpdatedAt time.Time
}

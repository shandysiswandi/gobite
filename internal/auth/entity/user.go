package entity

import (
	"time"
)

type User struct {
	ID        int64
	Email     string
	FullName  string
	AvatarURL string
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

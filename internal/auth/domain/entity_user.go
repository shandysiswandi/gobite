package domain

import (
	"time"
)

type UserStatus int16

const (
	UserStatusUnverified UserStatus = iota
	UserStatusActive
	UserStatusBanned
)

func (us UserStatus) Value() int16 {
	return int16(us)
}

type User struct {
	ID        int64
	Email     string
	FullName  string
	AvatarURL string
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

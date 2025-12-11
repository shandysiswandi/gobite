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

type User struct {
	ID        int64
	Email     string
	FullName  string
	AvatarURL string
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

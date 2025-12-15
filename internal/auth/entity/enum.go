package entity

import "errors"

var (
	ErrUserStatusBanned     = errors.New("auth: user status is banned")
	ErrUserStatusUnverified = errors.New("auth: user status verify yet")
)

type UserStatus int16

const (
	UserStatusUnverified UserStatus = 0
	UserStatusActive     UserStatus = 1
	UserStatusBanned     UserStatus = 2
)

func (us UserStatus) String() string {
	switch us {
	case UserStatusActive:
		return "Active"
	case UserStatusBanned:
		return "Banned"
	case UserStatusUnverified:
		return "Unverified"
	default:
		return "Unknown"
	}
}

func (us UserStatus) Validate() error {
	switch us {
	case UserStatusActive:
		return nil
	case UserStatusBanned:
		return ErrUserStatusBanned
	case UserStatusUnverified:
		return ErrUserStatusUnverified
	default:
		return ErrUserStatusUnverified
	}
}

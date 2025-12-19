package entity

import "errors"

var (
	ErrUserStatusUnknown    = errors.New("auth: user status is unknown")
	ErrUserStatusBanned     = errors.New("auth: user status is banned")
	ErrUserStatusUnverified = errors.New("auth: user status verify yet")
)

type UserStatus int16

const (
	UserStatusUnknown    UserStatus = 0
	UserStatusUnverified UserStatus = 1
	UserStatusActive     UserStatus = 2
	UserStatusBanned     UserStatus = 3
	UserStatusDeleted    UserStatus = 4
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

func (us UserStatus) Ensure() UserStatus {
	switch us {
	case UserStatusActive:
		return UserStatusActive
	case UserStatusBanned:
		return UserStatusBanned
	case UserStatusUnverified:
		return UserStatusUnverified
	default:
		return UserStatusUnknown
	}
}

type ChallengePurpose int16

const (
	ChallengePurposeUnknown             ChallengePurpose = 0
	ChallengePurposeMFALogin            ChallengePurpose = 1
	ChallengePurposeMFASetupConfirm     ChallengePurpose = 2
	ChallengePurposePasswordForgotReset ChallengePurpose = 3
)

type MFAType int16

const (
	MFATypeUnknown MFAType = 0
	MFATypeTOTP    MFAType = 1
	MFATypeSMS     MFAType = 2
)

func (mt MFAType) String() string {
	switch mt {
	case MFATypeTOTP:
		return "TOTP"
	case MFATypeSMS:
		return "SMS"
	default:
		return "Unknown"
	}
}

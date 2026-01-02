package entity

import "errors"

var (
	ErrUserStatusUnknown    = errors.New("identity: user status is unknown")
	ErrUserStatusBanned     = errors.New("identity: user status is banned")
	ErrUserStatusUnverified = errors.New("identity: user status is unverified")
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
	case UserStatusDeleted:
		return "Deleted"
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
	case UserStatusDeleted:
		return UserStatusDeleted
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
	ChallengePurposeRegisterVerify      ChallengePurpose = 4
)

type MFAType int16

const (
	MFATypeUnknown    MFAType = 0
	MFATypeTOTP       MFAType = 1
	MFATypeSMS        MFAType = 2
	MFATypeBackupCode MFAType = 3
)

func MFATypeFrom(str string) MFAType {
	switch str {
	case "TOTP":
		return MFATypeTOTP
	case "SMS":
		return MFATypeSMS
	case "BackupCode":
		return MFATypeBackupCode
	default:
		return MFATypeUnknown
	}
}

func (mt MFAType) String() string {
	switch mt {
	case MFATypeTOTP:
		return "TOTP"
	case MFATypeSMS:
		return "SMS"
	case MFATypeBackupCode:
		return "BackupCode"
	default:
		return "Unknown"
	}
}

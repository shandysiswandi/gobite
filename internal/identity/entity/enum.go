package entity

import (
	"errors"
	"strconv"
)

var (
	ErrUserStatusUnknown    = errors.New("identity: user status is unknown")
	ErrUserStatusBanned     = errors.New("identity: user status is banned")
	ErrUserStatusUnverified = errors.New("identity: user status is unverified")
)

type UserStatus int16

const (
	// UserStatusUnknown is mean status is not known / not set.
	UserStatusUnknown UserStatus = 0

	// UserStatusUnverified mean user exists but has not completed verification.
	UserStatusUnverified UserStatus = 1

	// UserStatusActive mean user is verified and allowed to use the app.
	UserStatusActive UserStatus = 2

	// UserStatusBanned mean user is blocked from using the app (policy/abuse/etc).
	UserStatusBanned UserStatus = 3

	// UserStatusInactive mean user is not currently active (e.g., deactivated, closed).
	UserStatusInactive UserStatus = 4
)

func (us UserStatus) String() string {
	switch us {
	case UserStatusActive:
		return "Active"
	case UserStatusBanned:
		return "Banned"
	case UserStatusInactive:
		return "Inactive"
	case UserStatusUnverified:
		return "Unverified"
	default:
		return "Unknown"
	}
}

func (s UserStatus) IsUnknown() bool {
	switch s {
	case UserStatusUnverified, UserStatusActive, UserStatusBanned, UserStatusInactive:
		return false
	default:
		return true
	}
}

func (us UserStatus) Ensure() UserStatus {
	switch us {
	case UserStatusActive:
		return UserStatusActive
	case UserStatusBanned:
		return UserStatusBanned
	case UserStatusInactive:
		return UserStatusInactive
	case UserStatusUnverified:
		return UserStatusUnverified
	default:
		return UserStatusUnknown
	}
}

func ParseSafeUserStatuses(raws []string) []UserStatus {
	out := make([]UserStatus, 0)
	seen := map[UserStatus]struct{}{}

	for _, v := range raws {
		n, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			continue
		}

		s := UserStatus(n)
		if s.IsUnknown() {
			continue
		}

		if _, ok := seen[s]; ok {
			continue
		}

		seen[s] = struct{}{}
		out = append(out, s)
	}

	return out
}

func ToInt16Slice(sts []UserStatus) []int16 {
	out := make([]int16, len(sts))
	for i, s := range sts {
		out[i] = int16(s)
	}
	return out
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

func MFATypeFromString(str string) MFAType {
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

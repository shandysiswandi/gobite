package entity

import (
	"time"

	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

type User struct {
	ID        int64
	Email     string
	FullName  string
	AvatarURL string
	Status    UserStatus
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type MFAFactor struct {
	ID           int64
	UserID       int64
	Type         MFAType
	FriendlyName string
	Secret       []byte
	KeyVersion   int16 // key rotation version
	IsVerified   bool
}

type UserCredential struct {
	UserID    int64
	Password  string // hashed
	UpdatedAt time.Time
}

type Challenge struct {
	ID        int64
	UserID    int64
	Token     string
	Purpose   ChallengePurpose
	ExpiresAt time.Time
	Metadata  valueobject.JSONMap
}

type MFABackupCode struct {
	ID     int64
	UserID int64
	Code   string
}

type RefreshToken struct {
	ID                int64
	UserID            int64
	Token             string
	ExpiresAt         time.Time
	Revoked           bool
	ReplacedByTokenID int64
	Metadata          valueobject.JSONMap
}

// ---- //

type ChallengeUser struct {
	ChallengeID       int64
	ChallengePurpose  ChallengePurpose
	ChallengeToken    string
	ChallengeMetadata valueobject.JSONMap
	UserID            int64
	UserEmail         string
	UserStatus        UserStatus
}

type UserLoginInfo struct {
	ID       int64
	Email    string
	Status   UserStatus
	Password string
	HasMFA   bool
}

type UserCredentialInfo struct {
	ID       int64
	Email    string
	Status   UserStatus
	Password string
}

type RotateRefreshToken struct {
	NewID        int64
	OldID        int64
	UserID       int64
	NewToken     string
	NewExpiresAt time.Time
}

type UserRefreshToken struct {
	UserID                   int64
	UserEmail                string
	UserStatus               UserStatus
	RefreshID                int64
	RefreshToken             string
	RefreshRevoked           bool
	RefreshReplacedByTokenID *int64
	RefreshExpiresAt         time.Time
}

type VerifyUserRegistration struct {
	ChallengeID   int64
	UserID        int64
	UpdatedBy     int64
	OldUserStatus UserStatus
	NewUserStatus UserStatus
}

type UserListFilterData struct {
	IsFilterBySearch bool
	IsFilterByStatus bool
	Search           string
	Statuses         []int16
	DateFrom         time.Time
	DateTo           time.Time
	Size             int32
	Page             int32
	OrderBy          string
	OrderDirection   string
}

type NewUser struct {
	ID        int64
	Email     string
	FullName  string
	AvatarURL string
	Status    UserStatus
	CreatedBy int64
	UpdatedBy int64
}

type PatchUser struct {
	ID        int64
	Email     string
	FullName  string
	AvatarURL string
	Status    UserStatus
	UpdatedBy int64
}

type UpsertUser struct {
	ID        int64
	Email     string
	FullName  string
	AvatarURL string
	Status    UserStatus
	CreatedBy int64
	UpdatedBy int64
}

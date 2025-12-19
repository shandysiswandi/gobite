package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/shandysiswandi/gobite/internal/auth/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/clock"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/hash"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/mfacrypto"
	"github.com/shandysiswandi/gobite/internal/pkg/otp"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/pkg/validator"
)

const (
	keyPayloadUserID    = "user_id"
	keyPayloadUserEmail = "user_email"
)

type repoMessaging interface {
	PublishUserRegistration(ctx context.Context, msg entity.UserRegistrationMessage) error
	PublishUserForgotPassword(ctx context.Context, msg entity.UserForgotPasswordMessage) error
}

type repoCache interface {
	RegisterResendAllow(ctx context.Context, key string, ttl time.Duration) (allowed bool, err error)
}

type repoDB interface {
	GetUserLoginInfo(ctx context.Context, email string) (*entity.UserLoginInfo, error)
	GetUserCredentialInfo(ctx context.Context, id int64) (*entity.UserCredentialInfo, error)
	CreateRefreshToken(ctx context.Context, in entity.RefreshToken) error
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllRefreshToken(ctx context.Context, userID int64) error
	CreateChallenge(ctx context.Context, in entity.Challenge) error
	GetChallengeUserByTokenPurpose(ctx context.Context, token string, p entity.ChallengePurpose) (*entity.ChallengeUser, error)
	DeleteChallenge(ctx context.Context, id int64) error
	GetUserRefreshToken(ctx context.Context, token string) (*entity.UserRefreshToken, error)
	RotateRefreshToken(ctx context.Context, ro entity.RotateRefreshToken) error
	GetUserByEmail(ctx context.Context, email string, includeDeleted bool) (*entity.User, error)
	GetMFAFactorByUserID(ctx context.Context, userID int64, isVerified bool) ([]entity.MFAFactor, error)
	CreateMFAFactorAndChallenge(ctx context.Context, factor entity.MFAFactor, challe entity.Challenge) error
	UpdateUserProfile(ctx context.Context, id int64, fullName string) error
	UserRegistration(ctx context.Context, user entity.User, hash string) error
	UpdateUserCredential(ctx context.Context, userID int64, hash string) error
	// CreateUserPasswordReset(ctx context.Context, userID int64, token string, expiresAt time.Time) error
}

type Usecase struct {
	repoDB        repoDB
	repoCache     repoCache
	repoMessaging repoMessaging
	validator     validator.Validator
	cfg           config.Config
	password      hash.Hash
	hash          hash.Hash
	mfaCrypto     mfacrypto.Encryptor
	uid           uid.NumberID
	uuid          uid.StringID
	oid           uid.StringID
	totp          otp.OTP
	clock         clock.Clocker
	jwt           jwt.JWT
}

type Dependency struct {
	RepoDB        repoDB
	RepoCache     repoCache
	RepoMessaging repoMessaging
	Validator     validator.Validator
	Config        config.Config
	Password      hash.Hash
	Hash          hash.Hash
	MFACrypto     mfacrypto.Encryptor
	UID           uid.NumberID
	UUID          uid.StringID
	OID           uid.StringID
	Totp          otp.OTP
	Clock         clock.Clocker
	JWT           jwt.JWT
}

func NewAuth(dep Dependency) *Usecase {
	return &Usecase{
		repoDB:        dep.RepoDB,
		repoCache:     dep.RepoCache,
		repoMessaging: dep.RepoMessaging,
		validator:     dep.Validator,
		password:      dep.Password,
		hash:          dep.Hash,
		mfaCrypto:     dep.MFACrypto,
		cfg:           dep.Config,
		uid:           dep.UID,
		uuid:          dep.UUID,
		oid:           dep.OID,
		totp:          dep.Totp,
		clock:         dep.Clock,
		jwt:           dep.JWT,
	}
}

func (s *Usecase) ensureUserAllowedToLogin(ctx context.Context, userID int64, status entity.UserStatus) error {
	sts := status.Ensure()
	switch sts {
	case entity.UserStatusUnknown:
		slog.WarnContext(ctx, "user account status is unrecognized", "user_id", userID)
		return goerror.NewServer(entity.ErrUserStatusUnknown)

	case entity.UserStatusUnverified:
		slog.WarnContext(ctx, "user account is unverified", "user_id", userID)
		return goerror.NewBusiness("email not verified", goerror.CodeForbidden)

	case entity.UserStatusBanned:
		slog.WarnContext(ctx, "user account is banned", "user_id", userID)
		return goerror.NewBusiness("account is banned", goerror.CodeForbidden)

	default:
		return nil
	}
}

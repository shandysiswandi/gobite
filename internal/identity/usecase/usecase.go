package usecase

import (
	"context"
	"log/slog"

	"github.com/casbin/casbin/v3"
	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/clock"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/goroutine"
	"github.com/shandysiswandi/gobite/internal/pkg/hash"
	"github.com/shandysiswandi/gobite/internal/pkg/idempotency"
	"github.com/shandysiswandi/gobite/internal/pkg/instrument"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/mfa"
	"github.com/shandysiswandi/gobite/internal/pkg/otp"
	"github.com/shandysiswandi/gobite/internal/pkg/storage"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/pkg/validator"
	"go.opentelemetry.io/otel/trace"
)

type UserRegistrationEvent struct {
	UserID         int64
	Email          string
	FullName       string
	ChallengeToken string
}

type UserForgotPasswordEvent struct {
	UserID         int64
	Email          string
	ChallengeToken string
}

type repoMessaging interface {
	PublishUserRegistration(ctx context.Context, msg UserRegistrationEvent) error
	PublishUserForgotPassword(ctx context.Context, msg UserForgotPasswordEvent) error
}

type repoDB interface {
	GetUserLoginInfo(ctx context.Context, email string) (*entity.UserLoginInfo, error)
	GetUserCredentialInfo(ctx context.Context, id int64) (*entity.UserCredentialInfo, error)
	GetChallengeUserByTokenPurpose(ctx context.Context, token string, p entity.ChallengePurpose) (*entity.ChallengeUser, error)
	GetUserRefreshToken(ctx context.Context, token string) (*entity.UserRefreshToken, error)
	GetUserByEmail(ctx context.Context, email string, includeDeleted bool) (*entity.User, error)
	GetUserList(ctx context.Context, filter entity.UserListFilterData) ([]entity.User, int64, error)
	GetUserByID(ctx context.Context, id int64, includeDeleted bool) (*entity.User, error)
	GetMFAFactorByUserID(ctx context.Context, userID int64, isVerified bool) ([]entity.MFAFactor, error)
	GetMFAFactorByID(ctx context.Context, id int64, userID int64) (*entity.MFAFactor, error)
	GetMFABackupCodeByUserID(ctx context.Context, userID int64) ([]entity.MFABackupCode, error)

	CreateRefreshToken(ctx context.Context, in entity.RefreshToken) error
	CreateChallenge(ctx context.Context, in entity.Challenge) error

	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllRefreshToken(ctx context.Context, userID int64) error
	MarkMFABackupCodeUsed(ctx context.Context, bcID, userID int64) (bool, error)
	UpdateMFALastUsedAt(ctx context.Context, factorID, userID int64) error
	UpdateUserProfile(ctx context.Context, id int64, fullName string) error
	UpdateUserAvatar(ctx context.Context, id int64, avatarURL string) error
	UpdateUserStatus(ctx context.Context, id int64, oldStatus, newStatus entity.UserStatus) error
	UpdateUserCredential(ctx context.Context, userID int64, hash string) error
	MarkUserDeleted(ctx context.Context, id, byID int64) error

	NewMFAFactorTOTP(ctx context.Context, fTOTP entity.MFAFactor, challengeID int64) error
	NewRefreshToken(ctx context.Context, ref entity.RefreshToken, challengeID int64) error
	NewRegistration(ctx context.Context, user entity.NewUser, chal entity.Challenge, hash string) error
	NewBackupCodes(ctx context.Context, userID int64, codes []entity.MFABackupCode, factor *entity.MFAFactor) error
	NewUser(ctx context.Context, user entity.NewUser, hash string) error
	UpsertUsers(ctx context.Context, users []entity.UpsertUser, hashes map[string]string) (created, updated int, err error)
	PatchUser(ctx context.Context, user entity.PatchUser, hash string) error
	VerifyUserRegistration(ctx context.Context, data entity.VerifyUserRegistration) error
	ResetUserPassword(ctx context.Context, userID, challengeID int64, newHash string) error
	VerifyUserMFAFactor(ctx context.Context, userID, challengeID, factorID int64) error
	RotateRefreshToken(ctx context.Context, ro entity.RotateRefreshToken) error

	DeleteChallenge(ctx context.Context, id int64) error
}

type Usecase struct {
	repoDB          repoDB
	repoMessaging   repoMessaging
	idemp           idempotency.Idempotency
	validator       validator.Validator
	cfg             config.Config
	storage         storage.Storage
	hmac            hash.Hash
	bcrypt          hash.Hash
	argon2id        hash.Hash
	mfaEncryptor    mfa.Encryptor
	mfaRecoveryCode mfa.RecoveryCodeGenerator
	uid             uid.NumberID
	uuid            uid.StringID
	oid             uid.StringID
	totp            otp.OTP
	clock           clock.Clocker
	jwt             jwt.JWT
	ins             instrument.Instrumentation
	enforcer        *casbin.Enforcer
	goroutine       *goroutine.Manager
}

type Dependency struct {
	RepoDB          repoDB
	Idempotency     idempotency.Idempotency
	RepoMessaging   repoMessaging
	Validator       validator.Validator
	Config          config.Config
	Storage         storage.Storage
	HMAC            hash.Hash
	Bcrypt          hash.Hash
	Argon2ID        hash.Hash
	MFAEncryptor    mfa.Encryptor
	MFARecoveryCode mfa.RecoveryCodeGenerator
	UID             uid.NumberID
	UUID            uid.StringID
	OID             uid.StringID
	Totp            otp.OTP
	Clock           clock.Clocker
	JWT             jwt.JWT
	Instrument      instrument.Instrumentation
	Enforcer        *casbin.Enforcer
	Goroutine       *goroutine.Manager
}

func New(dep Dependency) *Usecase {
	return &Usecase{
		repoDB:          dep.RepoDB,
		repoMessaging:   dep.RepoMessaging,
		idemp:           dep.Idempotency,
		validator:       dep.Validator,
		bcrypt:          dep.Bcrypt,
		hmac:            dep.HMAC,
		argon2id:        dep.Argon2ID,
		mfaEncryptor:    dep.MFAEncryptor,
		mfaRecoveryCode: dep.MFARecoveryCode,
		cfg:             dep.Config,
		storage:         dep.Storage,
		uid:             dep.UID,
		uuid:            dep.UUID,
		oid:             dep.OID,
		totp:            dep.Totp,
		clock:           dep.Clock,
		jwt:             dep.JWT,
		ins:             dep.Instrument,
		enforcer:        dep.Enforcer,
		goroutine:       dep.Goroutine,
	}
}

func (s *Usecase) startSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return s.ins.Tracer("identity.usecase").Start(ctx, name)
}

func (s *Usecase) ensureUserStatusAllowed(ctx context.Context, userID int64, status entity.UserStatus) error {
	sts := status.Ensure()
	switch sts {
	case entity.UserStatusUnknown:
		slog.WarnContext(ctx, "user account status is unrecognized", "user_id", userID)
		return goerror.NewBusiness("account status is unrecognized", goerror.CodeForbidden)

	case entity.UserStatusUnverified:
		slog.WarnContext(ctx, "user account is unverified", "user_id", userID)
		return goerror.NewBusiness("email not verified", goerror.CodeForbidden)

	case entity.UserStatusBanned:
		slog.WarnContext(ctx, "user account is banned", "user_id", userID)
		return goerror.NewBusiness("account is banned", goerror.CodeForbidden)

	case entity.UserStatusInactive:
		slog.WarnContext(ctx, "user account is deleted", "user_id", userID)
		return goerror.NewBusiness("account is deleted", goerror.CodeForbidden)

	default:
		return nil
	}
}

func (s *Usecase) authenticatedAndAuthorized(ctx context.Context, obj, act string) (*jwt.Claims, error) {
	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return nil, goerror.NewBusiness("Authentication required", goerror.CodeUnauthorized)
	}

	ok, err := s.enforcer.Enforce(clm.Subject, obj, act)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check authorization", "user_id", clm.Subject, "error", err)
		return nil, goerror.NewServer(err)
	}

	if !ok {
		return nil, goerror.NewBusiness("Account not allowed", goerror.CodeForbidden)
	}

	return clm, nil
}

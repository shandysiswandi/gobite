package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgconfig"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkghash"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmail"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgotp"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

type repoCache interface {
	SaveTokensID(ctx context.Context, acID, refID string) error
	DeleteTokensID(ctx context.Context, acID, refID string) error
	IsTokenIDExist(ctx context.Context, id string) (bool, error)
	RotateTokensID(ctx context.Context, oldRefID, newAcID, newRefID string) error
}

type repoDB interface {
	// users
	UserGetByEmail(ctx context.Context, email string) (*domain.User, error)
	UserGetByID(ctx context.Context, id int64) (*domain.User, error)
	UserUpdateStatus(ctx context.Context, id int64, oldStatus, newStatus domain.UserStatus) error

	// user_credentials
	UserCredentialGetByUserID(ctx context.Context, userID int64) (*domain.UserCredential, error)
	UserCredentialUpdate(ctx context.Context, userID int64, hash string) error

	// user + user_credentials (transaction)
	UserRegistration(ctx context.Context, user domain.User, hash string) error

	// mfa_factors
	MfaFactorGetByUserID(ctx context.Context, userID int64) ([]domain.MfaFactor, error)

	// user_password_resets
	UserPasswordResetCreate(ctx context.Context, userID int64, token string, expiresAt time.Time) error

	// user + user_password_resets (transaction)
	UserPasswordResetConsume(ctx context.Context, token string, newHash string, now time.Time) error
}

type Usecase struct {
	repoDB    repoDB
	repoCache repoCache

	validator pkgvalidator.Validator
	cfg       pkgconfig.Config
	hash      pkghash.Hash
	uid       pkguid.NumberID
	uuid      pkguid.StringID
	totp      pkgotp.OTP
	clock     pkgclock.Clocker
	mail      pkgmail.Mail

	jwtTempToken    pkgjwt.JWT[map[string]any]
	jwtAccessToken  pkgjwt.JWT[pkgjwt.AccessTokenPayload]
	jwtRefreshToken pkgjwt.JWT[pkgjwt.RefreshTokenPayload]
}

type Dependency struct {
	RepoDB    repoDB
	RepoCache repoCache

	Validator pkgvalidator.Validator
	Config    pkgconfig.Config
	Hash      pkghash.Hash
	UID       pkguid.NumberID
	UUID      pkguid.StringID
	Totp      pkgotp.OTP
	Clock     pkgclock.Clocker
	Mail      pkgmail.Mail

	JWTTempToken    pkgjwt.JWT[map[string]any]
	JWTAccessToken  pkgjwt.JWT[pkgjwt.AccessTokenPayload]
	JWTRefreshToken pkgjwt.JWT[pkgjwt.RefreshTokenPayload]
}

func NewAuth(dep Dependency) *Usecase {
	return &Usecase{
		repoDB:          dep.RepoDB,
		repoCache:       dep.RepoCache,
		validator:       dep.Validator,
		hash:            dep.Hash,
		cfg:             dep.Config,
		uid:             dep.UID,
		uuid:            dep.UUID,
		totp:            dep.Totp,
		clock:           dep.Clock,
		mail:            dep.Mail,
		jwtTempToken:    dep.JWTTempToken,
		jwtAccessToken:  dep.JWTAccessToken,
		jwtRefreshToken: dep.JWTRefreshToken,
	}
}

func (s *Usecase) getUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.repoDB.UserGetByEmail(ctx, email)
	if errors.Is(err, pkgerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found")
		return nil, pkgerror.NewBusiness("invalid email or password", pkgerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by email", "error", err)
		return nil, pkgerror.NewServer(err)
	}

	if err := s.ensureUserActive(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Usecase) getUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := s.repoDB.UserGetByID(ctx, userID)
	if errors.Is(err, pkgerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found", "user_id", userID)
		return nil, pkgerror.NewBusiness("invalid email or password", pkgerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by id", "user_id", userID, "error", err)
		return nil, pkgerror.NewServer(err)
	}

	if err := s.ensureUserActive(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Usecase) ensureUserActive(ctx context.Context, user *domain.User) error {
	if user.Status == domain.UserStatusUnverified {
		slog.WarnContext(ctx, "user not verified", "user_id", user.ID)
		return pkgerror.NewBusiness("user account is not verified", pkgerror.CodeUnauthorized)
	}

	if user.Status == domain.UserStatusBanned {
		slog.WarnContext(ctx, "user account banned", "user_id", user.ID)
		return pkgerror.NewBusiness("user account is banned", pkgerror.CodeUnauthorized)
	}

	return nil
}

func (s *Usecase) getCredential(ctx context.Context, userID int64) (*domain.UserCredential, error) {
	userCred, err := s.repoDB.UserCredentialGetByUserID(ctx, userID)
	if errors.Is(err, pkgerror.ErrNotFound) {
		slog.WarnContext(ctx, "user credential not found", "user_id", userID)
		return nil, pkgerror.NewBusiness("invalid email or password", pkgerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user credential by user_id", "user_id", userID, "error", err)
		return nil, pkgerror.NewServer(err)
	}

	return userCred, nil
}

func (s *Usecase) issueTokens(ctx context.Context, user *domain.User) (acToken, acJTI, refToken, refJTI string, err error) {
	subject := strconv.FormatInt(user.ID, 10)

	acToken, acJTI, err = s.jwtAccessToken.Generate(subject, pkgjwt.AccessTokenPayload{
		UserID: user.ID,
		Email:  user.Email,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate access jwt token", "user_id", user.ID, "error", err)
		return "", "", "", "", pkgerror.NewServer(err)
	}

	refToken, refJTI, err = s.jwtRefreshToken.Generate(subject, pkgjwt.RefreshTokenPayload{UserID: user.ID})
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate refresh jwt token", "user_id", user.ID, "error", err)
		return "", "", "", "", pkgerror.NewServer(err)
	}

	return acToken, acJTI, refToken, refJTI, nil
}

package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
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

	// user_credentials
	UserCredentialGetByUserID(ctx context.Context, userID int64) (*domain.UserCredential, error)
	UserCredentialUpdate(ctx context.Context, userID int64, hash string) error

	// user + user_credentials (transaction)
	UserRegistration(ctx context.Context, user domain.User, hash string) error

	// mfa_factors
	MfaFactorGetByUserID(ctx context.Context, userID int64) ([]domain.MfaFactor, error)
}

type Usecase struct {
	repoDB    repoDB
	repoCache repoCache
	mail      pkgmail.Mail

	validator pkgvalidator.Validator
	hash      pkghash.Hash
	uid       pkguid.NumberID
	totp      pkgotp.OTP
	clock     pkgclock.Clocker

	jwtTempToken    pkgjwt.JWT[map[string]any]
	jwtAccessToken  pkgjwt.JWT[pkgjwt.AccessTokenPayload]
	jwtRefreshToken pkgjwt.JWT[pkgjwt.RefreshTokenPayload]
}

type Dependency struct {
	RepoDB    repoDB
	RepoCache repoCache
	Mail      pkgmail.Mail

	Validator pkgvalidator.Validator
	Hash      pkghash.Hash
	UID       pkguid.NumberID
	Totp      pkgotp.OTP
	Clock     pkgclock.Clocker

	JWTTempToken    pkgjwt.JWT[map[string]any]
	JWTAccessToken  pkgjwt.JWT[pkgjwt.AccessTokenPayload]
	JWTRefreshToken pkgjwt.JWT[pkgjwt.RefreshTokenPayload]
}

func NewAuth(dep Dependency) *Usecase {
	return &Usecase{
		repoDB:          dep.RepoDB,
		repoCache:       dep.RepoCache,
		mail:            dep.Mail,
		validator:       dep.Validator,
		hash:            dep.Hash,
		uid:             dep.UID,
		totp:            dep.Totp,
		clock:           dep.Clock,
		jwtTempToken:    dep.JWTTempToken,
		jwtAccessToken:  dep.JWTAccessToken,
		jwtRefreshToken: dep.JWTRefreshToken,
	}
}

func (s *Usecase) getUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.repoDB.UserGetByEmail(ctx, email)
	if errors.Is(err, pkgerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found")
		return nil, pkgerror.ErrAuthUnauthenticated
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by email", "error", err)
		return nil, err
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
		return nil, pkgerror.ErrAuthUnauthenticated
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by id", "user_id", userID, "error", err)
		return nil, err
	}

	if err := s.ensureUserActive(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Usecase) ensureUserActive(ctx context.Context, user *domain.User) error {
	if user.Status == domain.UserStatusUnverified {
		slog.WarnContext(ctx, "user not verified", "user_id", user.ID)
		return pkgerror.ErrAuthNotVerified
	}

	if user.Status == domain.UserStatusBanned {
		slog.WarnContext(ctx, "user account banned", "user_id", user.ID)
		return pkgerror.ErrAuthBanned
	}

	return nil
}

func (s *Usecase) getCredential(ctx context.Context, userID int64) (*domain.UserCredential, error) {
	userCred, err := s.repoDB.UserCredentialGetByUserID(ctx, userID)
	if errors.Is(err, pkgerror.ErrNotFound) {
		slog.WarnContext(ctx, "user credential not found", "user_id", userID)
		return nil, pkgerror.ErrAuthUnauthenticated
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user credential by user_id", "user_id", userID, "error", err)
		return nil, err
	}

	return userCred, nil
}

func (s *Usecase) issueTokens(ctx context.Context, user *domain.User) (acToken, acJTI, refToken, refJTI string, err error) {
	subject := strconv.FormatInt(user.ID, 10)

	acToken, acJTI, err = s.jwtAccessToken.Generate(subject, pkgjwt.AccessTokenPayload{
		UserID: subject,
		Email:  user.Email,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate access jwt token", "user_id", user.ID, "error", err)
		return "", "", "", "", err
	}

	refToken, refJTI, err = s.jwtRefreshToken.Generate(subject, pkgjwt.RefreshTokenPayload{Message: "hack me"})
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate refresh jwt token", "user_id", user.ID, "error", err)
		return "", "", "", "", err
	}

	return acToken, acJTI, refToken, refJTI, nil
}

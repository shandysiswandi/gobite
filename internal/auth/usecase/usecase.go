package usecase

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkghash"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmail"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

type Usecase struct {
	repoDB    repoDB
	repoCache repoCache
	mail      pkgmail.Mail

	validator pkgvalidator.Validator
	hash      pkghash.Hash
	uid       pkguid.NumberID

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
		jwtTempToken:    dep.JWTTempToken,
		jwtAccessToken:  dep.JWTAccessToken,
		jwtRefreshToken: dep.JWTRefreshToken,
	}
}

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

	// user + user_credentials (transaction)
	UserRegistration(ctx context.Context, user domain.User, hash string) error

	// mfa_factors
	MfaFactorGetByUserID(ctx context.Context, userID int64) ([]domain.MfaFactor, error)
}

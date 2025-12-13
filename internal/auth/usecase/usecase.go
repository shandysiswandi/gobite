package usecase

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkghash"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

type Usecase struct {
	repoDB    repoDB
	repoCache repoCache

	validator pkgvalidator.Validator
	hash      pkghash.Hash
	uid       pkguid.StringID

	jwtTempToken    pkgjwt.JWT[map[string]any]
	jwtAccessToken  pkgjwt.JWT[pkgjwt.AccessTokenPayload]
	jwtRefreshToken pkgjwt.JWT[pkgjwt.RefreshTokenPayload]
}

type Dependency struct {
	RepoDB    repoDB
	RepoCache repoCache

	Validator pkgvalidator.Validator
	Hash      pkghash.Hash
	UID       pkguid.StringID

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
		uid:             dep.UID,
		jwtTempToken:    dep.JWTTempToken,
		jwtAccessToken:  dep.JWTAccessToken,
		jwtRefreshToken: dep.JWTRefreshToken,
	}
}

type repoCache interface {
	SaveTokensID(ctx context.Context, acID, refID string) error
}

type repoDB interface {
	// users
	UserGetByEmail(ctx context.Context, email string) (*domain.User, error)

	// user_credentials
	UserCredentialGetByUserID(ctx context.Context, userID int64) (*domain.UserCredential, error)

	// mfa_factors
	MfaFactorGetByUserID(ctx context.Context, userID int64) ([]domain.MfaFactor, error)
}

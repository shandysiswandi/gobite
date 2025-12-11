package usecase

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkghash"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

type Usecase struct {
	repoDB    repoDB
	repoCache repoCache

	validator pkgvalidator.Validator
	hash      pkghash.Hash
}

type Dependency struct {
	RepoDB    repoDB
	RepoCache repoCache

	Validator pkgvalidator.Validator
	Hash      pkghash.Hash
}

func NewAuth(dep Dependency) *Usecase {
	return &Usecase{
		repoDB:    dep.RepoDB,
		repoCache: dep.RepoCache,
		validator: dep.Validator,
		hash:      dep.Hash,
	}
}

type repoCache interface {
	//
}

type repoDB interface {
	// users
	UserGetByEmail(ctx context.Context, email string) (*domain.User, error)

	// user_credentials
	UserCredentialGetByUserID(ctx context.Context, userID int64) (*domain.UserCredential, error)

	// mfa_factors
	MfaFactorGetByUserID(ctx context.Context, userID int64) ([]domain.MfaFactor, error)
}

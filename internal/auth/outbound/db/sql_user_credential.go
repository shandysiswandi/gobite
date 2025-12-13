package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgsql"
)

func userCredentialFromSQL(u pkgsql.UserCredential) *domain.UserCredential {
	return &domain.UserCredential{
		UserID:    u.UserID,
		Password:  u.Password,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func (s *SQL) UserCredentialGetByUserID(ctx context.Context, userID int64) (*domain.UserCredential, error) {
	result, err := s.query.UserCredentialGetByUserID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pkgerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return userCredentialFromSQL(result), nil
}

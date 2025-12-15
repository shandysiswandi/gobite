package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/shandysiswandi/gobite/internal/auth/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgsql"
)

func userCredentialFromSQL(u pkgsql.UserCredential) *entity.UserCredential {
	item := &entity.UserCredential{
		UserID:   u.UserID,
		Password: u.Password,
	}

	if u.UpdatedAt.Valid {
		item.UpdatedAt = u.UpdatedAt.Time
	}

	return item
}

func (s *SQL) UserCredentialGetByUserID(ctx context.Context, userID int64) (*entity.UserCredential, error) {
	result, err := s.query.UserCredentialGetByUserID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pkgerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return userCredentialFromSQL(result), nil
}

func (s *SQL) UserCredentialUpdate(ctx context.Context, userID int64, hash string) error {
	return s.query.UserCredentialUpdate(ctx, pkgsql.UserCredentialUpdateParams{
		Password: hash,
		UserID:   userID,
	})
}

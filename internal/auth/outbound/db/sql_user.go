package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/shandysiswandi/gobite/internal/auth/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgsql"
)

func userFromSQL(u pkgsql.User) *entity.User {
	item := &entity.User{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		AvatarURL: u.AvatarUrl,
		Status:    u.Status,
	}

	if u.CreatedAt.Valid {
		item.CreatedAt = u.CreatedAt.Time
	}
	if u.UpdatedAt.Valid {
		item.UpdatedAt = u.UpdatedAt.Time
	}

	return item
}

func (s *SQL) UserGetByEmail(ctx context.Context, email string) (*entity.User, error) {
	result, err := s.query.UserGetByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pkgerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return userFromSQL(result), nil
}

func (s *SQL) UserGetByID(ctx context.Context, id int64) (*entity.User, error) {
	result, err := s.query.UserGetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pkgerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return userFromSQL(result), nil
}

func (s *SQL) UserUpdateStatus(ctx context.Context, id int64, oldStatus, newStatus entity.UserStatus) error {
	return s.query.UserUpdateStatus(ctx, pkgsql.UserUpdateStatusParams{
		ID:        id,
		NewStatus: newStatus,
		OldStatus: oldStatus,
	})
}

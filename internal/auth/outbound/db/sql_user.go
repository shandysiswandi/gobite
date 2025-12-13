package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgsql"
)

func userFromSQL(u pkgsql.User) *domain.User {
	return &domain.User{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		AvatarURL: u.AvatarUrl,
		Status:    domain.UserStatus(u.Status),
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func (s *SQL) UserGetByEmail(ctx context.Context, email string) (*domain.User, error) {
	result, err := s.query.UserGetByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pkgerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return userFromSQL(result), nil
}

func (s *SQL) UserGetByID(ctx context.Context, id int64) (*domain.User, error) {
	result, err := s.query.UserGetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pkgerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return userFromSQL(result), nil
}

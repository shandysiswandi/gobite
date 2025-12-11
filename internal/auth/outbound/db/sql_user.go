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
		AvatarURL: u.AvatarUrl.String,
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

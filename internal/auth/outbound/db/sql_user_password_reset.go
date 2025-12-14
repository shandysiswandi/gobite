package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgsql"
)

func (s *SQL) UserPasswordResetCreate(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	return s.query.UserPasswordResetCreate(ctx, pkgsql.UserPasswordResetCreateParams{
		UserID: userID,
		Token:  token,
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
	})
}

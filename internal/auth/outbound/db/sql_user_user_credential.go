package db

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgsql"
)

func (s *SQL) UserRegistration(ctx context.Context, user domain.User, hash string) error {
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	if err := wtx.UserCreate(ctx, pkgsql.UserCreateParams{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarUrl: user.AvatarURL,
		Status:    user.Status.Value(),
	}); err != nil {
		return err
	}

	if err := wtx.UserCredentialCreate(ctx, pkgsql.UserCredentialCreateParams{
		UserID:   user.ID,
		Password: hash,
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

package db

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgsql"
)

func (s *SQL) UserPasswordResetConsume(ctx context.Context, token string, newHash string, now time.Time) error {
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

	ups, err := wtx.UserPasswordResetGetValidForUpdate(ctx, pkgsql.UserPasswordResetGetValidForUpdateParams{
		Token: token,
		Now:   pgtype.Timestamptz{Time: now, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return pkgerror.ErrNotFound
	}
	if err != nil {
		return err
	}

	user, err := wtx.UserGetValidByIDForUpdate(ctx, ups.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return pkgerror.ErrNotFound
	}
	if err != nil {
		return err
	}

	if err := wtx.UserCredentialUpdate(ctx, pkgsql.UserCredentialUpdateParams{
		UserID:   user.ID,
		Password: newHash,
	}); err != nil {
		return err
	}

	if err := wtx.UserPasswordResetMarkUsed(ctx, pkgsql.UserPasswordResetMarkUsedParams{
		ID:     ups.ID,
		UsedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

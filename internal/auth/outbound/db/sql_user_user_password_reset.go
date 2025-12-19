package db

// import (
// 	"context"
// 	"errors"
// 	"log/slog"
// 	"time"

// 	"github.com/jackc/pgx/v5"
// 	"github.com/jackc/pgx/v5/pgtype"
// 	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
// 	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
// )

// func (s *DB) UserPasswordResetConsume(ctx context.Context, token string, newHash string, now time.Time) error {
// 	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	defer func() {
// 		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
// 			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
// 		}
// 	}()

// 	wtx := s.query.WithTx(tx)

// 	ups, err := wtx.UserPasswordResetGetValidForUpdate(ctx, sqlc.UserPasswordResetGetValidForUpdateParams{
// 		Token: token,
// 		Now:   pgtype.Timestamptz{Time: now, Valid: true},
// 	})
// 	if errors.Is(err, pgx.ErrNoRows) {
// 		return goerror.ErrNotFound
// 	}
// 	if err != nil {
// 		return err
// 	}

// 	user, err := wtx.UserGetValidByIDForUpdate(ctx, ups.UserID)
// 	if errors.Is(err, pgx.ErrNoRows) {
// 		return goerror.ErrNotFound
// 	}
// 	if err != nil {
// 		return err
// 	}

// 	if err := wtx.UserCredentialUpdate(ctx, sqlc.UserCredentialUpdateParams{
// 		UserID:   user.ID,
// 		Password: newHash,
// 	}); err != nil {
// 		return err
// 	}

// 	if err := wtx.UserPasswordResetMarkUsed(ctx, sqlc.UserPasswordResetMarkUsedParams{
// 		ID:     ups.ID,
// 		UsedAt: pgtype.Timestamptz{Time: now, Valid: true},
// 	}); err != nil {
// 		return err
// 	}

// 	return tx.Commit(ctx)
// }

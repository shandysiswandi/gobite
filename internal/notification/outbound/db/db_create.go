package db

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
)

func (s *DB) RegisterUserDevice(ctx context.Context, userID int64, deviceToken, platform string) (err error) {
	ctx, span := s.startSpan(ctx, "RegisterUserDevice")
	defer func() { s.endSpan(span, err) }()

	err = s.query.RegisterNotificationUserDevice(ctx, sqlc.RegisterNotificationUserDeviceParams{
		UserID:      userID,
		DeviceToken: deviceToken,
		Platform:    platform,
	})
	return s.mapError(err)
}

func (s *DB) CreateNotification(ctx context.Context, data entity.CreateNotification) (err error) {
	ctx, span := s.startSpan(ctx, "CreateNotification")
	defer func() { s.endSpan(span, err) }()

	err = s.query.CreateNotification(ctx, sqlc.CreateNotificationParams{
		ID:         data.ID,
		UserID:     data.UserID,
		CategoryID: data.CategoryID,
		TriggerKey: data.TriggerKey.String(),
		Data:       data.Data,
		Metadata:   data.Metadata,
	})
	return s.mapError(err)
}

func (s *DB) CreateNotificationWithDeliveryLog(ctx context.Context, n entity.CreateNotification, dl entity.CreateDeliveryLog) (_ int64, err error) {
	ctx, span := s.startSpan(ctx, "CreateNotificationWithDeliveryLog")
	defer func() { s.endSpan(span, err) }()

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	if err := wtx.CreateNotification(ctx, sqlc.CreateNotificationParams{
		ID:         n.ID,
		UserID:     n.UserID,
		CategoryID: n.CategoryID,
		TriggerKey: n.TriggerKey.String(),
		Data:       n.Data,
		Metadata:   n.Metadata,
	}); err != nil {
		return 0, s.mapError(err)
	}

	logID, err := wtx.CreateNotificationDeliveryLog(ctx, sqlc.CreateNotificationDeliveryLogParams(dl))
	if err != nil {
		return 0, s.mapError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return logID, nil
}

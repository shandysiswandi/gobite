package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
)

func (s *DB) MarkNotificationRead(ctx context.Context, userID, notificationID int64) (_ bool, err error) {
	ctx, span := s.startSpan(ctx, "MarkNotificationRead")
	defer func() { s.endSpan(span, err) }()

	rows, err := s.query.MarkNotificationRead(ctx, sqlc.MarkNotificationReadParams{
		ID:     notificationID,
		UserID: userID,
	})
	if err != nil {
		return false, s.mapError(err)
	}

	return rows == 1, nil
}

func (s *DB) MarkNotificationsReadAll(ctx context.Context, userID int64) (_ int64, err error) {
	ctx, span := s.startSpan(ctx, "MarkNotificationsReadAll")
	defer func() { s.endSpan(span, err) }()

	rows, err := s.query.MarkNotificationsReadAll(ctx, userID)
	if err != nil {
		return 0, s.mapError(err)
	}

	return rows, nil
}

func (s *DB) SoftDeleteNotification(ctx context.Context, userID, notificationID int64) (_ bool, err error) {
	ctx, span := s.startSpan(ctx, "SoftDeleteNotification")
	defer func() { s.endSpan(span, err) }()

	rows, err := s.query.SoftDeleteNotification(ctx, sqlc.SoftDeleteNotificationParams{
		ID:     notificationID,
		UserID: userID,
	})
	if err != nil {
		return false, s.mapError(err)
	}

	return rows == 1, nil
}

func (s *DB) UpdateDeliveryLogStatus(ctx context.Context, u entity.UpdateDeliveryLog) (err error) {
	ctx, span := s.startSpan(ctx, "UpdateDeliveryLogStatus")
	defer func() { s.endSpan(span, err) }()

	var next pgtype.Timestamptz
	if u.NextRetryAt != nil {
		next = pgtype.Timestamptz{Time: *u.NextRetryAt, Valid: true}
	}

	err = s.query.UpdateNotificationDeliveryLogStatus(ctx, sqlc.UpdateNotificationDeliveryLogStatusParams{
		Status:           u.Status,
		ProviderResponse: u.ProviderResponse,
		NextRetryAt:      next,
		ID:               u.ID,
	})
	return s.mapError(err)
}

package outbound

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

func (s *SQL) NotificationCreate(ctx context.Context, n entity.NotificationCreate) error {
	return s.query.NotificationCreate(ctx, sqlc.NotificationCreateParams{
		ID:         n.ID,
		UserID:     n.UserID,
		CategoryID: n.CategoryID,
		TriggerKey: n.TriggerKey,
		Data:       n.Data,
		Metadata:   n.Metadata,
	})
}

func (s *SQL) NotificationCreateWithDeliveryLog(ctx context.Context, n entity.NotificationCreate, log entity.DeliveryLogCreate) (logID int64, err error) {
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	qtx := s.query.WithTx(tx)

	if err := qtx.NotificationCreate(ctx, sqlc.NotificationCreateParams{
		ID:         n.ID,
		UserID:     n.UserID,
		CategoryID: n.CategoryID,
		TriggerKey: n.TriggerKey,
		Data:       n.Data,
		Metadata:   n.Metadata,
	}); err != nil {
		return 0, err
	}

	logID, err = qtx.NotificationDeliveryLogCreate(ctx, sqlc.NotificationDeliveryLogCreateParams{
		NotificationID: log.NotificationID,
		Channel:        log.Channel,
		Status:         log.Status,
	})
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return logID, nil
}

func (s *SQL) NotificationGetUserNotificationPaginate(ctx context.Context, userID int64, limit, offset int32) ([]entity.UserNotificationItem, error) {
	rows, err := s.query.NotificationGetUserNotificationPaginate(ctx, sqlc.NotificationGetUserNotificationPaginateParams{
		Limit:  limit,
		Offset: offset,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]entity.UserNotificationItem, 0, len(rows))
	for _, row := range rows {
		item := entity.UserNotificationItem{
			ID:                  row.ID,
			TriggerKey:          row.TriggerKey,
			Data:                row.Data,
			Metadata:            row.Metadata,
			CategoryName:        row.CategoryName,
			CategoryDescription: row.CategoryDescription,
		}

		if row.ReadAt.Valid {
			t := row.ReadAt.Time
			item.ReadAt = &t
		}

		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time
		}

		items = append(items, item)
	}

	return items, nil
}

func (s *SQL) NotificationCountUnread(ctx context.Context, userID int64) (int64, error) {
	return s.query.NotificationCountUnread(ctx, userID)
}

func (s *SQL) NotificationMarkRead(ctx context.Context, userID, notificationID int64) error {
	return s.query.NotificationMarkRead(ctx, sqlc.NotificationMarkReadParams{
		ID:     notificationID,
		UserID: userID,
	})
}

func (s *SQL) NotificationsMarkAllRead(ctx context.Context, userID int64) error {
	return s.query.NotificationsMarkAllRead(ctx, userID)
}

func (s *SQL) NotificationSoftDelete(ctx context.Context, userID, notificationID int64) error {
	return s.query.NotificationSoftDelete(ctx, sqlc.NotificationSoftDeleteParams{
		ID:     notificationID,
		UserID: userID,
	})
}

func (s *SQL) NotificationTemplateGetByTrigger(ctx context.Context, triggerKey string, channel entity.Channel) (*entity.Template, error) {
	row, err := s.query.NotificationTemplateGetByTrigger(ctx, sqlc.NotificationTemplateGetByTriggerParams{
		TriggerKey: triggerKey,
		Channel:    channel,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, goerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &entity.Template{
		ID:         row.ID,
		TriggerKey: row.TriggerKey,
		CategoryID: row.CategoryID,
		Channel:    row.Channel,
		Subject:    row.Subject,
		Body:       row.Body,
	}, nil
}

func (s *SQL) NotificationDeliveryLogUpdateStatus(ctx context.Context, id int64, status entity.DeliveryStatus, providerResponse valueobject.JSONMap, nextRetryAt *time.Time) error {
	var next pgtype.Timestamptz
	if nextRetryAt != nil {
		next = pgtype.Timestamptz{Time: *nextRetryAt, Valid: true}
	}

	return s.query.NotificationDeliveryLogUpdateStatus(ctx, sqlc.NotificationDeliveryLogUpdateStatusParams{
		ID:               id,
		Status:           status,
		ProviderResponse: providerResponse,
		NextRetryAt:      next,
	})
}

func (s *SQL) NotificationUserDeviceRegister(ctx context.Context, userID int64, deviceToken, platform string) error {
	return s.query.NotificationUserDeviceRegister(ctx, sqlc.NotificationUserDeviceRegisterParams{
		UserID:      userID,
		DeviceToken: deviceToken,
		Platform:    platform,
	})
}

func (s *SQL) NotificationUserDeviceRemove(ctx context.Context, deviceToken string) error {
	return s.query.NotificationUserDeviceRemove(ctx, deviceToken)
}

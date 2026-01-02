package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
)

func (s *DB) GetTemplateByTriggerChannel(ctx context.Context, tk entity.TriggerKey, ch entity.Channel) (_ *entity.Template, err error) {
	ctx, span := s.startSpan(ctx, "GetTemplateByTriggerChannel")
	defer func() { s.endSpan(span, err) }()

	row, err := s.query.GetNotificationTemplateByTriggerChannel(ctx, sqlc.GetNotificationTemplateByTriggerChannelParams{
		TriggerKey: tk.String(),
		Channel:    ch,
	})
	if err != nil {
		return nil, s.mapError(err)
	}

	return &entity.Template{
		ID:         row.ID,
		TriggerKey: entity.TriggerKey(row.TriggerKey),
		CategoryID: row.CategoryID,
		Channel:    row.Channel,
		Subject:    row.Subject,
		Body:       row.Body,
	}, nil
}

func (s *DB) ListCategories(ctx context.Context) (_ []entity.Category, err error) {
	ctx, span := s.startSpan(ctx, "ListCategories")
	defer func() { s.endSpan(span, err) }()

	rows, err := s.query.ListNotificationCategories(ctx)
	if err != nil {
		return nil, s.mapError(err)
	}

	items := make([]entity.Category, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.Category{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			IsMandatory: row.IsMandatory,
		})
	}

	return items, nil
}

func (s *DB) ListUserSettings(ctx context.Context, userID int64) (_ []entity.UserSetting, err error) {
	ctx, span := s.startSpan(ctx, "ListUserSettings")
	defer func() { s.endSpan(span, err) }()

	rows, err := s.query.ListNotificationUserSettings(ctx, userID)
	if err != nil {
		return nil, s.mapError(err)
	}

	items := make([]entity.UserSetting, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.UserSetting{
			CategoryID: row.CategoryID,
			Channel:    row.Channel,
			IsEnabled:  row.IsEnabled,
		})
	}

	return items, nil
}

func (s *DB) ListNotifications(ctx context.Context, userID int64, status entity.NotificationStatus, limit, offset int32) (_ []entity.NotificationItem, err error) {
	ctx, span := s.startSpan(ctx, "ListNotifications")
	defer func() { s.endSpan(span, err) }()

	var rows []sqlc.ListNotificationsByUserAllRow
	switch status {
	case entity.NotificationStatusUnread:
		unread, qErr := s.query.ListNotificationsByUserUnread(ctx, sqlc.ListNotificationsByUserUnreadParams{
			UserID:     userID,
			PageLimit:  limit,
			PageOffset: offset,
		})
		if qErr != nil {
			return nil, s.mapError(qErr)
		}
		rows = make([]sqlc.ListNotificationsByUserAllRow, 0, len(unread))
		for _, row := range unread {
			rows = append(rows, sqlc.ListNotificationsByUserAllRow(row))
		}
	case entity.NotificationStatusRead:
		read, qErr := s.query.ListNotificationsByUserRead(ctx, sqlc.ListNotificationsByUserReadParams{
			UserID:     userID,
			PageLimit:  limit,
			PageOffset: offset,
		})
		if qErr != nil {
			return nil, s.mapError(qErr)
		}
		rows = make([]sqlc.ListNotificationsByUserAllRow, 0, len(read))
		for _, row := range read {
			rows = append(rows, sqlc.ListNotificationsByUserAllRow(row))
		}
	default:
		all, qErr := s.query.ListNotificationsByUserAll(ctx, sqlc.ListNotificationsByUserAllParams{
			UserID:     userID,
			PageLimit:  limit,
			PageOffset: offset,
		})
		if qErr != nil {
			return nil, s.mapError(qErr)
		}
		rows = all
	}

	items := make([]entity.NotificationItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.NotificationItem{
			ID:         row.ID,
			CategoryID: row.CategoryID,
			TriggerKey: entity.TriggerKey(row.TriggerKey),
			Data:       row.Data,
			Metadata:   row.Metadata,
			ReadAt:     timePtrFromPgTimestamptz(row.ReadAt),
			CreatedAt:  timeFromPgTimestamptz(row.CreatedAt),
		})
	}

	return items, nil
}

func (s *DB) CountUnreadNotifications(ctx context.Context, userID int64) (_ int64, err error) {
	ctx, span := s.startSpan(ctx, "CountUnreadNotifications")
	defer func() { s.endSpan(span, err) }()

	count, err := s.query.CountNotificationsUnread(ctx, userID)
	return count, s.mapError(err)
}

func timePtrFromPgTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}

	tt := t.Time
	return &tt
}

func timeFromPgTimestamptz(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}

	return t.Time
}

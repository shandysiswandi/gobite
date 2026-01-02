package db

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
)

func (s *DB) UpsertUserSettings(ctx context.Context, userID int64, settings []entity.UserSetting) (err error) {
	ctx, span := s.startSpan(ctx, "UpsertUserSettings")
	defer func() { s.endSpan(span, err) }()

	if len(settings) == 0 {
		return nil
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return s.mapError(err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := s.query.WithTx(tx)
	for _, setting := range settings {
		err = qtx.UpsertNotificationUserSetting(ctx, sqlc.UpsertNotificationUserSettingParams{
			UserID:     userID,
			CategoryID: setting.CategoryID,
			Channel:    setting.Channel,
			IsEnabled:  setting.IsEnabled,
		})
		if err != nil {
			return s.mapError(err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return s.mapError(err)
	}

	return nil
}

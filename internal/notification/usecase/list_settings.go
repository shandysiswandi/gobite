package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

func (s *Usecase) ListSettings(ctx context.Context) (_ []entity.UserSetting, err error) {
	ctx, span := s.startSpan(ctx, "ListSettings")
	defer span.End()

	clm, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	categories, err := s.repoDB.ListCategories(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo list notification categories", "error", err)
		return nil, goerror.NewServer(err)
	}

	settings, err := s.repoDB.ListUserSettings(ctx, clm.UserID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo list notification settings", "user_id", clm.UserID, "error", err)
		return nil, goerror.NewServer(err)
	}

	settingMap := make(map[int64]map[entity.Channel]bool, len(categories))
	for _, setting := range settings {
		ch := setting.Channel
		if ch == entity.ChannelUnknown {
			ch = entity.ChannelInApp
		}
		if _, ok := settingMap[setting.CategoryID]; !ok {
			settingMap[setting.CategoryID] = map[entity.Channel]bool{}
		}
		settingMap[setting.CategoryID][ch] = setting.IsEnabled
	}

	channels := []entity.Channel{
		entity.ChannelInApp,
		entity.ChannelEmail,
		entity.ChannelSMS,
		entity.ChannelPush,
	}

	items := make([]entity.UserSetting, 0, len(categories)*len(channels))
	for _, category := range categories {
		for _, ch := range channels {
			isEnabled := true
			if v, ok := settingMap[category.ID][ch]; ok {
				isEnabled = v
			}
			if category.IsMandatory {
				isEnabled = true
			}
			items = append(items, entity.UserSetting{
				CategoryID: category.ID,
				Channel:    ch,
				IsEnabled:  isEnabled,
			})
		}
	}

	return items, nil
}

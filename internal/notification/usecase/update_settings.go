package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type UpdateSettingsInput struct {
	Settings []UpdateSettingInput `validate:"required,min=1,dive"`
}

type UpdateSettingInput struct {
	CategoryID int64  `validate:"required,gt=0"`
	Channel    string `validate:"required,oneof=in_app email sms push"`
	IsEnabled  bool   `validate:""`
}

func (s *Usecase) UpdateSettings(ctx context.Context, in UpdateSettingsInput) error {
	ctx, span := s.startSpan(ctx, "UpdateSettings")
	defer span.End()

	clm, err := s.requireAuth(ctx)
	if err != nil {
		return err
	}

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	categories, err := s.repoDB.ListCategories(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo list notification categories", "error", err)
		return goerror.NewServer(err)
	}

	categoryMap := make(map[int64]entity.Category, len(categories))
	for _, category := range categories {
		categoryMap[category.ID] = category
	}

	for _, setting := range in.Settings {
		category, ok := categoryMap[setting.CategoryID]
		if !ok {
			return goerror.NewInvalidInput(errors.New("invalid category_id"))
		}
		if category.IsMandatory && !setting.IsEnabled {
			return goerror.NewInvalidInput(errors.New("category is mandatory"))
		}

		channel, ok := channelFromString(setting.Channel)
		if !ok {
			return goerror.NewInvalidInput(errors.New("invalid channel"))
		}

		if err := s.repoDB.UpsertUserSetting(ctx, clm.UserID, entity.UserSetting{
			CategoryID: setting.CategoryID,
			Channel:    channel,
			IsEnabled:  setting.IsEnabled,
		}); err != nil {
			slog.ErrorContext(ctx, "failed to repo upsert notification setting", "user_id", clm.UserID, "category_id", setting.CategoryID, "error", err)
			return goerror.NewServer(err)
		}
	}

	return nil
}

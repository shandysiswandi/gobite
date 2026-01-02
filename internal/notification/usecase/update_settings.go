package usecase

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type UpdateSettingsInput struct {
	Settings []UpdateSettingInput `validate:"required,min=1,dive"`
}

type UpdateSettingInput struct {
	CategoryID int64  `validate:"required,gt=0"`
	Channel    string `validate:"required,lowercase,oneof=in_app email sms push"`
	IsEnabled  bool
}

func (s *Usecase) UpdateSettings(ctx context.Context, in UpdateSettingsInput) error {
	ctx, span := s.startSpan(ctx, "UpdateSettings")
	defer span.End()

	clm, err := s.requireAuth(ctx)
	if err != nil {
		return err
	}

	if err := s.validator.Validate(in); err != nil {
		slog.Error("error mas", "ini", err)
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

	settings := make([]entity.UserSetting, 0, len(in.Settings))
	for _, setting := range in.Settings {
		category, ok := categoryMap[setting.CategoryID]
		if !ok {
			return goerror.NewBusiness("category not found for category_id = "+strconv.FormatInt(setting.CategoryID, 10), goerror.CodeNotFound)
		}
		if category.IsMandatory && !setting.IsEnabled {
			return goerror.NewBusiness("mandatory category cannot be disabled for category_id = "+strconv.FormatInt(setting.CategoryID, 10), goerror.CodeInvalidFormat)
		}

		channel := entity.ChannelFromString(setting.Channel)
		if channel == entity.ChannelUnknown {
			return goerror.NewBusiness("channel is not supported", goerror.CodeInvalidFormat)
		}

		settings = append(settings, entity.UserSetting{
			CategoryID: setting.CategoryID,
			Channel:    channel,
			IsEnabled:  setting.IsEnabled,
		})
	}

	if err := s.repoDB.UpsertUserSettings(ctx, clm.UserID, settings); err != nil {
		slog.ErrorContext(ctx, "failed to repo upsert notification settings", "user_id", clm.UserID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}

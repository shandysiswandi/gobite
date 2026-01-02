package usecase

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

type (
	ConsumeUserForgotPasswordInput struct {
		UserID int64  `validate:"required,gt=0"`
		Email  string `validate:"required,email"`
		Token  string `validate:"required"`
	}
)

func (s *Usecase) ConsumeUserForgotPassword(ctx context.Context, in ConsumeUserForgotPasswordInput) error {
	ctx, span := s.startSpan(ctx, "ConsumeUserForgotPassword")
	defer span.End()

	if err := s.validator.Validate(in); err != nil {
		slog.ErrorContext(ctx, "Validation failed", "error", err)
		return nil
	}

	data := s.baseEmailTemplateData()
	data["reset_url"] = s.cfg.GetString("app.web") + "/reset-password?token=" + url.QueryEscape(in.Token)

	s.sendEmailNotification(ctx, emailNotificationInput{
		UserID:       in.UserID,
		Email:        in.Email,
		TriggerKey:   entity.TriggerKeyPasswordReset,
		TemplateData: data,
		NotificationData: valueobject.JSONMap{
			"user_id": in.UserID,
			"email":   in.Email,
		},
	})

	return nil
}

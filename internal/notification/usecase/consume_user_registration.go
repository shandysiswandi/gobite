package usecase

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

type (
	ConsumeUserRegistrationInput struct {
		UserID   int64  `validate:"required,gt=0"`
		Email    string `validate:"required,email"`
		FullName string `validate:"required,min=5,max=100,alphaspace"`
		Token    string `validate:"required"`
	}
)

func (s *Usecase) ConsumeUserRegistration(ctx context.Context, in ConsumeUserRegistrationInput) error {
	ctx, span := s.startSpan(ctx, "ConsumeUserRegistration")
	defer span.End()

	if err := s.validator.Validate(in); err != nil {
		slog.ErrorContext(ctx, "Validation failed", "error", err)
		return nil
	}

	data := s.baseEmailTemplateData()
	data["verify_url"] = s.cfg.GetString("app.web") + "/verify-email?token=" + url.QueryEscape(in.Token)

	s.sendEmailNotification(ctx, emailNotificationInput{
		UserID:       in.UserID,
		Email:        in.Email,
		TriggerKey:   entity.TriggerKeyEmailVerify,
		TemplateData: data,
		NotificationData: valueobject.JSONMap{
			"user_id":   in.UserID,
			"email":     in.Email,
			"full_name": in.FullName,
		},
	})
	s.createWelcomeNotification(ctx, in)

	return nil
}

func (s *Usecase) createWelcomeNotification(ctx context.Context, in ConsumeUserRegistrationInput) {
	tpl := s.getTemplate(ctx, entity.TriggerKeyUserWelcome, entity.ChannelInApp)
	if tpl == nil {
		return
	}

	n := entity.CreateNotification{
		ID:         s.uid.Generate(),
		UserID:     in.UserID,
		CategoryID: tpl.CategoryID,
		TriggerKey: tpl.TriggerKey,
		Data:       valueobject.JSONMap{"full_name": in.FullName},
		Metadata:   valueobject.JSONMap{},
	}
	if err := s.repoDB.CreateNotification(ctx, n); err != nil {
		slog.ErrorContext(ctx, "failed to repo create notification", "user_id", in.UserID, "error", err)
		return
	}

	s.publishNotification(s.buildStreamEvent(n))
}

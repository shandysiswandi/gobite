package usecase

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"text/template"
	"time"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/mail"
	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

func (s *Usecase) UserRegistrationNotification(ctx context.Context, msg entity.UserRegistrationMessage) error {
	if msg.UserID <= 0 || msg.Email == "" {
		return nil
	}

	s.createWelcomeNotification(ctx, msg)
	s.createAndSendEmailVerify(ctx, msg)

	return nil
}

func (s *Usecase) createWelcomeNotification(ctx context.Context, msg entity.UserRegistrationMessage) {
	tpl, err := s.repoDB.NotificationTemplateGetByTrigger(ctx, entity.TriggerUserWelcome, entity.ChannelInApp)
	if errors.Is(err, goerror.ErrNotFound) {
		return
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get welcome template", "error", err)
		return
	}

	n := entity.NotificationCreate{
		ID:         s.uid.Generate(),
		UserID:     msg.UserID,
		CategoryID: tpl.CategoryID,
		TriggerKey: tpl.TriggerKey,
		Data:       valueobject.JSONMap{"full_name": msg.FullName},
		Metadata:   valueobject.JSONMap{},
	}

	if err := s.repoDB.NotificationCreate(ctx, n); err != nil {
		slog.ErrorContext(ctx, "failed to repo create welcome notification", "user_id", msg.UserID, "error", err)
	}
}

func (s *Usecase) createAndSendEmailVerify(ctx context.Context, msg entity.UserRegistrationMessage) {
	subjectTpl, bodyTpl, categoryID := s.getEmailVerifyTemplate(ctx)

	token, err := s.jwt.Generate() // need to be clean here
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate email verify token", "user_id", msg.UserID, "error", err)
		return
	}

	data := map[string]any{
		"user_id":   msg.UserID,
		"email":     msg.Email,
		"full_name": msg.FullName,
		"token":     token,
	}

	subject, err := renderTextTemplate(subjectTpl, data)
	if err != nil {
		slog.ErrorContext(ctx, "failed to render email verify subject", "user_id", msg.UserID, "error", err)
		return
	}

	body, err := renderTextTemplate(bodyTpl, data)
	if err != nil {
		slog.ErrorContext(ctx, "failed to render email verify body", "user_id", msg.UserID, "error", err)
		return
	}

	notificationID := s.uid.Generate()
	n := entity.NotificationCreate{
		ID:         notificationID,
		UserID:     msg.UserID,
		CategoryID: categoryID,
		TriggerKey: entity.TriggerEmailVerify,
		Data:       valueobject.JSONMap{"full_name": msg.FullName, "token": token},
		Metadata:   valueobject.JSONMap{"email": msg.Email},
	}

	logID, err := s.repoDB.NotificationCreateWithDeliveryLog(ctx, n, entity.DeliveryLogCreate{
		NotificationID: notificationID,
		Channel:        entity.ChannelEmail,
		Status:         entity.StatusQueued,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo create email verify notification+log", "user_id", msg.UserID, "error", err)
		return
	}

	sendErr := s.repoMail.Send(ctx, mail.Message{
		To:       []string{msg.Email},
		Subject:  subject,
		TextBody: body,
	})

	if sendErr == nil {
		if err := s.repoDB.NotificationDeliveryLogUpdateStatus(ctx, logID, entity.StatusSent, valueobject.JSONMap{"sent": true}, nil); err != nil {
			slog.ErrorContext(ctx, "failed to repo update delivery log status sent", "log_id", logID, "error", err)
		}
		return
	}

	nextRetry := s.clock.Now().Add(5 * time.Minute)
	if err := s.repoDB.NotificationDeliveryLogUpdateStatus(ctx, logID, entity.StatusFailed, valueobject.JSONMap{"error": sendErr.Error()}, &nextRetry); err != nil {
		slog.ErrorContext(ctx, "failed to repo update delivery log status failed", "log_id", logID, "error", err)
	}
	slog.ErrorContext(ctx, "failed to send email verify", "log_id", logID, "user_id", msg.UserID, "error", sendErr)
}

func (s *Usecase) getEmailVerifyTemplate(ctx context.Context) (subjectTpl, bodyTpl string, categoryID int64) {
	tpl, err := s.repoDB.NotificationTemplateGetByTrigger(ctx, entity.TriggerEmailVerify, entity.ChannelEmail)
	if err == nil {
		return tpl.Subject, tpl.Body, tpl.CategoryID
	}

	if !errors.Is(err, goerror.ErrNotFound) {
		slog.ErrorContext(ctx, "failed to repo get email verify template", "error", err)
	}

	return "Verify your email", "Hi {{.full_name}},\n\nUse this token to verify your email:\n\n{{.token}}\n", 1
}

func renderTextTemplate(tpl string, data map[string]any) (string, error) {
	t, err := template.New("text").Option("missingkey=zero").Parse(tpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

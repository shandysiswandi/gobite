package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/mail"
	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

type emailNotificationInput struct {
	UserID           int64
	Email            string
	TriggerKey       entity.TriggerKey
	TemplateData     map[string]any
	NotificationData valueobject.JSONMap
}

func (s *Usecase) sendEmailNotification(ctx context.Context, in emailNotificationInput) {
	tpl := s.getTemplate(ctx, in.TriggerKey, entity.ChannelEmail)
	if tpl == nil {
		return
	}

	body, err := s.renderTemplate("body", tpl.Body, in.TemplateData)
	if err != nil {
		slog.ErrorContext(ctx, "failed to render email body", "user_id", in.UserID, "trigger_key", in.TriggerKey.String(), "error", err)
		return
	}

	n := entity.CreateNotification{
		ID:         s.uid.Generate(),
		UserID:     in.UserID,
		CategoryID: tpl.CategoryID,
		TriggerKey: in.TriggerKey,
		Data:       in.NotificationData,
		Metadata:   valueobject.JSONMap{},
	}

	dl := entity.CreateDeliveryLog{
		NotificationID: n.ID,
		Channel:        entity.ChannelEmail,
		Status:         entity.DeliveryStatusQueued,
	}

	logID, err := s.repoDB.CreateNotificationWithDeliveryLog(ctx, n, dl)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo create email notification+log", "user_id", in.UserID, "trigger_key", in.TriggerKey.String(), "error", err)
		return
	}

	mailErr := s.repoMail.Send(ctx, mail.Message{
		To:       []string{in.Email},
		Subject:  tpl.Subject,
		HTMLBody: body,
	})
	if mailErr == nil {
		up := entity.UpdateDeliveryLog{
			ID:               logID,
			Status:           entity.DeliveryStatusSent,
			ProviderResponse: valueobject.JSONMap{},
		}
		if err := s.repoDB.UpdateDeliveryLogStatus(ctx, up); err != nil {
			slog.ErrorContext(ctx, "failed to repo update delivery log status sent", "log_id", logID, "error", err)
		}
		return
	}

	nextRetry := s.clock.Now().Add(2 * time.Minute) // later will use from config
	up := entity.UpdateDeliveryLog{
		ID:               logID,
		Status:           entity.DeliveryStatusFailed,
		ProviderResponse: valueobject.JSONMap{"error": mailErr.Error()},
		NextRetryAt:      &nextRetry,
	}
	if err := s.repoDB.UpdateDeliveryLogStatus(ctx, up); err != nil {
		slog.ErrorContext(ctx, "failed to repo update delivery log status failed", "log_id", logID, "error", err)
	}

	slog.ErrorContext(ctx, "failed to send notification email", "log_id", logID, "user_id", in.UserID, "trigger_key", in.TriggerKey.String(), "error", mailErr)
}

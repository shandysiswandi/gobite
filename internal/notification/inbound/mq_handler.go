package inbound

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/notification/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/instrument"
	"github.com/shandysiswandi/gobite/internal/pkg/messaging"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/shared/event"
)

const keyOfCorrelationID string = "cID"

type MQHandler struct {
	uc   uc
	uuid uid.StringID
	ins  instrument.Instrumentation
}

func (h *MQHandler) ensureCorrelationID(ctx context.Context, headers []messaging.Header) context.Context {
	for i := range headers {
		if headers[i].Key == keyOfCorrelationID {
			return instrument.SetCorrelationID(ctx, string(headers[i].Value))
		}
	}
	return instrument.SetCorrelationID(ctx, h.uuid.Generate())
}

func (h *MQHandler) UserRegistrationNotification(ctx context.Context, msg messaging.Message) error {
	ctx = h.ensureCorrelationID(ctx, msg.Headers())

	ctx, span := h.ins.Tracer("notification.inbound.mq").Start(ctx, "UserRegistrationNotification")
	defer span.End()

	body := msg.Body()
	slog.InfoContext(ctx, "consume: user registration notification", "msg_body", string(body))

	var payload event.UserRegistrationMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		slog.ErrorContext(ctx, "failed to parse message body of user registration notification", "msg_body", string(body), "error", err)
		return nil
	}

	if err := h.uc.ConsumeUserRegistration(ctx, usecase.ConsumeUserRegistrationInput{
		UserID:   payload.UserID,
		Email:    payload.Email,
		FullName: payload.FullName,
		Token:    payload.ChallengeToken,
	}); err != nil {
		slog.ErrorContext(ctx, "failed to consume user registration", "msg_body", string(body), "error", err)
		return err
	}

	return nil
}

func (h *MQHandler) UserForgotPasswordNotification(ctx context.Context, msg messaging.Message) error {
	ctx = h.ensureCorrelationID(ctx, msg.Headers())

	ctx, span := h.ins.Tracer("notification.inbound.mq").Start(ctx, "UserForgotPasswordNotification")
	defer span.End()

	body := msg.Body()
	slog.InfoContext(ctx, "consume: user forgot password notification", "msg_body", string(body))

	var payload event.UserForgotPasswordMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		slog.ErrorContext(ctx, "failed to parse message body of user forgot password notification", "msg_body", string(body), "error", err)
		return nil
	}

	if err := h.uc.ConsumeUserForgotPassword(ctx, usecase.ConsumeUserForgotPasswordInput{
		UserID: payload.UserID,
		Email:  payload.Email,
		Token:  payload.ChallengeToken,
	}); err != nil {
		slog.ErrorContext(ctx, "failed to consume user forgot password", "msg_body", string(body), "error", err)
		return err
	}

	return nil
}

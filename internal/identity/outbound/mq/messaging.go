package mq

import (
	"context"
	"encoding/json"

	"github.com/shandysiswandi/gobite/internal/identity/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/instrument"
	"github.com/shandysiswandi/gobite/internal/pkg/messaging"
	"github.com/shandysiswandi/gobite/internal/shared/event"
	"go.opentelemetry.io/otel/codes"
)

const keyOfCorrelationID string = "cID"

type Messaging struct {
	client messaging.Messaging
	ins    instrument.Instrumentation
}

func NewMessaging(client messaging.Messaging, ins instrument.Instrumentation) *Messaging {
	return &Messaging{client: client, ins: ins}
}

func (m *Messaging) PublishUserRegistration(ctx context.Context, msg usecase.UserRegistrationEvent) error {
	ctx, span := m.ins.Tracer("identity.outbound.mq").Start(ctx, "PublishUserRegistration")
	defer span.End()

	body, err := json.Marshal(event.UserRegistrationMessage{
		UserID:         msg.UserID,
		Email:          msg.Email,
		FullName:       msg.FullName,
		ChallengeToken: msg.ChallengeToken,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	cID := instrument.GetCorrelationID(ctx)
	if _, err := m.client.Publish(ctx, event.UserRegistrationDestination, messaging.OutgoingMessage{
		Body:    body,
		Headers: []messaging.Header{{Key: keyOfCorrelationID, Value: []byte(cID)}},
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (m *Messaging) PublishUserForgotPassword(ctx context.Context, msg usecase.UserForgotPasswordEvent) error {
	ctx, span := m.ins.Tracer("identity.outbound.mq").Start(ctx, "PublishUserForgotPassword")
	defer span.End()

	body, err := json.Marshal(event.UserForgotPasswordMessage{
		UserID:         msg.UserID,
		Email:          msg.Email,
		ChallengeToken: msg.ChallengeToken,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	cID := instrument.GetCorrelationID(ctx)
	if _, err := m.client.Publish(ctx, event.UserForgotPasswordDestination, messaging.OutgoingMessage{
		Body:    body,
		Headers: []messaging.Header{{Key: keyOfCorrelationID, Value: []byte(cID)}},
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

package mq

import (
	"context"
	"encoding/json"

	"github.com/shandysiswandi/gobite/internal/auth/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/messaging"
)

type Messaging struct {
	client messaging.Messaging
}

func NewMessaging(client messaging.Messaging) *Messaging {
	return &Messaging{client: client}
}

func (m *Messaging) PublishUserRegistration(ctx context.Context, msg entity.UserRegistrationMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = m.client.Publish(ctx, entity.UserRegistrationDestination, messaging.OutgoingMessage{Body: body})
	return err
}

func (m *Messaging) PublishUserForgotPassword(ctx context.Context, msg entity.UserForgotPasswordMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = m.client.Publish(ctx, entity.UserForgotPasswordDestination, messaging.OutgoingMessage{Body: body})
	return err
}

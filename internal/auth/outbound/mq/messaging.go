package mq

import (
	"context"
	"encoding/json"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmessaging"
)

type Messaging struct {
	client pkgmessaging.Messaging
}

func NewMessaging(client pkgmessaging.Messaging) *Messaging {
	return &Messaging{client: client}
}

func (m *Messaging) PublishUserRegistration(ctx context.Context, msg domain.UserRegistrationMessage) error {

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = m.client.Publish(ctx, domain.UserRegistrationDestination, pkgmessaging.OutgoingMessage{Body: body})
	return err
}

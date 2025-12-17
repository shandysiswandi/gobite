package mq

import (
	"context"
	"encoding/json"

	"github.com/shandysiswandi/gobite/internal/auth/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmessaging"
)

type Messaging struct {
	client pkgmessaging.Messaging
}

func NewMessaging(client pkgmessaging.Messaging) *Messaging {
	return &Messaging{client: client}
}

func (m *Messaging) PublishUserRegistration(ctx context.Context, msg entity.UserRegistrationMessage) error {

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = m.client.Publish(ctx, entity.UserRegistrationDestination, pkgmessaging.OutgoingMessage{Body: body})
	return err
}

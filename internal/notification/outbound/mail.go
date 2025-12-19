package outbound

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/pkg/mail"
)

type Mail struct {
	client mail.Mail
}

func NewMail(client mail.Mail) *Mail {
	return &Mail{client: client}
}

func (m *Mail) Send(ctx context.Context, msg mail.Message) error {
	return m.client.Send(ctx, msg)
}

package outbound

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgmail"
)

type Mail struct {
	client pkgmail.Mail
}

func NewMail(client pkgmail.Mail) *Mail {
	return &Mail{client: client}
}

func (m *Mail) Send(ctx context.Context, msg pkgmail.Message) error {
	return m.client.Send(ctx, msg)
}

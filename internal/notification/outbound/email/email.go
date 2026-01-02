package email

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/pkg/instrument"
	"github.com/shandysiswandi/gobite/internal/pkg/mail"
	"go.opentelemetry.io/otel/codes"
)

type Mail struct {
	client mail.Mail
	ins    instrument.Instrumentation
}

func New(client mail.Mail, ins instrument.Instrumentation) *Mail {
	return &Mail{client: client, ins: ins}
}

func (m *Mail) Send(ctx context.Context, msg mail.Message) error {
	ctx, span := m.ins.Tracer("notification.outbound.email").Start(ctx, "Send")
	defer span.End()

	if err := m.client.Send(ctx, msg); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

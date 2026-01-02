package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

const (
	correlationID = "_correlation_id"
	traceID       = "_trace_id"
	spanID        = "_span_id"
)

type contextHandler struct {
	slog.Handler
	serviceName      string
	getCorrelationID func(context.Context) string
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.getCorrelationID != nil {
		if cID := h.getCorrelationID(ctx); cID != "" && cID != "[invalid_chain_id]" {
			r.AddAttrs(slog.String(correlationID, cID))
		}
	}

	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(
			slog.String(traceID, sc.TraceID().String()),
			slog.String(spanID, sc.SpanID().String()),
		)
	}

	r.AddAttrs(slog.String("service", h.serviceName))

	return h.Handler.Handle(ctx, r)
}

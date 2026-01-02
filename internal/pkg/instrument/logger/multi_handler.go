package logger

import (
	"context"
	"log/slog"
)

type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range m.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	var firstErr error
	for _, handler := range m.handlers {
		if !handler.Enabled(ctx, record.Level) {
			continue
		}
		rec := record.Clone()
		if err := handler.Handle(ctx, rec); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(m.handlers))
	for _, handler := range m.handlers {
		handlers = append(handlers, handler.WithAttrs(attrs))
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(m.handlers))
	for _, handler := range m.handlers {
		handlers = append(handlers, handler.WithGroup(name))
	}
	return &multiHandler{handlers: handlers}
}

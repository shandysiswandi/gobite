package logger

import (
	"context"
	"log/slog"
	"strings"
)

type maskHandler struct {
	handler  slog.Handler
	maskKeys map[string]struct{}
}

func (h *maskHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *maskHandler) Handle(ctx context.Context, record slog.Record) error {
	if len(h.maskKeys) == 0 {
		return h.handler.Handle(ctx, record)
	}

	masked := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		masked.AddAttrs(maskAttr(attr, h.maskKeys))
		return true
	})

	return h.handler.Handle(ctx, masked)
}

func (h *maskHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &maskHandler{
		handler:  h.handler.WithAttrs(attrs),
		maskKeys: h.maskKeys,
	}
}

func (h *maskHandler) WithGroup(name string) slog.Handler {
	return &maskHandler{
		handler:  h.handler.WithGroup(name),
		maskKeys: h.maskKeys,
	}
}

func buildMaskKeys(fields []string) map[string]struct{} {
	maskKeys := make(map[string]struct{})
	for _, field := range fields {
		field = strings.TrimSpace(strings.ToLower(field))
		if field == "" {
			continue
		}
		maskKeys[field] = struct{}{}
	}
	return maskKeys
}

func maskAttr(attr slog.Attr, maskKeys map[string]struct{}) slog.Attr {
	if _, found := maskKeys[strings.ToLower(attr.Key)]; found {
		return slog.String(attr.Key, "***")
	}

	switch attr.Value.Kind() {
	case slog.KindGroup:
		group := attr.Value.Group()
		masked := make([]slog.Attr, 0, len(group))
		for _, ga := range group {
			masked = append(masked, maskAttr(ga, maskKeys))
		}
		attr.Value = slog.GroupValue(masked...)
	case slog.KindString:
		if masked, ok := maskJSONString(attr.Value.String(), maskKeys); ok {
			attr.Value = slog.StringValue(masked)
		}
	case slog.KindAny:
		val := attr.Value.Any()
		if val == nil {
			return attr
		}
		if masked, ok := maskAny(val, maskKeys); ok {
			attr.Value = slog.AnyValue(masked)
			return attr
		}
		if b, ok := val.([]byte); ok {
			if masked, ok := maskJSONBytes(b, maskKeys); ok {
				attr.Value = slog.StringValue(masked)
			}
		}
	}

	return attr
}

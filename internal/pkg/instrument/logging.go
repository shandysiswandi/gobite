package instrument

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func initLogging(serviceName string, lp *sdklog.LoggerProvider, maskFields []string) {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Key = "ts"
			case slog.LevelKey:
				a.Key = "severity"
			case slog.SourceKey:
				if src, ok := a.Value.Any().(*slog.Source); ok {
					if strings.Contains(src.File, "/internal/") {
						relPath := filepath.Join("internal", strings.SplitAfter(src.File, "/internal/")[1])
						return slog.Attr{
							Key:   "file",
							Value: slog.StringValue(fmt.Sprintf("%s:%d", relPath, src.Line)),
						}
					}
					return slog.Attr{}
				}
			}
			return a
		},
	})

	handlers := []slog.Handler{jsonHandler}
	if lp != nil {
		handlers = append(handlers, otelslog.NewHandler(
			serviceName,
			otelslog.WithLoggerProvider(lp),
		))
	}

	var handler slog.Handler
	if len(handlers) == 1 {
		handler = handlers[0]
	} else {
		handler = &multiHandler{handlers: handlers}
	}

	slog.SetDefault(slog.New(&contextHandler{
		Handler:     &maskHandler{handler: handler, maskKeys: buildMaskKeys(maskFields)},
		serviceName: serviceName,
	}))
}

type contextHandler struct {
	slog.Handler
	serviceName string
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if cID := GetCorrelationID(ctx); cID != "" && cID != "[invalid_chain_id]" {
		r.AddAttrs(slog.String("_cID", cID))
	}
	r.AddAttrs(slog.String("service", h.serviceName))

	return h.Handler.Handle(ctx, r)
}

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

func maskAny(val any, maskKeys map[string]struct{}) (any, bool) {
	switch v := val.(type) {
	case map[string]any:
		return maskData(v, maskKeys), true
	case map[string]string:
		converted := make(map[string]any, len(v))
		for k, v2 := range v {
			converted[k] = v2
		}
		return maskData(converted, maskKeys), true
	case []any:
		return maskData(v, maskKeys), true
	default:
		return nil, false
	}
}

func maskJSONString(payload string, maskKeys map[string]struct{}) (string, bool) {
	if payload == "" || (payload[0] != '{' && payload[0] != '[') {
		return "", false
	}
	var jsonBody any
	if err := json.Unmarshal([]byte(payload), &jsonBody); err != nil {
		return "", false
	}
	masked := maskData(jsonBody, maskKeys)
	if maskedBytes, err := json.Marshal(masked); err == nil {
		return string(maskedBytes), true
	}
	return "", false
}

func maskJSONBytes(payload []byte, maskKeys map[string]struct{}) (string, bool) {
	if len(payload) == 0 {
		return "", false
	}
	var jsonBody any
	if err := json.Unmarshal(payload, &jsonBody); err != nil {
		return "", false
	}
	masked := maskData(jsonBody, maskKeys)
	if maskedBytes, err := json.Marshal(masked); err == nil {
		return string(maskedBytes), true
	}
	return "", false
}

func maskData(v any, maskKeys map[string]struct{}) any {
	switch val := v.(type) {
	case map[string]any:
		masked := make(map[string]any, len(val))
		for k, v2 := range val {
			if _, found := maskKeys[strings.ToLower(k)]; found {
				masked[k] = "***"
			} else {
				masked[k] = maskData(v2, maskKeys)
			}
		}
		return masked
	case []any:
		res := make([]any, len(val))
		for i, v2 := range val {
			res[i] = maskData(v2, maskKeys)
		}
		return res
	default:
		return v
	}
}

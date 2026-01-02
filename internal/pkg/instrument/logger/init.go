package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func Init(serviceName string, lp *sdklog.LoggerProvider, maskFields []string, getCorrelationID func(context.Context) string) {
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
					return slog.Attr{
						Key:   "file",
						Value: slog.StringValue(fmt.Sprintf("%s:%d", filepath.Base(src.File), src.Line)),
					}
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
		Handler:          &maskHandler{handler: handler, maskKeys: buildMaskKeys(maskFields)},
		serviceName:      serviceName,
		getCorrelationID: getCorrelationID,
	}))
}

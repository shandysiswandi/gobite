package pkglog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgcid"
)

func InitLogging() {
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

	slog.SetDefault(slog.New(&contextHandler{Handler: jsonHandler}))

}

type contextHandler struct {
	slog.Handler
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if cID := pkgcid.GetCorrelationID(ctx); cID != "" && cID != "[invalid_chain_id]" {
		r.AddAttrs(slog.String("_cID", cID))
	}
	r.AddAttrs(slog.String("service", "gobite"))

	return h.Handler.Handle(ctx, r)
}

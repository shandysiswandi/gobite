package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/shandysiswandi/gobite/internal/pkg/stacktrace"
)

func callHandlerWithRecover(ctx context.Context, kind string, fn func() error) (err error) {
	defer func() {
		if rvr := recover(); rvr != nil {
			stack := debug.Stack()
			paths := stacktrace.InternalPaths(stack)
			if len(paths) == 0 {
				slog.ErrorContext(ctx, "panic in messaging handler", "kind", kind, "panic", rvr, "stack", string(stack))
			} else {
				slog.ErrorContext(ctx, "panic in messaging handler", "kind", kind, "panic", rvr, "stack", paths)
			}
			err = fmt.Errorf("pkgmessage: panic in %s handler: %v", kind, rvr)
		}
	}()

	return fn()
}

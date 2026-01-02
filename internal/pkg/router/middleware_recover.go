package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/shandysiswandi/gobite/internal/pkg/stacktrace"
)

//nolint:errcheck,gosec,contextcheck // ignore error
func middlewareRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				//nolint:err113,errorlint // this must compare directly
				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}

				w.Header().Set("Content-Type", "application/json; charset=utf-8")

				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}

				paths := stacktrace.InternalPaths(debug.Stack())
				if len(paths) == 0 {
					slog.ErrorContext(r.Context(), "panic on the server trace debug", "because", rvr, "stack", string(debug.Stack()))
				} else {
					slog.ErrorContext(r.Context(), "panic on the server", "because", rvr, "stack", paths)
				}

				json.NewEncoder(w).Encode(map[string]string{
					"message": "Internal server error",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

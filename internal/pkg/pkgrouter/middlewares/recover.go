package middlewares

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				//nolint:err113,errorlint // this must compare directly
				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}

				slog.ErrorContext(r.Context(), "panic on the server", "because", rvr)

				w.Header().Set("Content-Type", "application/json; charset=utf-8")

				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}

				// Filtered stack trace for internal files only
				lines := strings.Split(string(debug.Stack()), "\n")
				printStackTrace(lines)

				// Send a default fallback response to the client.
				//nolint:errcheck,gosec // ignore error
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Internal server error",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func printStackTrace(lines []string) {
	fmt.Fprintln(os.Stderr, "===== ===== START ===== =====")
	for i := 0; i < len(lines)-1; i++ {
		line := strings.TrimSpace(lines[i+1])
		if strings.Contains(line, "/internal/") && strings.Contains(line, ".go") {
			if idx := strings.Index(line, ".go:"); idx != -1 {
				// Cut after ".go:xxx"
				end := strings.Index(line[idx:], " ") // find space after ":line"
				if end == -1 {
					end = len(line)
				} else {
					end += idx
				}
				shortPath := line[:end]
				// Trim full absolute path to just the internal path
				internalIdx := strings.Index(shortPath, "/internal/")
				if internalIdx != -1 {
					shortPath = shortPath[internalIdx+1:] // remove leading "/"
					fmt.Fprintln(os.Stderr, "stack trace: ", shortPath)
				}
			}
		}
	}
	fmt.Fprintln(os.Stderr, "===== ===== END ===== =====")
}

package pkgrouter

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
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

				slog.ErrorContext(r.Context(), "panic on the server", "because", rvr)

				w.Header().Set("Content-Type", "application/json; charset=utf-8")

				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}

				lines := strings.Split(string(debug.Stack()), "\n")
				printStackTrace(lines)

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
				end := strings.Index(line[idx:], " ")
				if end == -1 {
					end = len(line)
				} else {
					end += idx
				}
				shortPath := line[:end]
				internalIdx := strings.Index(shortPath, "/internal/")
				if internalIdx != -1 {
					shortPath = shortPath[internalIdx+1:]
					fmt.Fprintln(os.Stderr, "stack trace: ", shortPath)
				}
			}
		}
	}
	fmt.Fprintln(os.Stderr, "===== ===== END ===== =====")
}

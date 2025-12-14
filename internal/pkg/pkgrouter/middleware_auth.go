package pkgrouter

import (
	"net/http"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

//nolint:gochecknoglobals // global for fast reuse
var skipEndpoints = map[string]map[string]struct{}{
	http.MethodPost: {
		"/auth/login":           struct{}{},
		"/auth/login-2fa":       struct{}{},
		"/auth/register":        struct{}{},
		"/auth/password/forgot": struct{}{},
		"/auth/password/reset":  struct{}{},
		"/auth/refresh-token":   struct{}{},
	},
	http.MethodGet: {
		"/":       struct{}{},
		"/health": struct{}{},
	},
}

func middlewareAuthentication(verifier pkgjwt.JWT[pkgjwt.AccessTokenPayload]) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			if s, ok := skipEndpoints[r.Method]; ok {
				if _, skip := s[path]; skip {
					next.ServeHTTP(w, r)
					return
				}
			}

			p := strings.Fields(r.Header.Get("Authorization"))
			if len(p) != 2 || !strings.EqualFold(p[0], "Bearer") {
				writeJSON(w, map[string]string{"message": "authentication required"}, http.StatusUnauthorized)
				return
			}

			claims, err := verifier.Verify(p[1])
			if err != nil {
				writeJSON(w, map[string]string{"message": "invalid or expired token"}, http.StatusUnauthorized)
				return
			}

			ctx := pkgjwt.SetAuth(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

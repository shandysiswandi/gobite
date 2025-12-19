package router

import (
	"net/http"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

//nolint:gochecknoglobals // global for fast reuse
var skipEndpoints = map[string]map[string]struct{}{
	http.MethodPost: {
		"/api/v1/auth/logout":          struct{}{},
		"/api/v1/auth/login":           struct{}{},
		"/api/v1/auth/login/mfa":       struct{}{},
		"/api/v1/auth/register":        struct{}{},
		"/api/v1/auth/register/resend": struct{}{},
		"/api/v1/auth/refresh":         struct{}{},
		"/api/v1/auth/email/verify":    struct{}{},
		"/api/v1/auth/password/forgot": struct{}{},
		"/api/v1/auth/password/reset":  struct{}{},
	},
	http.MethodGet: {
		"/":       struct{}{},
		"/health": struct{}{},
	},
}

func middlewareAuthentication(_ config.Config, verifier jwt.JWT) func(next http.Handler) http.Handler {
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

			ctx := jwt.SetAuth(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

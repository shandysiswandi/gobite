package pkgrouter

import (
	"net/http"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

//nolint:gochecknoglobals // global for fast reuse
var skipEndpoints = map[string]map[string]struct{}{
	"POST": {
		"/auth/login":         struct{}{},
		"/auth/register":      struct{}{},
		"/auth/refresh-token": struct{}{},
	},
}

func Authentication(uid pkgjwt.JWT[pkgjwt.AccessTokenPayload]) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			if s, ok := skipEndpoints[r.Method]; ok {
				if _, skip := s[path]; skip {
					next.ServeHTTP(w, r)
					return
				}
			}

			token := parseAuthHeader(r.Header.Get("Authorization"))
			if token == "" {
				ResponseError(w, pkgerror.ErrUnauthenticated)
				return
			}

			claims, err := uid.Verify(token)
			if err != nil {
				ResponseError(w, pkgerror.ErrUnauthenticated)
				return
			}

			ctx := pkgjwt.SetAuth(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func parseAuthHeader(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

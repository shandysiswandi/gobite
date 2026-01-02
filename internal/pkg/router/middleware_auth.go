package router

import (
	"net/http"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

func middlewareAuthentication(verifier jwt.JWT, publicEndpoints map[string]map[string]struct{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := matchedRoutePath(r)

			if s, ok := publicEndpoints[r.Method]; ok {
				if _, skip := s[path]; skip {
					next.ServeHTTP(w, r)
					return
				}
			}

			p := strings.Fields(r.Header.Get("Authorization"))
			if len(p) != 2 || !strings.EqualFold(p[0], "Bearer") {
				writeJSON(w, map[string]string{"message": "Authentication required"}, http.StatusUnauthorized)
				return
			}

			claims, err := verifier.Verify(p[1])
			if err != nil {
				writeJSON(w, map[string]string{"message": "Invalid or expired token"}, http.StatusUnauthorized)
				return
			}

			ctx := jwt.SetAuth(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

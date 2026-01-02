package router

import (
	"net/http"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/config"
)

func middlewareMaintenance(cfg config.Config) Middleware {
	endpoints := make(map[string]struct{})
	if cfg != nil {
		for _, endpoint := range cfg.GetArray("app.maintenance.endpoints") {
			endpoint = strings.TrimSpace(endpoint)
			if endpoint == "" {
				continue
			}
			endpoints[endpoint] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := matchedRoutePath(r)
			if _, blocked := endpoints[route]; blocked {
				writeJSON(w, errorResponse{Message: "service is under maintenance"}, http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

package pkgrouter

import (
	"net/http"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/pkglog"
)

// Generator generates a unique string (used for correlation/request IDs).
type Generator interface {
	Generate() string
}

const (
	// HeaderCorrelationID is the canonical header used to track requests end-to-end.
	HeaderCorrelationID = "X-Correlation-ID"
	// HeaderRequestID is an accepted alternative header name used by some proxies.
	HeaderRequestID = "X-Request-ID"
)

func normalizeCID(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.ContainsAny(v, "\r\n") {
		return ""
	}
	const maxLen = 128
	if len(v) > maxLen {
		v = v[:maxLen]
	}
	return v
}

func middlewareCorrelationID(uid Generator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cid := normalizeCID(r.Header.Get(HeaderCorrelationID))
			if cid == "" {
				cid = normalizeCID(r.Header.Get(HeaderRequestID))
			}
			if cid == "" && uid != nil {
				cid = uid.Generate()
			}

			if cid != "" {
				w.Header().Set(HeaderCorrelationID, cid)
				r = r.WithContext(pkglog.SetCorrelationID(r.Context(), cid))
			}

			next.ServeHTTP(w, r)
		})
	}
}

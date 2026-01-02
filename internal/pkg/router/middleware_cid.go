package router

import (
	"net/http"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/instrument"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
)

const (
	// HeaderCorrelationID is the canonical header used to track requests end-to-end.
	HeaderCorrelationID = "X-Correlation-ID"
	// HeaderRequestID is an accepted alternative header name used by some proxies.
	HeaderRequestID = "X-Request-ID"
)

func normalizeCID(v string) string {
	if strings.ContainsAny(v, "\r\n") {
		return ""
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	const maxLen = 128
	if len(v) > maxLen {
		v = v[:maxLen]
	}
	return v
}

func middlewareCorrelationID(uid uid.StringID) Middleware {
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
				r = r.WithContext(instrument.SetCorrelationID(r.Context(), cid))
			}

			next.ServeHTTP(w, r)
		})
	}
}

package pkgrouter

import (
	"net/http"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/pkglog"
)

type Generator interface {
	Generate() string
}

const (
	HeaderCorrelationID = "X-Correlation-ID"
	HeaderRequestID     = "X-Request-ID"
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

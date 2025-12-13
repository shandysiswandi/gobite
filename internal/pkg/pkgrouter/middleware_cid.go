package pkgrouter

import (
	"net/http"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgcid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
)

func CorrelationID(uid pkguid.StringID) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(pkgcid.SetCorrelationID(r.Context(), uid.Generate())))
		})
	}
}

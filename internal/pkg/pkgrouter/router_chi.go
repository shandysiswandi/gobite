package pkgrouter

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
)

func NewChi(uid pkguid.StringID, jwtAccess pkgjwt.JWT[pkgjwt.AccessTokenPayload]) chi.Router {
	r := chi.NewRouter()
	r.Use(Authentication(jwtAccess))
	r.Use(Logging)
	r.Use(CorrelationID(uid))
	r.Use(Recoverer)

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		Response(w, map[string]string{"message": "hi from gobite"})
	})

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		Response(w, map[string]string{"message": "server is running well"})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		ResponseError(w, pkgerror.ErrNotFound)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		ResponseError(w, pkgerror.ErrMethodNotAllowed)
	})

	return r
}

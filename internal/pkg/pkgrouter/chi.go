package pkgrouter

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgrouter/middlewares"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
)

func NewChi(uid pkguid.StringID) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.CorrelationID(uid))
	r.Use(middlewares.Logging)
	r.Use(middlewares.Recoverer)

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{"message": "hi from gobite"}, http.StatusOK)
	})

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{"message": "server is running well"}, http.StatusOK)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, errorResponse{Message: "endpoint not found"}, http.StatusNotFound)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, errorResponse{Message: "method not allowed"}, http.StatusMethodNotAllowed)
	})

	return r
}

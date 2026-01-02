package router

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/casbin/casbin/v3"
	"github.com/julienschmidt/httprouter"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/instrument"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/pkg/validator"
)

type errorResponse struct {
	Message string            `json:"message" example:"example string message"`
	Error   map[string]string `json:"error,omitempty"`
}

type successResponse struct {
	Message string         `json:"message" example:"example string message"`
	Data    any            `json:"data" swaggertype:"object"`
	Meta    map[string]any `json:"meta,omitempty" swaggertype:"object"`
}

// Handler is the application-style handler used by this router.
//
// It returns a response payload (that will be JSON encoded) or an error.
type Handler func(r *Request) (any, error)

// Config holds dependencies required to build a Router.
type Config struct {
	// Config provides runtime configuration values.
	Config config.Config
	// UUID generates request correlation IDs.
	UUID uid.StringID
	// JWT validates and parses authentication tokens.
	JWT jwt.JWT
	// Instrument provides tracing and metrics helpers.
	Instrument instrument.Instrumentation
	// Enforcer applies authorization policies.
	Enforcer *casbin.Enforcer
}

// Router is an http.Handler that wraps httprouter and a middleware chain.
type Router struct {
	hr         *httprouter.Router
	errorCodec func(ctx context.Context, w http.ResponseWriter, err error)
	encoder    func(ctx context.Context, w http.ResponseWriter, resp any)
	mws        []Middleware
}

// NewRouter builds the default application router with standard middleware.
func NewRouter(cfg Config) *Router {
	hr := &httprouter.Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
		SaveMatchedRoutePath:   true,
		NotFound: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]string{"message": "endpoint not found"}, http.StatusNotFound)
		}),
		MethodNotAllowed: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]string{"message": "method not allowed"}, http.StatusMethodNotAllowed)
		}),
	}

	hr.GET("/", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		writeJSON(w, map[string]string{"message": "Welcome to API GoBite"}, http.StatusNotFound)
	})

	errorCodec := func(ctx context.Context, w http.ResponseWriter, err error) {
		var gerr *goerror.Error
		if !errors.As(err, &gerr) {
			writeJSON(w, errorResponse{Message: "Internal server error"}, http.StatusInternalServerError)
			return
		}

		errResp := errorResponse{Message: gerr.Msg()}

		var errValidate validator.V10ValidationError
		if errors.As(err, &errValidate) {
			errResp.Error = errValidate.Values()
		} else if len(gerr.Fields()) > 0 {
			errResp.Error = gerr.Fields()
		}

		writeJSON(w, errResp, gerr.StatusCode())
	}

	okCodec := func(ctx context.Context, w http.ResponseWriter, resp any) {
		code := http.StatusOK
		if sc, ok := resp.(interface {
			StatusCode() int
		}); ok {
			code = sc.StatusCode()
		}

		if code == http.StatusNoContent || resp == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		msg := "request has been successfully"
		if m, ok := resp.(interface {
			Message() string
		}); ok {
			msg = m.Message()
		}

		var meta map[string]any
		if m, ok := resp.(interface {
			Meta() map[string]any
		}); ok {
			meta = m.Meta()
		}

		writeJSON(w, successResponse{
			Message: msg,
			Data:    resp,
			Meta:    meta,
		}, code)
	}

	publicEndpoints := map[string]map[string]struct{}{
		http.MethodGet: {
			"/":       {},
			"/health": {},
		},
		http.MethodPost: {
			"/api/v1/identity/login":           {},
			"/api/v1/identity/login/2fa":       {},
			"/api/v1/identity/refresh":         {},
			"/api/v1/identity/register":        {},
			"/api/v1/identity/register/resend": {},
			"/api/v1/identity/register/verify": {},
			"/api/v1/identity/password/forgot": {},
			"/api/v1/identity/password/reset":  {},
		},
	}
	ro := &Router{
		hr:         hr,
		errorCodec: errorCodec,
		encoder:    okCodec,
		mws: []Middleware{
			middlewareRecoverer,
			middlewareIP,
			middlewareCorrelationID(cfg.UUID),
			middlewareObservability(cfg.Config, cfg.Instrument),
			middlewareMaintenance(cfg.Config),
			middlewareAuthentication(cfg.JWT, publicEndpoints),
		},
	}

	return ro
}

// GET registers a GET endpoint using the application Handler signature.
func (r *Router) GET(path string, h Handler, mws ...Middleware) {
	r.endpoint(http.MethodGet, path, h, mws...)
}

// GETRaw registers a GET endpoint that writes directly to the response writer.
func (r *Router) GETRaw(path string, h http.Handler, mws ...Middleware) {
	r.hr.Handler(http.MethodGet, path, Chain(h, append(r.mws, mws...)...))
}

// POST registers a POST endpoint using the application Handler signature.
func (r *Router) POST(path string, h Handler, mws ...Middleware) {
	r.endpoint(http.MethodPost, path, h, mws...)
}

// PUT registers a PUT endpoint using the application Handler signature.
func (r *Router) PUT(path string, h Handler, mws ...Middleware) {
	r.endpoint(http.MethodPut, path, h, mws...)
}

// PATCH registers a PATCH endpoint using the application Handler signature.
func (r *Router) PATCH(path string, h Handler, mws ...Middleware) {
	r.endpoint(http.MethodPatch, path, h, mws...)
}

// DELETE registers a DELETE endpoint using the application Handler signature.
func (r *Router) DELETE(path string, h Handler, mws ...Middleware) {
	r.endpoint(http.MethodDelete, path, h, mws...)
}

func (r *Router) endpoint(method, path string, h Handler, mws ...Middleware) {
	r.hr.Handler(method, path, Chain(http.HandlerFunc(func(w http.ResponseWriter, re *http.Request) {
		resp, err := h(&Request{Request: re})
		if err != nil {
			if setter, ok := w.(interface{ SetError(error) }); ok {
				setter.SetError(err)
			}
			r.errorCodec(re.Context(), w, err)
			return
		}
		r.encoder(re.Context(), w, resp)
	}), append(r.mws, mws...)...))
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.hr.ServeHTTP(w, req)
}

func writeJSON(w http.ResponseWriter, data any, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		slog.Error("server: failed to encode data to json", "error", err)
	}
}

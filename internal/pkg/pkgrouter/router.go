package pkgrouter

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

// Handler is the application-style handler used by this router.
//
// It returns a response payload (that will be JSON encoded) or an error.
type Handler func(ctx context.Context, r *http.Request) (any, error)

// Router is an http.Handler that wraps httprouter and a middleware chain.
type Router struct {
	hr         *httprouter.Router
	errorCodec func(ctx context.Context, w http.ResponseWriter, err error)
	encoder    func(ctx context.Context, w http.ResponseWriter, resp any)
	mws        []Middleware
}

// NewRouter builds the default application router with standard middleware.
func NewRouter(uuid Generator, jwtAccess pkgjwt.JWT[pkgjwt.AccessTokenPayload]) *Router {
	ro := &Router{
		hr: &httprouter.Router{
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
		},
		errorCodec: func(ctx context.Context, w http.ResponseWriter, err error) {
			var gerr *pkgerror.Error
			if !errors.As(err, &gerr) {
				writeJSON(w, errorResponse{Message: "Internal server error"}, http.StatusInternalServerError)
				return
			}

			errResp := errorResponse{Message: gerr.Msg()}

			var errValidate pkgvalidator.V10ValidationError
			if errors.As(err, &errValidate) {
				errResp.Error = errValidate.Values()
			}

			writeJSON(w, errResp, gerr.StatusCode())
		},
		encoder: func(ctx context.Context, w http.ResponseWriter, resp any) {
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

			writeJSON(w, successReponse{
				Message: msg,
				Data:    resp,
				Meta:    meta,
			}, code)
		},
		mws: []Middleware{
			middlewareRecoverer,
			middlewareCorrelationID(uuid),
			middlewareLogging,
			middlewareAuthentication(jwtAccess),
		},
	}

	ro.Handle(http.MethodGet, "/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{"message": "hi from gobite"}, http.StatusOK)
	}))

	ro.Handle(http.MethodGet, "/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{"message": "server is running well"}, http.StatusOK)
	}))

	return ro
}

// Use appends middleware to the existing middleware stack.
func (r *Router) Use(mws ...Middleware) {
	r.mws = append(r.mws, mws...)
}

// GET registers a GET endpoint using the application Handler signature.
func (r *Router) GET(path string, h Handler, mws ...Middleware) {
	r.endpoint(http.MethodGet, path, h, mws...)
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

// Handle registers a raw http.Handler with the router.
func (r *Router) Handle(method, path string, h http.Handler, mws ...Middleware) {
	r.hr.Handler(method, path, Chain(h, append(r.mws, mws...)...))
}

func (r *Router) endpoint(method, path string, h Handler, mws ...Middleware) {
	r.hr.Handler(method, path, Chain(http.HandlerFunc(func(w http.ResponseWriter, re *http.Request) {
		resp, err := h(re.Context(), re)
		if err != nil {
			r.errorCodec(re.Context(), w, err)
			return
		}
		r.encoder(re.Context(), w, resp)
	}), append(r.mws, mws...)...))
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.hr.ServeHTTP(w, req)
}

type errorResponse struct {
	Message string            `json:"message"`
	Error   map[string]string `json:"error,omitempty"`
}

type successReponse struct {
	Message string         `json:"message"`
	Data    any            `json:"data"`
	Meta    map[string]any `json:"meta,omitempty"`
}

func writeJSON(w http.ResponseWriter, data any, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		slog.Error("server: failed to encode data to json", "error", err)
	}
}

// func ResponseError(w http.ResponseWriter, err error) {
// 	if err == nil {
// 		writeJSON(w, errorResponse{Message: "unknown error"}, http.StatusInternalServerError)
// 		return
// 	}

// 	var (
// 		validation    interface{ Values() map[string]string }
// 		jsonSyntaxErr *json.SyntaxError
// 		jsonTypeErr   *json.UnmarshalTypeError
// 	)

// 	switch {
// 	case
// 		errors.As(err, &validation):
// 		writeJSON(w, errorResponse{Message: "validation error", Error: validation.Values()}, http.StatusUnprocessableEntity)

// 	case
// 		errors.As(err, &jsonSyntaxErr),
// 		errors.As(err, &jsonTypeErr),
// 		errors.Is(err, io.ErrUnexpectedEOF):
// 		writeJSON(w, errorResponse{Message: "invalid request body"}, http.StatusBadRequest)

// 	case
// 		errors.Is(err, pkgerror.ErrNotFound):
// 		writeJSON(w, errorResponse{Message: err.Error()}, http.StatusNotFound)

// 	case
// 		errors.Is(err, pkgerror.ErrMethodNotAllowed):
// 		writeJSON(w, errorResponse{Message: err.Error()}, http.StatusMethodNotAllowed)

// 	case
// 		errors.Is(err, pkgerror.ErrUnauthenticated),
// 		errors.Is(err, pkgerror.ErrAuthUnauthenticated):
// 		writeJSON(w, errorResponse{Message: err.Error()}, http.StatusUnauthorized)

// 	case
// 		errors.Is(err, pkgerror.ErrUnauthorized),
// 		errors.Is(err, pkgerror.ErrAuthNotVerified),
// 		errors.Is(err, pkgerror.ErrAuthBanned):
// 		writeJSON(w, errorResponse{Message: err.Error()}, http.StatusForbidden)

// 	case
// 		errors.Is(err, pkgerror.ErrAuthEmailUsed):
// 		writeJSON(w, errorResponse{Message: err.Error()}, http.StatusConflict)

// 	default:
// 		writeJSON(w, errorResponse{Message: err.Error()}, http.StatusInternalServerError)
// 	}
// }

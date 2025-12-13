package pkgrouter

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
)

func writeJSON(w http.ResponseWriter, data any, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		slog.Error("server: failed to encode data to json", "error", err)
	}
}

func Response(w http.ResponseWriter, data any) {
	writeJSON(w, successReponse{Message: "request has been successfully", Data: data}, http.StatusOK)
}

func ResponseError(w http.ResponseWriter, err error) {
	if err == nil {
		writeJSON(w, errorResponse{Message: "unknown error"}, http.StatusInternalServerError)
		return
	}

	var (
		validation    interface{ Values() map[string]string }
		jsonSyntaxErr *json.SyntaxError
		jsonTypeErr   *json.UnmarshalTypeError
	)

	switch {
	case errors.As(err, &validation):
		writeJSON(w, errorResponse{Message: "validation error", Error: validation.Values()}, http.StatusUnprocessableEntity)
	case errors.As(err, &jsonSyntaxErr), errors.As(err, &jsonTypeErr), errors.Is(err, io.ErrUnexpectedEOF):
		writeJSON(w, errorResponse{Message: "invalid request body"}, http.StatusBadRequest)
	case errors.Is(err, pkgerror.ErrNotFound):
		writeJSON(w, errorResponse{Message: err.Error()}, http.StatusNotFound)
	case errors.Is(err, pkgerror.ErrAuthUnauthenticated), errors.Is(err, pkgerror.ErrAuthMfaRequired):
		writeJSON(w, errorResponse{Message: err.Error()}, http.StatusUnauthorized)
	case errors.Is(err, pkgerror.ErrAuthNotVerified), errors.Is(err, pkgerror.ErrAuthBanned):
		writeJSON(w, errorResponse{Message: err.Error()}, http.StatusForbidden)
	default:
		writeJSON(w, errorResponse{Message: err.Error()}, http.StatusInternalServerError)
	}
}

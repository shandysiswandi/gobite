package pkgrouter

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

//nolint:gochecknoglobals // global for fast reuse
var sensitiveKeys = map[string]struct{}{
	"password":      {},
	"access_token":  {},
	"refresh_token": {},
	"authorization": {},
	"cookie":        {},
}

func maskHeaders(headers http.Header) http.Header {
	result := headers.Clone()
	for key := range result {
		if _, found := sensitiveKeys[strings.ToLower(key)]; found {
			result.Set(key, "***")
		}
	}
	return result
}

func maskData(v any) any {
	switch val := v.(type) {
	case map[string]any:
		masked := make(map[string]any, len(val))
		for k, v2 := range val {
			if _, found := sensitiveKeys[strings.ToLower(k)]; found {
				masked[k] = "***"
			} else {
				masked[k] = maskData(v2)
			}
		}
		return masked

	case []any:
		res := make([]any, len(val))
		for i, v2 := range val {
			res[i] = maskData(v2)
		}
		return res

	default:
		return v
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

//nolint:errcheck,gosec // ignore error
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		// --- Read and restore body ---
		var reqBodyBytes []byte
		if r.Body != nil {
			reqBodyBytes, _ = io.ReadAll(r.Body)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))

		var reqJSON any
		json.Unmarshal(reqBodyBytes, &reqJSON)
		maskedReqBody := maskData(reqJSON)

		slog.InfoContext(
			r.Context(),
			"Request received",
			"path", r.URL.Path,
			"method", r.Method,
			"headers", maskHeaders(r.Header),
			"body", maskedReqBody,
		)

		// --- Capture response ---
		rec := &responseRecorder{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			status:         200,
		}

		next.ServeHTTP(rec, r)

		// parse JSON response if possible
		var respJSON any
		json.Unmarshal(rec.body.Bytes(), &respJSON)
		maskedRespBody := maskData(respJSON)

		slog.InfoContext(
			r.Context(),
			"Response sent",
			"path", r.URL.Path,
			"method", r.Method,
			"status", rec.status,
			"latency_ms", time.Since(start).Milliseconds(),
			"body", maskedRespBody,
		)
	})
}

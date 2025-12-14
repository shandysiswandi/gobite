package pkgrouter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/julienschmidt/httprouter"
)

const maxLoggedBodyBytes = 64 * 1024

//nolint:gochecknoglobals // global for fast reuse
var sensitiveKeys = map[string]struct{}{
	"password":         {},
	"new_password":     {},
	"current_password": {},
	"access_token":     {},
	"refresh_token":    {},
	"authorization":    {},
	"cookie":           {},
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

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
	body   *bytes.Buffer
	capped bool
}

func (w *statusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusRecorder) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	if w.body != nil && !w.capped && len(p) > 0 {
		remaining := maxLoggedBodyBytes - w.body.Len()
		if remaining > 0 {
			if len(p) > remaining {
				w.body.Write(p[:remaining])
				w.capped = true
			} else {
				w.body.Write(p)
			}
		} else {
			w.capped = true
		}
	}

	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

func (w *statusRecorder) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

//nolint:err113 // it use dynamic error
func (w *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

func (w *statusRecorder) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func matchedRoutePath(r *http.Request) string {
	pattern := httprouter.ParamsFromContext(r.Context()).MatchedRoutePath()
	if pattern != "" {
		return pattern
	}
	return r.URL.Path
}

func parseAndMaskBody(contentType string, body []byte) any {
	if len(body) == 0 {
		return nil
	}

	var jsonBody any
	if err := json.Unmarshal(body, &jsonBody); err == nil {
		return maskData(jsonBody)
	}

	if strings.HasPrefix(strings.ToLower(contentType), "application/x-www-form-urlencoded") {
		values, err := url.ParseQuery(string(body))
		if err == nil {
			masked := make(map[string]any, len(values))
			for k, v := range values {
				if _, found := sensitiveKeys[strings.ToLower(k)]; found {
					masked[k] = "***"
					continue
				}
				if len(v) == 1 {
					masked[k] = v[0]
				} else {
					masked[k] = v
				}
			}
			return masked
		}
	}

	if !utf8.Valid(body) {
		return "<binary body omitted>"
	}
	if len(body) > maxLoggedBodyBytes {
		return string(body[:maxLoggedBodyBytes]) + "...(truncated)"
	}
	return string(body)
}

func middlewareLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := matchedRoutePath(r)
		start := time.Now()

		var reqBodyBytes []byte
		if r.Body != nil {
			//nolint:errcheck // best effort for logging only
			reqBodyBytes, _ = io.ReadAll(r.Body)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))

		slog.InfoContext(
			r.Context(),
			"request received",
			"method", r.Method,
			"route", route,
			"path", r.URL.Path,
			"headers", maskHeaders(r.Header),
			"body", parseAndMaskBody(r.Header.Get("Content-Type"), reqBodyBytes),
		)

		rec := &statusRecorder{ResponseWriter: w, body: &bytes.Buffer{}}
		next.ServeHTTP(rec, r)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}

		var respBody any
		if rec.body != nil {
			var respJSON any
			if err := json.Unmarshal(rec.body.Bytes(), &respJSON); err == nil {
				respBody = maskData(respJSON)
			} else if utf8.Valid(rec.body.Bytes()) {
				respBody = rec.body.String()
			} else if rec.body.Len() > 0 {
				respBody = "<binary body omitted>"
			}
			if rec.capped {
				respBody = map[string]any{
					"body":      respBody,
					"truncated": true,
				}
			}
		}

		slog.InfoContext(
			r.Context(),
			"response sent",
			"method", r.Method,
			"route", route,
			"path", r.URL.Path,
			"status", status,
			"bytes", rec.bytes,
			"latency_ms", time.Since(start).Milliseconds(),
			"body", respBody,
		)
	})
}

package router

import (
	"bufio"
	"bytes"
	"context"
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
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/instrument"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const maxLoggedBodyBytes = 32 * 1024 // 32KB

func maskHeaders(headers http.Header, maskKeys map[string]struct{}) http.Header {
	if len(maskKeys) == 0 {
		return headers
	}

	result := headers.Clone()
	for key := range result {
		if _, found := maskKeys[strings.ToLower(key)]; found {
			result.Set(key, "***")
		}
	}
	return result
}

func maskData(v any, maskKeys map[string]struct{}) any {
	switch val := v.(type) {
	case map[string]any:
		masked := make(map[string]any, len(val))
		for k, v2 := range val {
			if _, found := maskKeys[strings.ToLower(k)]; found {
				masked[k] = "***"
			} else {
				masked[k] = maskData(v2, maskKeys)
			}
		}
		return masked
	case []any:
		res := make([]any, len(val))
		for i, v2 := range val {
			res[i] = maskData(v2, maskKeys)
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
	err    error
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

func (w *statusRecorder) SetError(err error) {
	w.err = err
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

func parseAndMaskBody(contentType string, body []byte, maskKeys map[string]struct{}) any {
	if len(body) == 0 {
		return nil
	}

	var jsonBody any
	if err := json.Unmarshal(body, &jsonBody); err == nil {
		return maskData(jsonBody, maskKeys)
	}

	if strings.HasPrefix(strings.ToLower(contentType), "application/x-www-form-urlencoded") {
		values, err := url.ParseQuery(string(body))
		if err == nil {
			masked := make(map[string]any, len(values))
			for k, v := range values {
				if _, found := maskKeys[strings.ToLower(k)]; found {
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

func getMaskKeys(cfg config.Config) map[string]struct{} {
	maskKeys := make(map[string]struct{})
	if cfg != nil {
		for _, field := range cfg.GetArray("instrument.log_mask_fields") {
			field = strings.TrimSpace(strings.ToLower(field))
			if field == "" {
				continue
			}
			maskKeys[field] = struct{}{}
		}
	}

	return maskKeys
}

func readRequestBody(r *http.Request) []byte {
	if r.Body == nil {
		return nil
	}

	limited := io.LimitReader(r.Body, maxLoggedBodyBytes+1)
	//nolint:errcheck // best effort for logging only
	reqBodyBytes, _ := io.ReadAll(limited)
	r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(reqBodyBytes), r.Body))
	if len(reqBodyBytes) > maxLoggedBodyBytes {
		return reqBodyBytes[:maxLoggedBodyBytes]
	}
	return reqBodyBytes
}

func logRequest(ctx context.Context, r *http.Request, route string, body []byte, maskKeys map[string]struct{}) {
	slog.InfoContext(
		ctx,
		"request received",
		"method", r.Method,
		"path", route,
		"uri", r.RequestURI,
		"headers", maskHeaders(r.Header, maskKeys),
		"body", parseAndMaskBody(r.Header.Get("Content-Type"), body, maskKeys),
	)
}

func responseStatus(rec *statusRecorder) int {
	if rec.status == 0 {
		return http.StatusOK
	}
	return rec.status
}

func buildResponseBody(rec *statusRecorder, maskKeys map[string]struct{}) any {
	if rec.body == nil {
		return nil
	}

	var respBody any
	var respJSON any
	if err := json.Unmarshal(rec.body.Bytes(), &respJSON); err == nil {
		respBody = maskData(respJSON, maskKeys)
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

	return respBody
}

func middlewareObservability(cfg config.Config, ins instrument.Instrumentation) Middleware {
	maskKeys := getMaskKeys(cfg)
	tracer := ins.Tracer("http.server")
	meter := ins.Meter("http.server")

	requestCounter, err := meter.Int64Counter("http.server.requests", metric.WithDescription("Number of HTTP requests received"))
	if err != nil {
		slog.Error("failed to create http request counter", "error", err)
	}

	durationHistogram, err := meter.Float64Histogram("http.server.duration", metric.WithDescription("HTTP request duration in milliseconds"))
	if err != nil {
		slog.Error("failed to create http duration histogram", "error", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := matchedRoutePath(r)
			start := time.Now()

			ctx, span := tracer.Start(
				r.Context(),
				r.Method+" "+route,
				trace.WithAttributes(
					semconv.HTTPRequestMethodKey.String(r.Method),
					semconv.HTTPRouteKey.String(route),
				),
			)
			defer span.End()

			reqBodyBytes := readRequestBody(r)
			logRequest(ctx, r, route, reqBodyBytes, maskKeys)

			rec := &statusRecorder{ResponseWriter: w, body: &bytes.Buffer{}}
			next.ServeHTTP(rec, r.WithContext(ctx))

			status := responseStatus(rec)
			respBody := buildResponseBody(rec, maskKeys)
			elapsedMs := float64(time.Since(start).Milliseconds())

			attrs := []attribute.KeyValue{
				semconv.HTTPRequestMethodKey.String(r.Method),
				semconv.HTTPRouteKey.String(route),
				semconv.HTTPResponseStatusCodeKey.Int(status),
			}

			if rec.err != nil {
				span.RecordError(rec.err)
			}

			if status >= 500 {
				if rec.err != nil {
					span.SetStatus(codes.Error, rec.err.Error())
				} else {
					span.SetStatus(codes.Error, http.StatusText(status))
				}
			} else {
				span.SetStatus(codes.Ok, "")
			}

			span.SetAttributes(attrs...)
			if requestCounter != nil {
				requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
			}
			if durationHistogram != nil {
				durationHistogram.Record(ctx, elapsedMs, metric.WithAttributes(attrs...))
			}

			span.SetAttributes(
				semconv.NetworkProtocolVersionKey.String(r.Proto),
				semconv.ServerAddressKey.String(r.Host),
				attribute.String("http.target", r.URL.Path),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.Int("http.response_content_length", rec.bytes),
			)

			slog.InfoContext(
				ctx,
				"response sent",
				"method", r.Method,
				"path", route,
				"uri", r.RequestURI,
				"status", status,
				"bytes", rec.bytes,
				"latency_ms", time.Since(start).Milliseconds(),
				"body", respBody,
			)
		})
	}
}

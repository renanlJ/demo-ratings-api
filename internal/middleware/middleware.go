// Package middleware provides HTTP middleware for request-context propagation,
// structured access logging and metrics collection.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/devexio/demo-ratings-api/internal/metrics"
)

type ctxKey string

const (
	requestIDKey ctxKey = "requestID"
	traceIDKey   ctxKey = "traceID"
)

var traceparentRe = regexp.MustCompile(`^[0-9a-f]{2}-([0-9a-f]{32})-[0-9a-f]{16}-[0-9a-f]{2}$`)

// RequestIDFrom returns the request ID stored in ctx, or an empty string.
func RequestIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}

// TraceIDFrom returns the W3C trace ID stored in ctx, or an empty string.
func TraceIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(traceIDKey).(string)
	return v
}

// Context enriches each request with a request ID (X-Request-ID, generated when
// absent) and the W3C trace ID parsed from the traceparent header. The request
// ID is echoed back on the response so callers can correlate.
func Context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if reqID == "" {
			reqID = newID()
		}
		traceID := ""
		if tp := r.Header.Get("traceparent"); tp != "" {
			if m := traceparentRe.FindStringSubmatch(strings.ToLower(tp)); m != nil {
				traceID = m[1]
			}
		}
		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		ctx = context.WithValue(ctx, traceIDKey, traceID)
		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logging emits one structured JSON log line per completed request.
func Logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			log.Info("http request",
				"method", r.Method,
				"route", routePattern(r),
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", RequestIDFrom(r.Context()),
				"trace_id", TraceIDFrom(r.Context()),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// Metrics records Prometheus-style observations for each completed request.
func Metrics(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			m.Observe(routePattern(r), r.Method, rec.status, time.Since(start))
		})
	}
}

func routePattern(r *http.Request) string {
	// r.Pattern is the ServeMux pattern, e.g. "GET /api/v1/ratings/{productId}".
	if p := strings.TrimPrefix(r.Pattern, r.Method+" "); p != r.Pattern {
		return p
	}
	if r.Pattern != "" {
		return r.Pattern
	}
	return r.URL.Path
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

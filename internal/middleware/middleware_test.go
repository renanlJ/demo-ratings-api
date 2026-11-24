package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devexio/demo-ratings-api/internal/metrics"
)

func captureCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Captured-Request-ID", RequestIDFrom(r.Context()))
		w.Header().Set("X-Captured-Trace-ID", TraceIDFrom(r.Context()))
		next.ServeHTTP(w, r)
	})
}

func TestContextGeneratesRequestID(t *testing.T) {
	h := Context(captureCtx(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Request-ID"); len(got) != 32 {
		t.Errorf("expected generated 32-char request ID, got %q", got)
	}
	if got := rec.Header().Get("X-Captured-Request-ID"); got != rec.Header().Get("X-Request-ID") {
		t.Errorf("request ID not propagated to context")
	}
}

func TestContextPropagatesInboundHeaders(t *testing.T) {
	h := Context(captureCtx(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "req-abc-123")
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "req-abc-123" {
		t.Errorf("expected propagated request ID, got %q", got)
	}
	if got := rec.Header().Get("X-Captured-Trace-ID"); got != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Errorf("expected trace ID, got %q", got)
	}
}

func TestMetricsAndLoggingMiddleware(t *testing.T) {
	m := metrics.New()
	log := slog.New(slog.NewJSONHandler(discardWriter{}, nil))

	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	h = Metrics(m)(h)
	h = Logging(log)(h)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := rec.Body.String()
	if !strings.Contains(body, "http_requests_total") || !strings.Contains(body, `status="418"`) {
		t.Errorf("expected recorded request with status 418, got:\n%s", body)
	}
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

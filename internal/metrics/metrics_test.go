package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestObserveAndExposition(t *testing.T) {
	m := New()
	m.Observe("/api/v1/ratings/{productId}", "GET", 200, 5*time.Millisecond)
	m.Observe("/api/v1/ratings/{productId}", "GET", 200, 15*time.Millisecond)
	m.Observe("/api/v1/ratings/{productId}", "GET", 503, 100*time.Millisecond)

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	body := rec.Body.String()

	for _, want := range []string{
		"http_requests_total{method=\"GET\",route=\"/api/v1/ratings/{productId}\",status=\"200\"} 2",
		"http_requests_errors_total{method=\"GET\",route=\"/api/v1/ratings/{productId}\",status=\"503\"} 1",
		"http_request_duration_seconds_count",
		"le=\"+Inf\"",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("expected exposition to contain %q, got:\n%s", want, body)
		}
	}
}

func TestContentType(t *testing.T) {
	m := New()
	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("unexpected content type %q", ct)
	}
}

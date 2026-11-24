package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devexio/demo-ratings-api/internal/metrics"
	"github.com/devexio/demo-ratings-api/internal/store"
)

func newMux(ready bool) *http.ServeMux {
	return NewAPI(store.New(), metrics.New(), func() bool { return ready }).Mux()
}

func TestGetRatingsByProduct(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ratings/SKU-1002", nil)
	rec := httptest.NewRecorder()
	newMux(true).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body struct {
		ProductID    string         `json:"productId"`
		Average      float64        `json:"average"`
		Count        int            `json:"count"`
		Distribution map[string]int `json:"distribution"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body.ProductID != "SKU-1002" || body.Count != 6 || body.Average != 4.33 {
		t.Errorf("unexpected body: %+v", body)
	}
}

func TestGetRatingsUnknownProductReturns404(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ratings/SKU-9999", nil)
	rec := httptest.NewRecorder()
	newMux(true).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestListRatings(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ratings", nil)
	rec := httptest.NewRecorder()
	newMux(true).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(body) != 3 {
		t.Errorf("expected 3 products, got %d", len(body))
	}
}

func TestHealthAndReadiness(t *testing.T) {
	mux := newMux(true)
	for _, path := range []string{"/healthz", "/readyz"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", path, rec.Code)
		}
	}
}

func TestReadinessFailsWhenNotReady(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux(false).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestMetricsEndpointExposesPrometheusFormat(t *testing.T) {
	mux := newMux(true)
	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "# TYPE http_requests_total counter") {
		t.Error("expected http_requests_total metric in exposition")
	}
}

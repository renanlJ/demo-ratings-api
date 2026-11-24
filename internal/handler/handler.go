// Package handler wires the HTTP API endpoints.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/devexio/demo-ratings-api/internal/metrics"
	"github.com/devexio/demo-ratings-api/internal/store"
)

// API bundles the dependencies the HTTP handlers need.
type API struct {
	store   *store.Store
	metrics *metrics.Metrics
	ready   func() bool
}

// NewAPI builds the API handler set.
func NewAPI(st *store.Store, m *metrics.Metrics, ready func() bool) *API {
	return &API{store: st, metrics: m, ready: ready}
}

// Mux returns the service router.
func (a *API) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/ratings", a.list)
	mux.HandleFunc("GET /api/v1/ratings/{productId}", a.byProduct)
	mux.HandleFunc("GET /healthz", a.health)
	mux.HandleFunc("GET /readyz", a.readyz)
	mux.Handle("GET /metrics", a.metrics.Handler())
	return mux
}

func (a *API) list(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.store.List())
}

func (a *API) byProduct(w http.ResponseWriter, r *http.Request) {
	sum, ok := a.store.Summary(r.PathValue("productId"))
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
		return
	}
	writeJSON(w, http.StatusOK, sum)
}

func (a *API) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) readyz(w http.ResponseWriter, _ *http.Request) {
	if !a.ready() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

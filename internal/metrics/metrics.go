// Package metrics implements a minimal Prometheus text-format registry.
//
// The service is intentionally dependency-free (Go standard library only), so
// instead of importing prometheus/client_golang this file emits the subset of
// the exposition format the platform needs.
package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5}

// histogram keeps cumulative bucket counters, matching Prometheus semantics.
type histogram struct {
	counts []uint64
	sum    float64
	total  uint64
}

func (h *histogram) observe(v float64) {
	h.total++
	h.sum += v
	for i, b := range buckets {
		if v <= b {
			h.counts[i]++
		}
	}
}

type labelSet struct {
	route  string
	method string
	status int
}

// Metrics is a thread-safe in-process metrics registry.
type Metrics struct {
	mu        sync.Mutex
	requests  map[labelSet]uint64
	errors    map[labelSet]uint64
	durations map[labelSet]*histogram
}

// New returns an empty registry.
func New() *Metrics {
	return &Metrics{
		requests:  make(map[labelSet]uint64),
		errors:    make(map[labelSet]uint64),
		durations: make(map[labelSet]*histogram),
	}
}

// Observe records one completed HTTP request. Statuses >= 500 count as errors.
func (m *Metrics) Observe(route, method string, status int, d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	l := labelSet{route: route, method: method, status: status}
	m.requests[l]++
	if status >= 500 {
		m.errors[l]++
	}
	h, ok := m.durations[l]
	if !ok {
		h = &histogram{counts: make([]uint64, len(buckets))}
		m.durations[l] = h
	}
	h.observe(d.Seconds())
}

// Handler exposes the registry in Prometheus text exposition format.
func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		var b strings.Builder

		b.WriteString("# HELP http_requests_total Total number of HTTP requests processed.\n")
		b.WriteString("# TYPE http_requests_total counter\n")
		for _, k := range sortedKeys(m.requests) {
			fmt.Fprintf(&b, "http_requests_total{method=%q,route=%q,status=%q} %d\n",
				k.method, k.route, strconv.Itoa(k.status), m.requests[k])
		}

		b.WriteString("# HELP http_requests_errors_total Total number of HTTP requests that returned a server error.\n")
		b.WriteString("# TYPE http_requests_errors_total counter\n")
		for _, k := range sortedKeys(m.errors) {
			fmt.Fprintf(&b, "http_requests_errors_total{method=%q,route=%q,status=%q} %d\n",
				k.method, k.route, strconv.Itoa(k.status), m.errors[k])
		}

		b.WriteString("# HELP http_request_duration_seconds HTTP request duration in seconds.\n")
		b.WriteString("# TYPE http_request_duration_seconds histogram\n")
		for _, k := range sortedKeys(m.durations) {
			h := m.durations[k]
			labels := fmt.Sprintf("method=%q,route=%q,status=%q", k.method, k.route, strconv.Itoa(k.status))
			for i, bound := range buckets {
				fmt.Fprintf(&b, "http_request_duration_seconds_bucket{%s,le=%q} %d\n",
					labels, strconv.FormatFloat(bound, 'g', -1, 64), h.counts[i])
			}
			fmt.Fprintf(&b, "http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n", labels, h.total)
			fmt.Fprintf(&b, "http_request_duration_seconds_sum{%s} %g\n", labels, h.sum)
			fmt.Fprintf(&b, "http_request_duration_seconds_count{%s} %d\n", labels, h.total)
		}

		_, _ = w.Write([]byte(b.String()))
	})
}

func sortedKeys[V any](m map[labelSet]V) []labelSet {
	keys := make([]labelSet, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].route != keys[j].route {
			return keys[i].route < keys[j].route
		}
		if keys[i].method != keys[j].method {
			return keys[i].method < keys[j].method
		}
		return keys[i].status < keys[j].status
	})
	return keys
}

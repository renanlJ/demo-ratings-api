// Package store provides a deterministic in-memory ratings repository.
package store

import (
	"math"
	"sort"
	"strconv"
)

// Summary is the aggregated rating view returned by the API.
type Summary struct {
	ProductID    string         `json:"productId"`
	Average      float64        `json:"average"`
	Count        int            `json:"count"`
	Distribution map[string]int `json:"distribution"`
}

// Store is a read-only, deterministic in-memory repository keyed by product ID.
type Store struct {
	ratings map[string][]int
}

// New returns a Store preloaded with demo fixtures shared across the platform.
func New() *Store {
	return &Store{ratings: map[string][]int{
		"SKU-1001": {5, 4, 5, 4, 5, 3, 5},
		"SKU-1002": {4, 4, 5, 4, 4, 5},
		"SKU-1003": {3, 4, 2, 4, 3},
	}}
}

// Summary returns the aggregated rating summary for a product.
func (s *Store) Summary(productID string) (Summary, bool) {
	values, ok := s.ratings[productID]
	if !ok || len(values) == 0 {
		return Summary{}, false
	}
	sum := Summary{
		ProductID:    productID,
		Count:        len(values),
		Distribution: map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0},
	}
	var total int
	for _, v := range values {
		total += v
		sum.Distribution[strconv.Itoa(v)]++
	}
	sum.Average = math.Round((float64(total)/float64(len(values)))*100) / 100
	return sum, true
}

// List returns summaries for all known products, sorted by product ID.
func (s *Store) List() []Summary {
	ids := make([]string, 0, len(s.ratings))
	for id := range s.ratings {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	out := make([]Summary, 0, len(ids))
	for _, id := range ids {
		sum, _ := s.Summary(id)
		out = append(out, sum)
	}
	return out
}

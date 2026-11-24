package store

import "testing"

func TestSummaryAggregatesValues(t *testing.T) {
	s := New()
	sum, ok := s.Summary("SKU-1001")
	if !ok {
		t.Fatal("expected SKU-1001 to exist")
	}
	if sum.ProductID != "SKU-1001" {
		t.Errorf("unexpected productId: %s", sum.ProductID)
	}
	if sum.Count != 7 {
		t.Errorf("expected count 7, got %d", sum.Count)
	}
	if sum.Average != 4.43 {
		t.Errorf("expected average 4.43, got %v", sum.Average)
	}
	if sum.Distribution["5"] != 4 || sum.Distribution["4"] != 2 || sum.Distribution["3"] != 1 {
		t.Errorf("unexpected distribution: %v", sum.Distribution)
	}
	if len(sum.Distribution) != 5 {
		t.Errorf("expected 5 distribution buckets, got %d", len(sum.Distribution))
	}
}

func TestSummaryUnknownProduct(t *testing.T) {
	s := New()
	if _, ok := s.Summary("SKU-9999"); ok {
		t.Error("expected SKU-9999 to be unknown")
	}
}

func TestListIsSortedAndComplete(t *testing.T) {
	s := New()
	list := s.List()
	if len(list) != 3 {
		t.Fatalf("expected 3 summaries, got %d", len(list))
	}
	for i := 1; i < len(list); i++ {
		if list[i-1].ProductID >= list[i].ProductID {
			t.Errorf("list not sorted: %v", list)
		}
	}
}

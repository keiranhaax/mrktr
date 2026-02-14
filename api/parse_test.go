package api

import "testing"

func TestParseSearchResults(t *testing.T) {
	data := []SearchResult{
		{
			URL:         "https://www.ebay.com/itm/1",
			Title:       "Console Bundle $1,299.99 Sold",
			Description: "brand new sealed",
		},
		{
			URL:         "https://www.mercari.com/us/item/2",
			Title:       "Controller",
			Description: "Good condition only $80.00",
		},
		{
			URL:         "https://example.com/ignore",
			Title:       "No price here",
			Description: "missing amount",
		},
	}

	got := ParseSearchResults(data)
	if len(got) != 2 {
		t.Fatalf("expected 2 parsed results, got %d", len(got))
	}

	first := got[0]
	if first.Platform != "eBay" {
		t.Fatalf("expected eBay platform, got %q", first.Platform)
	}
	if first.Price != 1299.99 {
		t.Fatalf("expected first price 1299.99, got %v", first.Price)
	}
	if first.Condition != "New" {
		t.Fatalf("expected first condition New, got %q", first.Condition)
	}
	if first.Status != "Sold" {
		t.Fatalf("expected first status Sold, got %q", first.Status)
	}

	second := got[1]
	if second.Platform != "Mercari" {
		t.Fatalf("expected Mercari platform, got %q", second.Platform)
	}
	if second.Condition != "Good" {
		t.Fatalf("expected second condition Good, got %q", second.Condition)
	}
	if second.Status != "Active" {
		t.Fatalf("expected second status Active, got %q", second.Status)
	}
}

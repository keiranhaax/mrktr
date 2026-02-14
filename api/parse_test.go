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

func TestParseSearchResultsUsesWordBoundariesForConditionAndStatus(t *testing.T) {
	data := []SearchResult{
		{
			URL:         "https://www.ebay.com/itm/1",
			Title:       "Renewed Controller $99.00",
			Description: "latest news drop",
		},
		{
			URL:         "https://www.ebay.com/itm/2",
			Title:       "Collector item $49.00",
			Description: "unsold inventory",
		},
	}

	got := ParseSearchResults(data)
	if len(got) != 2 {
		t.Fatalf("expected 2 parsed results, got %d", len(got))
	}

	if got[0].Condition != "Used" {
		t.Fatalf("expected renewed/news text to remain Used, got %q", got[0].Condition)
	}
	if got[1].Status != "Active" {
		t.Fatalf("expected unsold text to remain Active, got %q", got[1].Status)
	}
}

func TestParseSearchResultsSupportsFlexibleAndMultiplePrices(t *testing.T) {
	data := []SearchResult{
		{
			URL:         "https://www.ebay.com/itm/1",
			Title:       "Deal was $150 now $99",
			Description: "good condition",
		},
		{
			URL:         "https://www.ebay.com/itm/2",
			Title:       "Controller for $99.9",
			Description: "used",
		},
		{
			URL:         "https://www.ebay.com/itm/3",
			Title:       "Console $1299 with extras",
			Description: "sealed",
		},
	}

	got := ParseSearchResults(data)
	if len(got) != 3 {
		t.Fatalf("expected 3 parsed results, got %d", len(got))
	}

	if got[0].Price != 99 {
		t.Fatalf("expected lowest matched price 99, got %v", got[0].Price)
	}
	if got[1].Price != 99.9 {
		t.Fatalf("expected single-decimal price 99.9, got %v", got[1].Price)
	}
	if got[2].Price != 1299 {
		t.Fatalf("expected non-comma 4-digit price 1299, got %v", got[2].Price)
	}
}

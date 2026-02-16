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

func TestParseSearchResultsSupportsUSDPriceFormats(t *testing.T) {
	data := []SearchResult{
		{
			URL:         "https://www.ebay.co.uk/itm/1",
			Title:       "Console bundle USD 1,099.50",
			Description: "used",
		},
		{
			URL:         "https://www.amazon.co.jp/item/2",
			Title:       "Headset",
			Description: "asking 89.9 excellent condition",
		},
	}

	got := ParseSearchResults(data)
	if len(got) != 2 {
		t.Fatalf("expected 2 parsed results, got %d", len(got))
	}
	if got[0].Price != 1099.50 {
		t.Fatalf("expected USD-prefixed price 1099.50, got %v", got[0].Price)
	}
	if got[1].Price != 89.9 {
		t.Fatalf("expected context-inferred price 89.9, got %v", got[1].Price)
	}
}

func TestParseSearchResultsDetectsRegionalPlatformHosts(t *testing.T) {
	data := []SearchResult{
		{URL: "https://www.ebay.co.uk/itm/1", Title: "$99 listing", Description: "used"},
		{URL: "https://www.amazon.de/dp/1", Title: "$75 listing", Description: "used"},
		{URL: "https://www.mercari.jp/item/1", Title: "$65 listing", Description: "used"},
		{URL: "https://marketplace.facebook.com/item/1", Title: "$55 listing", Description: "used"},
	}

	got := ParseSearchResults(data)
	if len(got) != 4 {
		t.Fatalf("expected 4 parsed results, got %d", len(got))
	}
	if got[0].Platform != "eBay" {
		t.Fatalf("expected eBay platform for regional host, got %q", got[0].Platform)
	}
	if got[1].Platform != "Amazon" {
		t.Fatalf("expected Amazon platform for regional host, got %q", got[1].Platform)
	}
	if got[2].Platform != "Mercari" {
		t.Fatalf("expected Mercari platform for regional host, got %q", got[2].Platform)
	}
	if got[3].Platform != "Facebook" {
		t.Fatalf("expected Facebook platform for regional host, got %q", got[3].Platform)
	}
}

func TestParseSearchResultsSoldNegationStaysActive(t *testing.T) {
	data := []SearchResult{
		{
			URL:         "https://www.ebay.com/itm/1",
			Title:       "Controller $99",
			Description: "not sold yet",
		},
	}

	got := ParseSearchResults(data)
	if len(got) != 1 {
		t.Fatalf("expected 1 parsed result, got %d", len(got))
	}
	if got[0].Status != "Active" {
		t.Fatalf("expected negated sold status to remain Active, got %q", got[0].Status)
	}
}

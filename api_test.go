package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseSearchResults(t *testing.T) {
	data := []struct {
		URL         string `json:"url"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}{
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

	got := parseSearchResults(data)
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

func TestSearchPricesDemoModeWithoutKeys(t *testing.T) {
	t.Setenv("FIRECRAWL_API_KEY", "")
	t.Setenv("TAVILY_API_KEY", "")

	resp := SearchPrices("ps5")
	if resp.Mode != searchModeDemo {
		t.Fatalf("expected demo mode, got %q", resp.Mode)
	}
	if len(resp.Results) == 0 {
		t.Fatal("expected demo data results")
	}
	if resp.Warning != "" {
		t.Fatalf("expected no warning in demo mode, got %q", resp.Warning)
	}
}

func TestSearchPricesLiveModeFromFirecrawl(t *testing.T) {
	t.Setenv("FIRECRAWL_API_KEY", "firecrawl-key")
	t.Setenv("TAVILY_API_KEY", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"url":"https://ebay.com/1","title":"PS5 $499.00","description":"new"}]}`))
	}))
	defer server.Close()

	setTestHTTPGlobals(t, server.Client(), server.URL, tavilySearchURL)

	resp := SearchPrices("ps5")
	if resp.Mode != searchModeLive {
		t.Fatalf("expected live mode, got %q", resp.Mode)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 live result, got %d", len(resp.Results))
	}
	if resp.Warning != "" {
		t.Fatalf("expected no warning for live result, got %q", resp.Warning)
	}
}

func TestSearchPricesLiveModeWithEmptyResults(t *testing.T) {
	t.Setenv("FIRECRAWL_API_KEY", "firecrawl-key")
	t.Setenv("TAVILY_API_KEY", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	setTestHTTPGlobals(t, server.Client(), server.URL, tavilySearchURL)

	resp := SearchPrices("nothing")
	if resp.Mode != searchModeLive {
		t.Fatalf("expected live mode for successful empty response, got %q", resp.Mode)
	}
	if len(resp.Results) != 0 {
		t.Fatalf("expected empty results, got %d", len(resp.Results))
	}
	if resp.Warning != "" {
		t.Fatalf("expected no warning for successful empty response, got %q", resp.Warning)
	}
}

func TestSearchPricesFallbackModeWhenProvidersFail(t *testing.T) {
	t.Setenv("FIRECRAWL_API_KEY", "firecrawl-key")
	t.Setenv("TAVILY_API_KEY", "tavily-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream unavailable"))
	}))
	defer server.Close()

	setTestHTTPGlobals(
		t,
		server.Client(),
		server.URL+"/firecrawl",
		server.URL+"/tavily",
	)

	resp := SearchPrices("ps5")
	if resp.Mode != searchModeFallback {
		t.Fatalf("expected fallback mode, got %q", resp.Mode)
	}
	if len(resp.Results) == 0 {
		t.Fatal("expected fallback mock results")
	}
	if !strings.Contains(resp.Warning, "Firecrawl") || !strings.Contains(resp.Warning, "Tavily") {
		t.Fatalf("expected warning to mention failed providers, got %q", resp.Warning)
	}
}

func TestSearchFirecrawlChecksHTTPStatus(t *testing.T) {
	t.Setenv("FIRECRAWL_API_KEY", "firecrawl-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server exploded"))
	}))
	defer server.Close()

	setTestHTTPGlobals(t, server.Client(), server.URL, tavilySearchURL)

	_, err := searchFirecrawl("ps5")
	if err == nil {
		t.Fatal("expected error for non-2xx firecrawl response")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}

func setTestHTTPGlobals(t *testing.T, client *http.Client, firecrawlURL, tavilyURL string) {
	t.Helper()

	oldClient := httpClient
	oldFirecrawlURL := firecrawlSearchURL
	oldTavilyURL := tavilySearchURL

	httpClient = client
	firecrawlSearchURL = firecrawlURL
	tavilySearchURL = tavilyURL

	t.Cleanup(func() {
		httpClient = oldClient
		firecrawlSearchURL = oldFirecrawlURL
		tavilySearchURL = oldTavilyURL
	})
}

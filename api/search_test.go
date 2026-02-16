package api

import (
	"context"
	"errors"
	"strings"
	"testing"

	"mrktr/types"
)

type stubProvider struct {
	name       string
	configured bool
	results    []types.Listing
	err        error
}

type contextProbeProvider struct {
	sawCanceled bool
}

func (p stubProvider) Name() string {
	return p.name
}

func (p stubProvider) Configured() bool {
	return p.configured
}

func (p stubProvider) Search(_ context.Context, _ string) ([]types.Listing, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.results, nil
}

func (p *contextProbeProvider) Name() string {
	return "Probe"
}

func (p *contextProbeProvider) Configured() bool {
	return true
}

func (p *contextProbeProvider) Search(ctx context.Context, _ string) ([]types.Listing, error) {
	if err := ctx.Err(); err != nil {
		p.sawCanceled = true
		return nil, err
	}
	return []types.Listing{}, nil
}

func TestSearchPricesUnavailableWithoutProviders(t *testing.T) {
	client := NewClient()

	resp := client.SearchPrices("ps5")
	if resp.Mode != SearchModeUnavailable {
		t.Fatalf("expected unavailable mode, got %q", resp.Mode)
	}
	if len(resp.Results) != 0 {
		t.Fatalf("expected no fallback results, got %d", len(resp.Results))
	}
	if resp.Err == nil {
		t.Fatal("expected error when no live providers are configured")
	}
}

func TestSearchPricesUnavailableWithNilClient(t *testing.T) {
	var client *Client
	resp := client.SearchPrices("ps5")
	if resp.Mode != SearchModeUnavailable {
		t.Fatalf("expected unavailable mode, got %q", resp.Mode)
	}
	if resp.Err == nil {
		t.Fatal("expected error when client is nil")
	}
}

func TestSearchPricesLiveModeFromFirstConfiguredProvider(t *testing.T) {
	client := NewClient(
		stubProvider{
			name:       "Brave",
			configured: true,
			results: []types.Listing{
				{URL: "https://ebay.com/1", Price: 499.0, Platform: "eBay", Status: "Active", Condition: "Used", Title: "PS5"},
			},
		},
		stubProvider{
			name:       "Tavily",
			configured: true,
			results: []types.Listing{
				{URL: "https://mercari.com/1", Price: 450.0, Platform: "Mercari", Status: "Active", Condition: "Used", Title: "PS5"},
			},
		},
	)

	resp := client.SearchPrices("ps5")
	if resp.Mode != SearchModeLive {
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
	client := NewClient(
		stubProvider{name: "Brave", configured: true, results: []types.Listing{}},
	)

	resp := client.SearchPrices("nothing")
	if resp.Mode != SearchModeLive {
		t.Fatalf("expected live mode for successful empty response, got %q", resp.Mode)
	}
	if len(resp.Results) != 0 {
		t.Fatalf("expected empty results, got %d", len(resp.Results))
	}
	if resp.Warning != "" {
		t.Fatalf("expected no warning for successful empty response, got %q", resp.Warning)
	}
}

func TestSearchPricesLiveEmptyIncludesWarningsWhenSomeProvidersFail(t *testing.T) {
	client := NewClient(
		stubProvider{name: "Brave", configured: true, err: errors.New("upstream unavailable")},
		stubProvider{name: "Tavily", configured: true, results: []types.Listing{}},
	)

	resp := client.SearchPrices("ps5")
	if resp.Mode != SearchModeLive {
		t.Fatalf("expected live mode when at least one provider succeeds, got %q", resp.Mode)
	}
	if len(resp.Results) != 0 {
		t.Fatalf("expected no listings for empty successful response, got %d", len(resp.Results))
	}
	if !strings.Contains(resp.Warning, "Brave") {
		t.Fatalf("expected warning to mention failed provider, got %q", resp.Warning)
	}
	if len(resp.ProviderErrors) != 1 {
		t.Fatalf("expected one provider error, got %d", len(resp.ProviderErrors))
	}
}

func TestSearchPricesUnavailableWhenProvidersFail(t *testing.T) {
	client := NewClient(
		stubProvider{name: "Brave", configured: true, err: errors.New("upstream unavailable")},
		stubProvider{name: "Tavily", configured: true, err: errors.New("upstream unavailable")},
		stubProvider{name: "Firecrawl", configured: true, err: errors.New("upstream unavailable")},
	)

	resp := client.SearchPrices("ps5")
	if resp.Mode != SearchModeUnavailable {
		t.Fatalf("expected unavailable mode, got %q", resp.Mode)
	}
	if len(resp.Results) != 0 {
		t.Fatalf("expected no results when all providers fail, got %d", len(resp.Results))
	}
	if resp.Err == nil {
		t.Fatal("expected error when all providers fail")
	}
	if !strings.Contains(resp.Warning, "Brave") ||
		!strings.Contains(resp.Warning, "Tavily") ||
		!strings.Contains(resp.Warning, "Firecrawl") {
		t.Fatalf("expected warning to mention failed providers, got %q", resp.Warning)
	}
}

func TestSearchPricesActionableWarningsFromHTTPStatus(t *testing.T) {
	client := NewClient(
		stubProvider{name: "Brave", configured: true, err: &HTTPStatusError{Provider: "Brave", Status: 401, Body: "bad token"}},
		stubProvider{name: "Tavily", configured: true, err: &HTTPStatusError{Provider: "Tavily", Status: 429, Body: "slow down"}},
	)

	resp := client.SearchPrices("ps5")
	if resp.Mode != SearchModeUnavailable {
		t.Fatalf("expected unavailable mode, got %q", resp.Mode)
	}
	if !strings.Contains(resp.Warning, "BRAVE_API_KEY") {
		t.Fatalf("expected actionable auth warning, got %q", resp.Warning)
	}
	if !strings.Contains(resp.Warning, "Try again in 60s") {
		t.Fatalf("expected actionable rate-limit warning, got %q", resp.Warning)
	}
}

func TestSearchPricesSkipsUnconfiguredProviders(t *testing.T) {
	client := NewClient(
		stubProvider{name: "Brave", configured: false},
		stubProvider{
			name:       "Tavily",
			configured: true,
			results: []types.Listing{
				{URL: "https://ebay.com/1", Price: 200.0, Platform: "eBay", Status: "Active", Condition: "Used", Title: "Switch"},
			},
		},
	)

	resp := client.SearchPrices("switch")
	if resp.Mode != SearchModeLive {
		t.Fatalf("expected live mode, got %q", resp.Mode)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 listing, got %d", len(resp.Results))
	}
}

func TestSearchPricesContextPassesContextToProviders(t *testing.T) {
	probe := &contextProbeProvider{}
	client := NewClient(probe)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resp := client.SearchPricesContext(ctx, "ps5")
	if resp.Mode != SearchModeUnavailable {
		t.Fatalf("expected unavailable mode for canceled context, got %q", resp.Mode)
	}
	if !errors.Is(resp.Err, context.Canceled) {
		t.Fatalf("expected canceled root error, got %v", resp.Err)
	}
	if !probe.sawCanceled {
		t.Fatal("expected provider to observe canceled context")
	}
}

func TestSearchPricesProviderErrorClassification(t *testing.T) {
	client := NewClient(
		stubProvider{name: "Brave", configured: true, err: &HTTPStatusError{Provider: "Brave", Status: 401, Body: "bad token"}},
		stubProvider{name: "Tavily", configured: true, err: context.DeadlineExceeded},
	)

	resp := client.SearchPrices("ps5")
	if len(resp.ProviderErrors) != 2 {
		t.Fatalf("expected two provider errors, got %d", len(resp.ProviderErrors))
	}

	kinds := map[string]ProviderErrorKind{}
	for _, providerErr := range resp.ProviderErrors {
		kinds[providerErr.Provider] = providerErr.Kind
	}
	if kinds["Brave"] != ProviderErrorAuth {
		t.Fatalf("expected Brave auth classification, got %q", kinds["Brave"])
	}
	if kinds["Tavily"] != ProviderErrorTimeout {
		t.Fatalf("expected Tavily timeout classification, got %q", kinds["Tavily"])
	}
}

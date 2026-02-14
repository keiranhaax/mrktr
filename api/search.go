package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"mrktr/types"
)

// SearchMode indicates how search results were produced.
type SearchMode string

const (
	SearchModeLive        SearchMode = "live"
	SearchModeUnavailable SearchMode = "unavailable"
)

const (
	DefaultBraveSearchURL     = "https://api.search.brave.com/res/v1/web/search"
	DefaultFirecrawlSearchURL = "https://api.firecrawl.dev/v1/search"
	DefaultTavilySearchURL    = "https://api.tavily.com/search"
)

// SearchResponse wraps search results with metadata used by the UI.
type SearchResponse struct {
	Results []types.Listing
	Mode    SearchMode
	Warning string
	Err     error
}

// SearchProvider abstracts a single provider implementation.
type SearchProvider interface {
	Name() string
	Configured() bool
	Search(ctx context.Context, query string) ([]types.Listing, error)
}

// Client coordinates provider execution order.
type Client struct {
	providers []SearchProvider
}

// NewClient creates a search client from providers.
func NewClient(providers ...SearchProvider) *Client {
	return &Client{providers: providers}
}

// NewEnvClient builds a default client from process environment variables.
func NewEnvClient() *Client {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	return NewClient(
		NewBraveProvider(os.Getenv("BRAVE_API_KEY"), DefaultBraveSearchURL, httpClient),
		NewTavilyProvider(os.Getenv("TAVILY_API_KEY"), DefaultTavilySearchURL, httpClient),
		// Firecrawl remains available as a tertiary live provider.
		NewFirecrawlProvider(os.Getenv("FIRECRAWL_API_KEY"), DefaultFirecrawlSearchURL, httpClient),
	)
}

// SearchPrices searches for item prices across available providers.
func (c *Client) SearchPrices(query string) SearchResponse {
	return c.SearchPricesContext(context.Background(), query)
}

// SearchPricesContext searches for item prices using a caller-provided context.
func (c *Client) SearchPricesContext(ctx context.Context, query string) SearchResponse {
	if ctx == nil {
		ctx = context.Background()
	}

	q := strings.TrimSpace(query)

	if c == nil {
		return SearchResponse{
			Results: []types.Listing{},
			Mode:    SearchModeUnavailable,
			Err:     fmt.Errorf("search client not configured"),
		}
	}

	var configuredProviders int
	for _, provider := range c.providers {
		if provider != nil && provider.Configured() {
			configuredProviders++
		}
	}

	if configuredProviders == 0 {
		return SearchResponse{
			Results: []types.Listing{},
			Mode:    SearchModeUnavailable,
			Err: fmt.Errorf(
				"no live search providers configured; set BRAVE_API_KEY, TAVILY_API_KEY, or FIRECRAWL_API_KEY",
			),
		}
	}

	var successfulProviders int
	var failedProviders []string

	tryProvider := func(provider SearchProvider) *SearchResponse {
		if provider == nil || !provider.Configured() {
			return nil
		}

		results, err := provider.Search(ctx, q)
		if err != nil {
			name := provider.Name()
			if strings.TrimSpace(name) == "" {
				name = "Provider"
			}
			failedProviders = append(failedProviders, name)
			return nil
		}

		successfulProviders++
		if len(results) > 0 {
			resp := SearchResponse{Results: results, Mode: SearchModeLive}
			return &resp
		}
		return nil
	}

	for _, provider := range c.providers {
		if resp := tryProvider(provider); resp != nil {
			return *resp
		}
	}

	if successfulProviders > 0 {
		return SearchResponse{Results: []types.Listing{}, Mode: SearchModeLive}
	}

	warning := "Live search unavailable."
	if len(failedProviders) > 0 {
		warning = fmt.Sprintf(
			"Live search unavailable (%s).",
			strings.Join(failedProviders, ", "),
		)
	}

	return SearchResponse{
		Results: []types.Listing{},
		Mode:    SearchModeUnavailable,
		Warning: warning,
		Err:     fmt.Errorf("%s", warning),
	}
}

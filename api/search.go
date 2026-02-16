package api

import (
	"context"
	"errors"
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

// ProviderErrorKind classifies search-provider failures.
type ProviderErrorKind string

const (
	ProviderErrorUnknown   ProviderErrorKind = "unknown"
	ProviderErrorCanceled  ProviderErrorKind = "canceled"
	ProviderErrorTimeout   ProviderErrorKind = "timeout"
	ProviderErrorAuth      ProviderErrorKind = "auth"
	ProviderErrorRateLimit ProviderErrorKind = "rate_limit"
	ProviderErrorHTTP      ProviderErrorKind = "http"
	ProviderErrorTransport ProviderErrorKind = "transport"
)

const (
	DefaultBraveSearchURL     = "https://api.search.brave.com/res/v1/web/search"
	DefaultFirecrawlSearchURL = "https://api.firecrawl.dev/v1/search"
	DefaultTavilySearchURL    = "https://api.tavily.com/search"
)

// SearchResponse wraps search results with metadata used by the UI.
type SearchResponse struct {
	Results        []types.Listing
	Mode           SearchMode
	Warning        string
	Err            error
	ProviderErrors []ProviderError
}

// ProviderError captures a failed provider call with classification metadata.
type ProviderError struct {
	Provider string
	Kind     ProviderErrorKind
	Err      error
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

// HasConfiguredProvider reports whether at least one provider has usable credentials.
func (c *Client) HasConfiguredProvider() bool {
	if c == nil {
		return false
	}
	for _, provider := range c.providers {
		if provider != nil && provider.Configured() {
			return true
		}
	}
	return false
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
			Results:        []types.Listing{},
			Mode:           SearchModeUnavailable,
			Err:            fmt.Errorf("search client not configured"),
			ProviderErrors: []ProviderError{},
		}
	}

	if !c.HasConfiguredProvider() {
		return SearchResponse{
			Results: []types.Listing{},
			Mode:    SearchModeUnavailable,
			Err: fmt.Errorf(
				"no live search providers configured; set BRAVE_API_KEY, TAVILY_API_KEY, or FIRECRAWL_API_KEY",
			),
			ProviderErrors: []ProviderError{},
		}
	}

	var successfulProviders int
	providerErrors := make([]ProviderError, 0, len(c.providers))
	var failedProviders []string
	var failedHints []string

	for _, provider := range c.providers {
		if provider == nil || !provider.Configured() {
			continue
		}

		results, err := provider.Search(ctx, q)
		if err != nil {
			name := providerName(provider)
			providerErrors = append(providerErrors, ProviderError{
				Provider: name,
				Kind:     classifyProviderError(err),
				Err:      err,
			})
			failedProviders = append(failedProviders, name)
			if hint := actionableProviderError(name, err); hint != "" {
				failedHints = append(failedHints, hint)
			}
			continue
		}

		successfulProviders++
		if len(results) > 0 {
			return SearchResponse{
				Results:        results,
				Mode:           SearchModeLive,
				Warning:        buildSearchWarning(failedProviders, failedHints),
				ProviderErrors: providerErrors,
			}
		}
	}

	warning := buildSearchWarning(failedProviders, failedHints)
	if successfulProviders > 0 {
		return SearchResponse{
			Results:        []types.Listing{},
			Mode:           SearchModeLive,
			Warning:        warning,
			ProviderErrors: providerErrors,
		}
	}

	return SearchResponse{
		Results:        []types.Listing{},
		Mode:           SearchModeUnavailable,
		Warning:        warning,
		Err:            buildSearchError(warning, providerErrors),
		ProviderErrors: providerErrors,
	}
}

func providerName(provider SearchProvider) string {
	name := strings.TrimSpace(provider.Name())
	if name == "" {
		return "Provider"
	}
	return name
}

func classifyProviderError(err error) ProviderErrorKind {
	if err == nil {
		return ProviderErrorUnknown
	}
	if errors.Is(err, context.Canceled) {
		return ProviderErrorCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ProviderErrorTimeout
	}

	var statusErr *HTTPStatusError
	if errors.As(err, &statusErr) {
		switch {
		case statusErr.Status == http.StatusUnauthorized || statusErr.Status == http.StatusForbidden:
			return ProviderErrorAuth
		case statusErr.Status == http.StatusTooManyRequests:
			return ProviderErrorRateLimit
		default:
			return ProviderErrorHTTP
		}
	}

	return ProviderErrorTransport
}

func buildSearchWarning(failedProviders, failedHints []string) string {
	if len(failedHints) > 0 {
		return strings.Join(failedHints, " ")
	}
	if len(failedProviders) > 0 {
		return fmt.Sprintf("Live search unavailable (%s).", strings.Join(failedProviders, ", "))
	}
	return ""
}

func buildSearchError(warning string, providerErrors []ProviderError) error {
	root := primaryProviderError(providerErrors)
	if root == nil {
		if warning == "" {
			warning = "Live search unavailable."
		}
		return errors.New(warning)
	}
	if warning == "" {
		return root
	}
	return fmt.Errorf("%s: %w", warning, root)
}

func primaryProviderError(providerErrors []ProviderError) error {
	for _, providerErr := range providerErrors {
		if errors.Is(providerErr.Err, context.Canceled) {
			return context.Canceled
		}
	}
	for _, providerErr := range providerErrors {
		if errors.Is(providerErr.Err, context.DeadlineExceeded) {
			return context.DeadlineExceeded
		}
	}
	for _, providerErr := range providerErrors {
		if providerErr.Err != nil {
			return providerErr.Err
		}
	}
	return nil
}

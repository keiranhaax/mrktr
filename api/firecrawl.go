package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"mrktr/types"
)

// FirecrawlProvider implements SearchProvider via Firecrawl.
type FirecrawlProvider struct {
	apiKey    string
	searchURL string
	client    *http.Client
}

// NewFirecrawlProvider creates a Firecrawl provider.
func NewFirecrawlProvider(apiKey, searchURL string, client *http.Client) *FirecrawlProvider {
	if strings.TrimSpace(searchURL) == "" {
		searchURL = DefaultFirecrawlSearchURL
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &FirecrawlProvider{
		apiKey:    strings.TrimSpace(apiKey),
		searchURL: searchURL,
		client:    client,
	}
}

func (p *FirecrawlProvider) Name() string {
	return "Firecrawl"
}

func (p *FirecrawlProvider) Configured() bool {
	return p != nil && p.apiKey != ""
}

func (p *FirecrawlProvider) Search(ctx context.Context, query string) ([]types.Listing, error) {
	if !p.Configured() {
		return nil, fmt.Errorf("FIRECRAWL_API_KEY not set")
	}

	searchQuery := fmt.Sprintf("%s price site:ebay.com OR site:mercari.com OR site:amazon.com", query)

	reqBody := map[string]any{
		"query": searchQuery,
		"limit": 20,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal firecrawl request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.searchURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create firecrawl request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request firecrawl: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read firecrawl response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"firecrawl status %d: %s",
			resp.StatusCode,
			summarizeHTTPBody(body),
		)
	}

	var result struct {
		Data []SearchResult `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode firecrawl response: %w", err)
	}

	return ParseSearchResults(result.Data), nil
}

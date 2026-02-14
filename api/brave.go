package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mrktr/types"
)

// BraveProvider implements SearchProvider via Brave Search.
type BraveProvider struct {
	apiKey    string
	searchURL string
	client    *http.Client
}

// NewBraveProvider creates a Brave provider.
func NewBraveProvider(apiKey, searchURL string, client *http.Client) *BraveProvider {
	if strings.TrimSpace(searchURL) == "" {
		searchURL = DefaultBraveSearchURL
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	return &BraveProvider{
		apiKey:    strings.TrimSpace(apiKey),
		searchURL: searchURL,
		client:    client,
	}
}

func (p *BraveProvider) Name() string {
	return "Brave"
}

func (p *BraveProvider) Configured() bool {
	return p != nil && p.apiKey != ""
}

func (p *BraveProvider) Search(query string) ([]types.Listing, error) {
	if !p.Configured() {
		return nil, fmt.Errorf("BRAVE_API_KEY not set")
	}

	searchQuery := fmt.Sprintf(
		"%s price (site:ebay.com OR site:mercari.com OR site:amazon.com)",
		query,
	)

	searchURL, err := url.Parse(p.searchURL)
	if err != nil {
		return nil, fmt.Errorf("parse brave search URL: %w", err)
	}

	params := searchURL.Query()
	params.Set("q", searchQuery)
	params.Set("count", "20")
	searchURL.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, searchURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create brave request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Subscription-Token", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request brave: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read brave response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"brave status %d: %s",
			resp.StatusCode,
			summarizeHTTPBody(body),
		)
	}

	var result struct {
		Web struct {
			Results []SearchResult `json:"results"`
		} `json:"web"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode brave response: %w", err)
	}

	return ParseSearchResults(result.Web.Results), nil
}

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

// TavilyProvider implements SearchProvider via Tavily.
type TavilyProvider struct {
	apiKey    string
	searchURL string
	client    *http.Client
}

// NewTavilyProvider creates a Tavily provider.
func NewTavilyProvider(apiKey, searchURL string, client *http.Client) *TavilyProvider {
	if strings.TrimSpace(searchURL) == "" {
		searchURL = DefaultTavilySearchURL
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &TavilyProvider{
		apiKey:    strings.TrimSpace(apiKey),
		searchURL: searchURL,
		client:    client,
	}
}

func (p *TavilyProvider) Name() string {
	return "Tavily"
}

func (p *TavilyProvider) Configured() bool {
	return p != nil && p.apiKey != ""
}

func (p *TavilyProvider) Search(ctx context.Context, query string) ([]types.Listing, error) {
	if !p.Configured() {
		return nil, fmt.Errorf("TAVILY_API_KEY not set")
	}

	searchQuery := fmt.Sprintf("%s price ebay OR mercari OR amazon", query)

	reqBody := map[string]any{
		"api_key":     p.apiKey,
		"query":       searchQuery,
		"max_results": 20,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal tavily request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.searchURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create tavily request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request tavily: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read tavily response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &HTTPStatusError{
			Provider: "Tavily",
			Status:   resp.StatusCode,
			Body:     summarizeHTTPBody(body),
		}
	}

	var result struct {
		Results []struct {
			URL     string `json:"url"`
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode tavily response: %w", err)
	}

	data := make([]SearchResult, len(result.Results))
	for i, r := range result.Results {
		data[i].URL = r.URL
		data[i].Title = r.Title
		data[i].Description = r.Content
	}

	return ParseSearchResults(data), nil
}

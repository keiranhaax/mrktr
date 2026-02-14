package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSearchTavilyChecksHTTPStatus(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Body:       io.NopCloser(strings.NewReader("bad gateway")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	provider := NewTavilyProvider("tavily-key", "https://tavily.test/search", client)

	_, err := provider.Search(context.Background(), "ps5")
	if err == nil {
		t.Fatal("expected error for non-2xx tavily response")
	}
	if !strings.Contains(err.Error(), "status 502") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}

func TestSearchTavilyBuildsRequestAndParsesResults(t *testing.T) {
	var gotPath string
	var gotMethod string
	var gotContentType string
	var gotBody map[string]any

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotPath = req.URL.Path
			gotMethod = req.Method
			gotContentType = req.Header.Get("Content-Type")

			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(bodyBytes, &gotBody); err != nil {
				return nil, err
			}

			body := `{
				"results": [
					{
						"url": "https://ebay.com/itm/123",
						"title": "Nintendo Switch OLED",
						"content": "Excellent condition for $249.99"
					}
				]
			}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	provider := NewTavilyProvider("tavily-key", "https://tavily.test/search", client)

	results, err := provider.Search(context.Background(), "switch")
	if err != nil {
		t.Fatalf("expected successful tavily search, got %v", err)
	}

	if gotPath != "/search" {
		t.Fatalf("expected path %q, got %q", "/search", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("expected %s request, got %s", http.MethodPost, gotMethod)
	}
	if gotContentType != "application/json" {
		t.Fatalf("expected application/json content-type, got %q", gotContentType)
	}
	if gotBody["api_key"] != "tavily-key" {
		t.Fatalf("expected api_key in request body")
	}
	queryVal, ok := gotBody["query"].(string)
	if !ok {
		t.Fatalf("expected query string in request body, got %T", gotBody["query"])
	}
	if !strings.Contains(queryVal, "ebay OR mercari OR amazon") {
		t.Fatalf("expected marketplace filters in query, got %q", queryVal)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 parsed listing, got %d", len(results))
	}
	if results[0].Platform != "eBay" {
		t.Fatalf("expected eBay platform parsing, got %q", results[0].Platform)
	}
	if results[0].Price != 249.99 {
		t.Fatalf("expected parsed price 249.99, got %.2f", results[0].Price)
	}
}

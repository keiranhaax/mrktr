package api

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSearchBraveChecksHTTPStatus(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(strings.NewReader("too many requests")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	provider := NewBraveProvider("brave-key", "https://brave.test/res/v1/web/search", client)

	_, err := provider.Search(context.Background(), "ps5")
	if err == nil {
		t.Fatal("expected error for non-2xx brave response")
	}
	if !strings.Contains(err.Error(), "status 429") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}

func TestSearchBraveBuildsQueryAndParsesResults(t *testing.T) {
	var gotPath string
	var gotQuery string
	var gotToken string
	var gotEncoding string

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotPath = req.URL.Path
			gotQuery = req.URL.Query().Get("q")
			gotToken = req.Header.Get("X-Subscription-Token")
			gotEncoding = req.Header.Get("Accept-Encoding")

			body := `{
				"web": {
					"results": [
						{
							"url": "https://ebay.com/itm/123",
							"title": "Nintendo Switch OLED - $299.99",
							"description": "Used condition"
						}
					]
				}
			}`

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	provider := NewBraveProvider("brave-key", "https://brave.test/res/v1/web/search", client)

	results, err := provider.Search(context.Background(), "switch")
	if err != nil {
		t.Fatalf("expected successful brave search, got %v", err)
	}

	if gotPath != "/res/v1/web/search" {
		t.Fatalf("expected path %q, got %q", "/res/v1/web/search", gotPath)
	}
	if !strings.Contains(gotQuery, "site:ebay.com") {
		t.Fatalf("expected marketplace site filters in query, got %q", gotQuery)
	}
	if gotToken != "brave-key" {
		t.Fatalf("expected subscription token header, got %q", gotToken)
	}
	if gotEncoding != "" {
		t.Fatalf("expected Accept-Encoding header to be unset, got %q", gotEncoding)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 parsed listing, got %d", len(results))
	}
	if results[0].Platform != "eBay" {
		t.Fatalf("expected eBay platform parsing, got %q", results[0].Platform)
	}
	if results[0].Price != 299.99 {
		t.Fatalf("expected parsed price 299.99, got %.2f", results[0].Price)
	}
}

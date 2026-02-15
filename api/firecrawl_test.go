package api

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestSearchFirecrawlChecksHTTPStatus(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("server exploded")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	provider := NewFirecrawlProvider("firecrawl-key", "https://firecrawl.test/v1/search", client)

	_, err := provider.Search(context.Background(), "ps5")
	if err == nil {
		t.Fatal("expected error for non-2xx firecrawl response")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}

func TestSearchFirecrawlParsesSuccessResponse(t *testing.T) {
	var gotAuth string
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotAuth = req.Header.Get("Authorization")
			body := `{
				"data": [
					{
						"url": "https://ebay.com/itm/123",
						"title": "Nintendo Switch OLED",
						"description": "Used $249.99"
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

	provider := NewFirecrawlProvider("firecrawl-key", "https://firecrawl.test/v1/search", client)
	results, err := provider.Search(context.Background(), "switch")
	if err != nil {
		t.Fatalf("expected successful firecrawl response, got %v", err)
	}
	if gotAuth != "Bearer firecrawl-key" {
		t.Fatalf("expected bearer auth header, got %q", gotAuth)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 parsed listing, got %d", len(results))
	}
	if results[0].Platform != "eBay" || results[0].Price != 249.99 {
		t.Fatalf("unexpected parsed listing: %+v", results[0])
	}
}

func TestSearchFirecrawlHandlesEmptyResults(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"data":[]}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	provider := NewFirecrawlProvider("firecrawl-key", "https://firecrawl.test/v1/search", client)
	results, err := provider.Search(context.Background(), "switch")
	if err != nil {
		t.Fatalf("expected no error for empty data, got %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearchFirecrawlMalformedJSON(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"data":`)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	provider := NewFirecrawlProvider("firecrawl-key", "https://firecrawl.test/v1/search", client)
	_, err := provider.Search(context.Background(), "switch")
	if err == nil {
		t.Fatal("expected decode error for malformed json")
	}
	if !strings.Contains(err.Error(), "decode firecrawl response") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestSearchFirecrawlContextCanceled(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, req.Context().Err()
		}),
	}

	provider := NewFirecrawlProvider("firecrawl-key", "https://firecrawl.test/v1/search", client)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := provider.Search(ctx, "switch")
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
	if !strings.Contains(err.Error(), "request firecrawl") {
		t.Fatalf("expected request firecrawl context error, got %v", err)
	}
}

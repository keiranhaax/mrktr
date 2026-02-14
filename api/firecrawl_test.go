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

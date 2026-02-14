package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mrktr/types"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type searchMode string

const (
	searchModeLive     searchMode = "live"
	searchModeDemo     searchMode = "demo"
	searchModeFallback searchMode = "fallback"
)

type SearchResponse struct {
	Results []types.Listing
	Mode    searchMode
	Warning string
	Err     error
}

var (
	httpClient         = &http.Client{Timeout: 30 * time.Second}
	firecrawlSearchURL = "https://api.firecrawl.dev/v1/search"
	tavilySearchURL    = "https://api.tavily.com/search"
	pricePattern       = regexp.MustCompile(`\$(\d{1,3}(?:,\d{3})*(?:\.\d{2})?)`)
)

// SearchPrices searches for item prices across marketplaces
func SearchPrices(query string) SearchResponse {
	firecrawlConfigured := os.Getenv("FIRECRAWL_API_KEY") != ""
	tavilyConfigured := os.Getenv("TAVILY_API_KEY") != ""

	if !firecrawlConfigured && !tavilyConfigured {
		return SearchResponse{
			Results: mockSearch(query),
			Mode:    searchModeDemo,
		}
	}

	var successfulProviders int
	var failedProviders []string

	if firecrawlConfigured {
		results, err := searchFirecrawl(query)
		if err != nil {
			failedProviders = append(failedProviders, "Firecrawl")
		} else {
			successfulProviders++
			if len(results) > 0 {
				return SearchResponse{
					Results: results,
					Mode:    searchModeLive,
				}
			}
		}
	}

	if tavilyConfigured {
		results, err := searchTavily(query)
		if err != nil {
			failedProviders = append(failedProviders, "Tavily")
		} else {
			successfulProviders++
			if len(results) > 0 {
				return SearchResponse{
					Results: results,
					Mode:    searchModeLive,
				}
			}
		}
	}

	if successfulProviders > 0 {
		return SearchResponse{
			Results: []types.Listing{},
			Mode:    searchModeLive,
		}
	}

	warning := "Live search unavailable; showing demo data."
	if len(failedProviders) > 0 {
		warning = fmt.Sprintf(
			"Live search unavailable (%s); showing demo data.",
			strings.Join(failedProviders, ", "),
		)
	}

	return SearchResponse{
		Results: mockSearch(query),
		Mode:    searchModeFallback,
		Warning: warning,
	}
}

// searchFirecrawl uses Firecrawl API for search
func searchFirecrawl(query string) ([]types.Listing, error) {
	apiKey := os.Getenv("FIRECRAWL_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("FIRECRAWL_API_KEY not set")
	}

	searchQuery := fmt.Sprintf("%s price site:ebay.com OR site:mercari.com OR site:amazon.com", query)

	reqBody := map[string]interface{}{
		"query": searchQuery,
		"limit": 20,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal firecrawl request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, firecrawlSearchURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create firecrawl request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
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
		Data []struct {
			URL         string `json:"url"`
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode firecrawl response: %w", err)
	}

	return parseSearchResults(result.Data), nil
}

// searchTavily uses Tavily API for search
func searchTavily(query string) ([]types.Listing, error) {
	apiKey := os.Getenv("TAVILY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("TAVILY_API_KEY not set")
	}

	searchQuery := fmt.Sprintf("%s price ebay OR mercari OR amazon", query)

	reqBody := map[string]interface{}{
		"api_key":     apiKey,
		"query":       searchQuery,
		"max_results": 20,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal tavily request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, tavilySearchURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create tavily request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request tavily: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read tavily response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"tavily status %d: %s",
			resp.StatusCode,
			summarizeHTTPBody(body),
		)
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

	// Convert to our format
	data := make([]struct {
		URL         string `json:"url"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}, len(result.Results))

	for i, r := range result.Results {
		data[i].URL = r.URL
		data[i].Title = r.Title
		data[i].Description = r.Content
	}

	return parseSearchResults(data), nil
}

// parseSearchResults extracts listing data from search results
func parseSearchResults(data []struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}) []types.Listing {
	listings := make([]types.Listing, 0, len(data))

	for _, item := range data {
		listing := types.Listing{
			URL:   item.URL,
			Title: item.Title,
		}

		// Detect platform from URL
		urlLower := strings.ToLower(item.URL)
		switch {
		case strings.Contains(urlLower, "ebay.com"):
			listing.Platform = "eBay"
		case strings.Contains(urlLower, "mercari.com"):
			listing.Platform = "Mercari"
		case strings.Contains(urlLower, "amazon.com"):
			listing.Platform = "Amazon"
		case strings.Contains(urlLower, "facebook.com"):
			listing.Platform = "Facebook"
		default:
			listing.Platform = "Other"
		}

		// Extract price from title or description
		text := item.Title + " " + item.Description
		if matches := pricePattern.FindStringSubmatch(text); len(matches) > 1 {
			priceStr := strings.ReplaceAll(matches[1], ",", "")
			if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
				listing.Price = price
			}
		}

		// Skip if no price found
		if listing.Price == 0 {
			continue
		}

		// Detect condition
		textLower := strings.ToLower(text)
		switch {
		case strings.Contains(textLower, "new") || strings.Contains(textLower, "sealed"):
			listing.Condition = "New"
		case strings.Contains(textLower, "good"):
			listing.Condition = "Good"
		case strings.Contains(textLower, "fair"):
			listing.Condition = "Fair"
		default:
			listing.Condition = "Used"
		}

		// Detect status (sold vs active)
		if strings.Contains(textLower, "sold") {
			listing.Status = "Sold"
		} else {
			listing.Status = "Active"
		}

		listings = append(listings, listing)
	}

	return listings
}

func summarizeHTTPBody(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return "empty response body"
	}
	if len(trimmed) > 120 {
		return trimmed[:117] + "..."
	}
	return trimmed
}

// mockSearch returns demo data when no API keys are configured
func mockSearch(query string) []types.Listing {
	// Generate some realistic mock data based on the query
	basePrice := 100.0
	if strings.Contains(strings.ToLower(query), "iphone") {
		basePrice = 800.0
	} else if strings.Contains(strings.ToLower(query), "switch") {
		basePrice = 250.0
	} else if strings.Contains(strings.ToLower(query), "ps5") {
		basePrice = 400.0
	} else if strings.Contains(strings.ToLower(query), "airpods") {
		basePrice = 150.0
	}

	return []types.Listing{
		{Platform: "eBay", Price: basePrice * 0.95, Condition: "Used", Status: "Sold", URL: "https://ebay.com/1", Title: query + " - Great condition"},
		{Platform: "eBay", Price: basePrice * 1.05, Condition: "New", Status: "Active", URL: "https://ebay.com/2", Title: query + " - Brand New Sealed"},
		{Platform: "Mercari", Price: basePrice * 0.90, Condition: "Good", Status: "Active", URL: "https://mercari.com/1", Title: query + " - Good condition"},
		{Platform: "Mercari", Price: basePrice * 0.85, Condition: "Used", Status: "Sold", URL: "https://mercari.com/2", Title: query + " - Used but works"},
		{Platform: "Amazon", Price: basePrice * 1.15, Condition: "New", Status: "Active", URL: "https://amazon.com/1", Title: query + " - New"},
		{Platform: "eBay", Price: basePrice * 0.92, Condition: "Used", Status: "Sold", URL: "https://ebay.com/3", Title: query + " - Excellent"},
		{Platform: "Facebook", Price: basePrice * 0.80, Condition: "Fair", Status: "Active", URL: "https://facebook.com/1", Title: query + " - Fair condition"},
		{Platform: "eBay", Price: basePrice * 1.00, Condition: "Good", Status: "Active", URL: "https://ebay.com/4", Title: query + " - Like New"},
	}
}

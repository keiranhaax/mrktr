package api

import (
	"mrktr/types"
	"regexp"
	"strconv"
	"strings"
)

var pricePattern = regexp.MustCompile(`\$(\d{1,3}(?:,\d{3})*(?:\.\d{2})?)`)

// SearchResult normalizes provider payload fields for parsing.
type SearchResult struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ParseSearchResults extracts listing data from search results.
func ParseSearchResults(data []SearchResult) []types.Listing {
	listings := make([]types.Listing, 0, len(data))

	for _, item := range data {
		listing := types.Listing{
			URL:   item.URL,
			Title: item.Title,
		}

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

		text := item.Title + " " + item.Description
		if matches := pricePattern.FindStringSubmatch(text); len(matches) > 1 {
			priceStr := strings.ReplaceAll(matches[1], ",", "")
			if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
				listing.Price = price
			}
		}

		if listing.Price == 0 {
			continue
		}

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

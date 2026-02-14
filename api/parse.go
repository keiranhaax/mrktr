package api

import (
	"mrktr/types"
	"regexp"
	"strconv"
	"strings"
)

var (
	pricePattern           = regexp.MustCompile(`\$(\d{1,3}(?:,\d{3})+|\d+)(?:\.(\d{1,2}))?`)
	conditionNewPattern    = regexp.MustCompile(`\bnew\b`)
	conditionSealedPattern = regexp.MustCompile(`\bsealed\b`)
	conditionGoodPattern   = regexp.MustCompile(`\bgood\b`)
	conditionFairPattern   = regexp.MustCompile(`\bfair\b`)
	statusSoldPattern      = regexp.MustCompile(`\bsold\b`)
)

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
		if price, ok := extractBestPrice(text); ok {
			listing.Price = price
		}

		if listing.Price == 0 {
			continue
		}

		textLower := strings.ToLower(text)
		switch {
		case conditionNewPattern.MatchString(textLower) || conditionSealedPattern.MatchString(textLower):
			listing.Condition = "New"
		case conditionGoodPattern.MatchString(textLower):
			listing.Condition = "Good"
		case conditionFairPattern.MatchString(textLower):
			listing.Condition = "Fair"
		default:
			listing.Condition = "Used"
		}

		if statusSoldPattern.MatchString(textLower) {
			listing.Status = "Sold"
		} else {
			listing.Status = "Active"
		}

		listings = append(listings, listing)
	}

	return listings
}

// extractBestPrice returns the lowest positive USD amount found in the text.
// This helps pick current prices in snippets like "Was $150, now $99".
func extractBestPrice(text string) (float64, bool) {
	matches := pricePattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return 0, false
	}

	best := 0.0
	found := false

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		whole := strings.ReplaceAll(match[1], ",", "")
		if whole == "" {
			continue
		}

		priceStr := whole
		if len(match) > 2 && match[2] != "" {
			decimal := match[2]
			if len(decimal) == 1 {
				decimal += "0"
			}
			priceStr += "." + decimal
		}

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil || price <= 0 {
			continue
		}

		if !found || price < best {
			best = price
			found = true
		}
	}

	return best, found
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

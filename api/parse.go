package api

import (
	"mrktr/types"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	pricePatternSymbolPrefix = regexp.MustCompile(`(?i)(?:\busd\b|us\s*\$|us\$|\$)\s*(\d{1,3}(?:,\d{3})+|\d+)(?:\.(\d{1,2}))?`)
	pricePatternUSDSuffix    = regexp.MustCompile(`(?i)(\d{1,3}(?:,\d{3})+|\d+)(?:\.(\d{1,2}))?\s*\busd\b`)
	pricePatternContext      = regexp.MustCompile(`(?i)\b(?:price|asking|ask|obo|offer|now|for)\s*[:\-]?\s*(\d{1,3}(?:,\d{3})+|\d{2,})(?:\.(\d{1,2}))?\b`)
	conditionNewPattern      = regexp.MustCompile(`\bnew\b`)
	conditionSealedPattern   = regexp.MustCompile(`\bsealed\b`)
	conditionGoodPattern     = regexp.MustCompile(`\bgood\b`)
	conditionFairPattern     = regexp.MustCompile(`\bfair\b`)
	statusSoldPattern        = regexp.MustCompile(`\bsold\b`)
	statusUnsoldPattern      = regexp.MustCompile(`\b(?:not\s+sold|unsold|never\s+sold)\b`)
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

		listing.Platform = detectPlatform(item.URL)

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

		if statusUnsoldPattern.MatchString(textLower) {
			listing.Status = "Active"
		} else if statusSoldPattern.MatchString(textLower) {
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
	best := 0.0
	found := false

	patterns := []*regexp.Regexp{
		pricePatternSymbolPrefix,
		pricePatternUSDSuffix,
		pricePatternContext,
	}
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			price, ok := parsePriceMatch(match)
			if !ok {
				continue
			}
			if !found || price < best {
				best = price
				found = true
			}
		}
	}

	return best, found
}

func parsePriceMatch(match []string) (float64, bool) {
	if len(match) < 2 {
		return 0, false
	}

	whole := strings.ReplaceAll(match[1], ",", "")
	if whole == "" {
		return 0, false
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
		return 0, false
	}
	return price, true
}

func detectPlatform(rawURL string) string {
	host := ""
	if parsed, err := url.Parse(strings.TrimSpace(rawURL)); err == nil {
		host = strings.ToLower(parsed.Hostname())
	}
	if host == "" {
		host = strings.ToLower(rawURL)
	}

	switch {
	case hostHasLabel(host, "ebay"):
		return "eBay"
	case hostHasLabel(host, "mercari"):
		return "Mercari"
	case hostHasLabel(host, "amazon"):
		return "Amazon"
	case hostHasLabel(host, "facebook") || hostHasLabel(host, "fb"):
		return "Facebook"
	default:
		return "Other"
	}
}

func hostHasLabel(host, label string) bool {
	if host == "" || label == "" {
		return false
	}
	parts := strings.Split(strings.Trim(strings.ToLower(host), "."), ".")
	for _, part := range parts {
		if part == label {
			return true
		}
	}
	return false
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

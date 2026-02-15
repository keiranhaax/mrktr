package types

import "strings"

// ResultFilter constrains visible listings in the results panel.
type ResultFilter struct {
	Platform  string
	Condition string
	Status    string
}

// ApplyFilter returns only listings matching the configured filter values.
func ApplyFilter(in []Listing, f ResultFilter) []Listing {
	if len(in) == 0 {
		return nil
	}

	platform := normalizeFilterToken(f.Platform)
	condition := normalizeFilterToken(f.Condition)
	status := normalizeFilterToken(f.Status)

	out := make([]Listing, 0, len(in))
	for _, listing := range in {
		if platform != "" && !strings.EqualFold(listing.Platform, platform) {
			continue
		}
		if condition != "" && !conditionMatches(listing.Condition, condition) {
			continue
		}
		if status != "" && !strings.EqualFold(listing.Status, status) {
			continue
		}
		out = append(out, listing)
	}
	return out
}

func normalizeFilterToken(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.EqualFold(trimmed, "all") {
		return ""
	}
	return trimmed
}

func conditionMatches(condition, target string) bool {
	if strings.EqualFold(target, "Used") {
		switch strings.ToLower(strings.TrimSpace(condition)) {
		case "used", "good", "fair":
			return true
		default:
			return false
		}
	}
	return strings.EqualFold(condition, target)
}

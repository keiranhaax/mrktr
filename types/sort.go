package types

import (
	"sort"
	"strings"
)

// SortField selects which listing column to sort by.
type SortField string

const (
	SortFieldPrice     SortField = "price"
	SortFieldPlatform  SortField = "platform"
	SortFieldCondition SortField = "condition"
	SortFieldStatus    SortField = "status"
)

// SortDirection selects ascending or descending sort order.
type SortDirection string

const (
	SortDirectionAsc  SortDirection = "asc"
	SortDirectionDesc SortDirection = "desc"
)

// SortResults returns a sorted copy of input listings.
func SortResults(in []Listing, field SortField, dir SortDirection) []Listing {
	if len(in) <= 1 {
		return append([]Listing(nil), in...)
	}

	field = normalizeSortField(field)
	dir = normalizeSortDirection(dir)

	out := append([]Listing(nil), in...)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		cmp := compareListing(a, b, field)
		if cmp == 0 {
			return false
		}
		if dir == SortDirectionDesc {
			return cmp > 0
		}
		return cmp < 0
	})

	return out
}

func normalizeSortField(field SortField) SortField {
	switch strings.ToLower(strings.TrimSpace(string(field))) {
	case string(SortFieldPlatform):
		return SortFieldPlatform
	case string(SortFieldCondition):
		return SortFieldCondition
	case string(SortFieldStatus):
		return SortFieldStatus
	default:
		return SortFieldPrice
	}
}

func normalizeSortDirection(dir SortDirection) SortDirection {
	switch strings.ToLower(strings.TrimSpace(string(dir))) {
	case string(SortDirectionDesc):
		return SortDirectionDesc
	default:
		return SortDirectionAsc
	}
}

func compareListing(a, b Listing, field SortField) int {
	switch field {
	case SortFieldPlatform:
		return strings.Compare(strings.ToLower(a.Platform), strings.ToLower(b.Platform))
	case SortFieldCondition:
		return conditionSortRank(a.Condition) - conditionSortRank(b.Condition)
	case SortFieldStatus:
		return statusSortRank(a.Status) - statusSortRank(b.Status)
	default:
		switch {
		case a.Price < b.Price:
			return -1
		case a.Price > b.Price:
			return 1
		default:
			return 0
		}
	}
}

func conditionSortRank(condition string) int {
	switch strings.ToLower(strings.TrimSpace(condition)) {
	case "new":
		return 0
	case "used":
		return 1
	case "good":
		return 2
	case "fair":
		return 3
	default:
		return 4
	}
}

func statusSortRank(status string) int {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active":
		return 0
	case "sold":
		return 1
	default:
		return 2
	}
}

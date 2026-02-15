package types

import "testing"

func TestSortResults(t *testing.T) {
	input := []Listing{
		{Platform: "Mercari", Price: 110, Condition: "Used", Status: "Sold"},
		{Platform: "eBay", Price: 90, Condition: "New", Status: "Active"},
		{Platform: "Amazon", Price: 140, Condition: "Fair", Status: "Active"},
	}

	tests := []struct {
		name      string
		field     SortField
		dir       SortDirection
		wantOrder []string
	}{
		{
			name:      "price ascending",
			field:     SortFieldPrice,
			dir:       SortDirectionAsc,
			wantOrder: []string{"eBay", "Mercari", "Amazon"},
		},
		{
			name:      "price descending",
			field:     SortFieldPrice,
			dir:       SortDirectionDesc,
			wantOrder: []string{"Amazon", "Mercari", "eBay"},
		},
		{
			name:      "platform ascending",
			field:     SortFieldPlatform,
			dir:       SortDirectionAsc,
			wantOrder: []string{"Amazon", "eBay", "Mercari"},
		},
		{
			name:      "status ascending active before sold",
			field:     SortFieldStatus,
			dir:       SortDirectionAsc,
			wantOrder: []string{"eBay", "Amazon", "Mercari"},
		},
		{
			name:      "invalid field defaults to price asc",
			field:     SortField("unknown"),
			dir:       SortDirectionAsc,
			wantOrder: []string{"eBay", "Mercari", "Amazon"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SortResults(input, tc.field, tc.dir)
			if len(got) != len(tc.wantOrder) {
				t.Fatalf("expected %d results, got %d", len(tc.wantOrder), len(got))
			}
			for i := range tc.wantOrder {
				if got[i].Platform != tc.wantOrder[i] {
					t.Fatalf("index %d: expected platform %q, got %q", i, tc.wantOrder[i], got[i].Platform)
				}
			}
		})
	}
}

func TestSortResultsReturnsCopy(t *testing.T) {
	input := []Listing{
		{Platform: "Mercari", Price: 110},
		{Platform: "eBay", Price: 90},
	}

	got := SortResults(input, SortFieldPrice, SortDirectionAsc)
	if &got[0] == &input[0] {
		t.Fatal("expected sort to return a copied slice")
	}
	if input[0].Platform != "Mercari" {
		t.Fatalf("expected original input to remain unchanged, got %q", input[0].Platform)
	}
}

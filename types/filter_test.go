package types

import "testing"

func TestApplyFilter(t *testing.T) {
	in := []Listing{
		{Platform: "eBay", Condition: "New", Status: "Active", Price: 100},
		{Platform: "Mercari", Condition: "Used", Status: "Sold", Price: 80},
		{Platform: "eBay", Condition: "Good", Status: "Active", Price: 75},
	}

	tests := []struct {
		name string
		f    ResultFilter
		want int
	}{
		{
			name: "no filters returns all",
			f:    ResultFilter{},
			want: 3,
		},
		{
			name: "platform filter",
			f:    ResultFilter{Platform: "eBay"},
			want: 2,
		},
		{
			name: "condition new filter",
			f:    ResultFilter{Condition: "New"},
			want: 1,
		},
		{
			name: "used includes good and fair",
			f:    ResultFilter{Condition: "Used"},
			want: 2,
		},
		{
			name: "status filter",
			f:    ResultFilter{Status: "Sold"},
			want: 1,
		},
		{
			name: "combined filter",
			f:    ResultFilter{Platform: "eBay", Status: "Active"},
			want: 2,
		},
		{
			name: "all treated as unset",
			f:    ResultFilter{Platform: "all", Condition: "All", Status: "ALL"},
			want: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ApplyFilter(in, tc.f)
			if len(got) != tc.want {
				t.Fatalf("expected %d results, got %d", tc.want, len(got))
			}
		})
	}
}

func TestApplyFilterNoMatch(t *testing.T) {
	in := []Listing{
		{Platform: "eBay", Condition: "New", Status: "Active", Price: 100},
	}
	got := ApplyFilter(in, ResultFilter{Platform: "Amazon"})
	if len(got) != 0 {
		t.Fatalf("expected zero results, got %d", len(got))
	}
}

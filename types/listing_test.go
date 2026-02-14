package types

import "testing"

func TestCalculateStats(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		stats := CalculateStats(nil)
		if stats.Count != 0 || stats.Min != 0 || stats.Max != 0 || stats.Average != 0 || stats.Median != 0 {
			t.Fatalf("unexpected stats for empty input: %+v", stats)
		}
	})

	t.Run("single item", func(t *testing.T) {
		stats := CalculateStats([]Listing{{Price: 99.5}})
		if stats.Count != 1 {
			t.Fatalf("expected count 1, got %d", stats.Count)
		}
		if stats.Min != 99.5 || stats.Max != 99.5 || stats.Average != 99.5 || stats.Median != 99.5 {
			t.Fatalf("unexpected stats for single item: %+v", stats)
		}
	})

	t.Run("odd and even median", func(t *testing.T) {
		odd := CalculateStats([]Listing{
			{Price: 40},
			{Price: 10},
			{Price: 20},
		})
		if odd.Median != 20 {
			t.Fatalf("expected odd median 20, got %v", odd.Median)
		}

		even := CalculateStats([]Listing{
			{Price: 40},
			{Price: 10},
			{Price: 20},
			{Price: 30},
		})
		if even.Median != 25 {
			t.Fatalf("expected even median 25, got %v", even.Median)
		}
	})
}

func TestCalculateProfit(t *testing.T) {
	tests := []struct {
		name        string
		cost        float64
		sell        float64
		wantProfit  float64
		wantPercent float64
	}{
		{
			name:        "profit",
			cost:        100,
			sell:        125,
			wantProfit:  25,
			wantPercent: 25,
		},
		{
			name:        "loss",
			cost:        100,
			sell:        70,
			wantProfit:  -30,
			wantPercent: -30,
		},
		{
			name:        "zero cost",
			cost:        0,
			sell:        50,
			wantProfit:  50,
			wantPercent: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateProfit(tc.cost, tc.sell)
			if got.Profit != tc.wantProfit {
				t.Fatalf("expected profit %v, got %v", tc.wantProfit, got.Profit)
			}
			if got.ProfitPercent != tc.wantPercent {
				t.Fatalf("expected percent %v, got %v", tc.wantPercent, got.ProfitPercent)
			}
		})
	}
}

package idea

import (
	"math"
	"mrktr/types"
	"testing"
)

func TestCalculateExtendedStatsEmpty(t *testing.T) {
	got := CalculateExtendedStats(nil)

	if got.Count != 0 {
		t.Fatalf("expected count 0, got %d", got.Count)
	}
	if got.Spread != "N/A" {
		t.Fatalf("expected spread N/A for empty input, got %q", got.Spread)
	}
	if got.PlatformStats == nil || got.ConditionStats == nil {
		t.Fatal("expected stats maps to be initialized")
	}
	if len(got.PlatformStats) != 0 || len(got.ConditionStats) != 0 {
		t.Fatalf("expected empty maps, got platforms=%d conditions=%d", len(got.PlatformStats), len(got.ConditionStats))
	}
	if len(got.Histogram) != 0 {
		t.Fatalf("expected no histogram bins for empty input, got %d", len(got.Histogram))
	}
}

func TestCalculateExtendedStatsAggregates(t *testing.T) {
	listings := []types.Listing{
		{Platform: "eBay", Condition: "New", Status: "Sold", Price: 100},
		{Platform: "eBay", Condition: "Used", Status: "Active", Price: 200},
		{Platform: "Mercari", Condition: "Good", Status: "Sold", Price: 300},
		{Platform: "Amazon", Condition: "New", Status: "Active", Price: 400},
	}

	got := CalculateExtendedStats(listings)

	if got.Count != 4 || got.Min != 100 || got.Max != 400 || got.Average != 250 || got.Median != 250 {
		t.Fatalf("unexpected base stats: %+v", got.Statistics)
	}

	assertFloatNear(t, got.P10, 130, 1e-9, "P10")
	assertFloatNear(t, got.P25, 175, 1e-9, "P25")
	assertFloatNear(t, got.P75, 325, 1e-9, "P75")
	assertFloatNear(t, got.P90, 370, 1e-9, "P90")
	assertFloatNear(t, got.StdDev, 111.80339887498948, 1e-9, "StdDev")
	assertFloatNear(t, got.CoV, 0.4472135954999579, 1e-9, "CoV")
	if got.Spread != "Wide" {
		t.Fatalf("expected spread Wide, got %q", got.Spread)
	}

	if got.SoldCount != 2 || got.ActiveCount != 2 {
		t.Fatalf("unexpected sold/active counts: sold=%d active=%d", got.SoldCount, got.ActiveCount)
	}
	assertFloatNear(t, got.SoldAvg, 200, 1e-9, "SoldAvg")
	assertFloatNear(t, got.ActiveAvg, 300, 1e-9, "ActiveAvg")

	eBay := got.PlatformStats["eBay"]
	if eBay.Count != 2 {
		t.Fatalf("expected eBay count 2, got %d", eBay.Count)
	}
	assertFloatNear(t, eBay.Average, 150, 1e-9, "eBay average")
	assertFloatNear(t, eBay.Min, 100, 1e-9, "eBay min")
	assertFloatNear(t, eBay.Max, 200, 1e-9, "eBay max")

	newCond := got.ConditionStats["New"]
	if newCond.Count != 2 {
		t.Fatalf("expected New count 2, got %d", newCond.Count)
	}
	assertFloatNear(t, newCond.Average, 250, 1e-9, "New average")

	if len(got.Histogram) == 0 {
		t.Fatal("expected histogram bins for non-empty input")
	}
}

func TestCalculateExtendedStatsNormalizesUnknownBuckets(t *testing.T) {
	listings := []types.Listing{
		{Platform: "", Condition: "", Status: "sold", Price: 100},
		{Platform: "  ", Condition: " ", Status: "active", Price: 200},
	}

	got := CalculateExtendedStats(listings)

	unknownPlatform, ok := got.PlatformStats["Unknown"]
	if !ok {
		t.Fatalf("expected Unknown platform bucket, got %#v", got.PlatformStats)
	}
	if unknownPlatform.Count != 2 {
		t.Fatalf("expected Unknown platform count 2, got %d", unknownPlatform.Count)
	}

	unknownCondition, ok := got.ConditionStats["Unknown"]
	if !ok {
		t.Fatalf("expected Unknown condition bucket, got %#v", got.ConditionStats)
	}
	if unknownCondition.Count != 2 {
		t.Fatalf("expected Unknown condition count 2, got %d", unknownCondition.Count)
	}
}

func TestCalculatePercentile(t *testing.T) {
	sorted := []float64{10, 20, 30, 40}

	tests := []struct {
		name string
		p    float64
		want float64
	}{
		{name: "p0", p: 0, want: 10},
		{name: "p10", p: 0.10, want: 13},
		{name: "p25", p: 0.25, want: 17.5},
		{name: "p50", p: 0.50, want: 25},
		{name: "p75", p: 0.75, want: 32.5},
		{name: "p100", p: 1, want: 40},
		{name: "belowRange", p: -0.5, want: 10},
		{name: "aboveRange", p: 1.5, want: 40},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calculatePercentile(sorted, tc.p)
			assertFloatNear(t, got, tc.want, 1e-9, tc.name)
		})
	}
}

func TestClassifySpread(t *testing.T) {
	tests := []struct {
		name string
		cov  float64
		want string
	}{
		{name: "tight", cov: 0.10, want: "Tight"},
		{name: "moderateLowerBound", cov: 0.15, want: "Moderate"},
		{name: "moderate", cov: 0.34, want: "Moderate"},
		{name: "wideLowerBound", cov: 0.35, want: "Wide"},
		{name: "wide", cov: 0.70, want: "Wide"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := classifySpread(tc.cov)
			if got != tc.want {
				t.Fatalf("expected %q for cov %.2f, got %q", tc.want, tc.cov, got)
			}
		})
	}
}

func TestCalculateHistogramBins(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := calculateHistogramBins(nil)
		if len(got) != 0 {
			t.Fatalf("expected empty bins, got %d", len(got))
		}
	})

	t.Run("single", func(t *testing.T) {
		got := calculateHistogramBins([]float64{42})
		if len(got) != 1 {
			t.Fatalf("expected 1 bin, got %d", len(got))
		}
		if got[0].Count != 1 {
			t.Fatalf("expected single bin count 1, got %d", got[0].Count)
		}
	})

	t.Run("allSamePrice", func(t *testing.T) {
		got := calculateHistogramBins([]float64{99, 99, 99, 99})
		if len(got) != 1 {
			t.Fatalf("expected 1 bin for uniform data, got %d", len(got))
		}
		if got[0].Count != 4 {
			t.Fatalf("expected uniform bin count 4, got %d", got[0].Count)
		}
	})

	t.Run("totalCountPreserved", func(t *testing.T) {
		input := []float64{100, 120, 140, 220, 260, 400, 1000}
		got := calculateHistogramBins(input)
		if len(got) == 0 {
			t.Fatal("expected at least one bin")
		}
		total := 0
		for _, bin := range got {
			total += bin.Count
		}
		if total != len(input) {
			t.Fatalf("expected total count %d, got %d", len(input), total)
		}
	})
}

func assertFloatNear(t *testing.T, got, want, tol float64, label string) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Fatalf("%s: expected %.12f, got %.12f", label, want, got)
	}
}

package idea

import (
	"mrktr/types"
	"sort"
	"strings"
)

// StatsViewMode controls which statistics view is active in the UI.
type StatsViewMode int

const (
	StatsViewSummary StatsViewMode = iota
	StatsViewDistribution
	StatsViewMarket
)

type ExtendedStatistics struct {
	types.Statistics

	StdDev float64
	P10    float64
	P25    float64
	P75    float64
	P90    float64
	Spread string
	CoV    float64

	PlatformStats  map[string]PlatformStat
	ConditionStats map[string]ConditionStat

	SoldCount   int
	ActiveCount int
	SoldAvg     float64
	ActiveAvg   float64

	Histogram []HistogramBin
}

type PlatformStat struct {
	Count   int
	Average float64
	Min     float64
	Max     float64
}

type ConditionStat struct {
	Count   int
	Average float64
}

type HistogramBin struct {
	Label    string
	Count    int
	MinPrice float64
	MaxPrice float64
}

type runningPlatformStat struct {
	Count int
	Sum   float64
	Min   float64
	Max   float64
}

type runningConditionStat struct {
	Count int
	Sum   float64
}

func CalculateExtendedStats(listings []types.Listing) ExtendedStatistics {
	stats := ExtendedStatistics{
		Statistics:     types.CalculateStats(listings),
		Spread:         "N/A",
		PlatformStats:  map[string]PlatformStat{},
		ConditionStats: map[string]ConditionStat{},
	}
	if len(listings) == 0 {
		return stats
	}

	prices := make([]float64, len(listings))
	runningPlatforms := map[string]runningPlatformStat{}
	runningConditions := map[string]runningConditionStat{}

	var soldSum float64
	var activeSum float64

	for i, listing := range listings {
		price := listing.Price
		prices[i] = price

		platform := normalizeBucketKey(listing.Platform, "Unknown")
		p := runningPlatforms[platform]
		if p.Count == 0 {
			p.Min = price
			p.Max = price
		} else {
			if price < p.Min {
				p.Min = price
			}
			if price > p.Max {
				p.Max = price
			}
		}
		p.Count++
		p.Sum += price
		runningPlatforms[platform] = p

		condition := normalizeBucketKey(listing.Condition, "Unknown")
		c := runningConditions[condition]
		c.Count++
		c.Sum += price
		runningConditions[condition] = c

		if strings.EqualFold(strings.TrimSpace(listing.Status), "sold") {
			stats.SoldCount++
			soldSum += price
		} else {
			stats.ActiveCount++
			activeSum += price
		}
	}

	if stats.SoldCount > 0 {
		stats.SoldAvg = soldSum / float64(stats.SoldCount)
	}
	if stats.ActiveCount > 0 {
		stats.ActiveAvg = activeSum / float64(stats.ActiveCount)
	}

	for name, p := range runningPlatforms {
		avg := 0.0
		if p.Count > 0 {
			avg = p.Sum / float64(p.Count)
		}
		stats.PlatformStats[name] = PlatformStat{
			Count:   p.Count,
			Average: avg,
			Min:     p.Min,
			Max:     p.Max,
		}
	}

	for name, c := range runningConditions {
		avg := 0.0
		if c.Count > 0 {
			avg = c.Sum / float64(c.Count)
		}
		stats.ConditionStats[name] = ConditionStat{
			Count:   c.Count,
			Average: avg,
		}
	}

	sort.Float64s(prices)
	stats.P10 = calculatePercentile(prices, 0.10)
	stats.P25 = calculatePercentile(prices, 0.25)
	stats.P75 = calculatePercentile(prices, 0.75)
	stats.P90 = calculatePercentile(prices, 0.90)

	stats.StdDev = calculateStdDev(prices, stats.Average)
	stats.CoV = calculateCoV(stats.StdDev, stats.Average)
	stats.Spread = classifySpread(stats.CoV)
	stats.Histogram = calculateHistogramBins(prices)

	return stats
}

func normalizeBucketKey(raw, fallback string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

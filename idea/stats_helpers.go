package idea

import (
	"fmt"
	"math"
	"sort"
)

func calculatePercentile(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[n-1]
	}

	rank := p * float64(n-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))
	if lower == upper {
		return sorted[lower]
	}

	weight := rank - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var squaredDiffSum float64
	for _, value := range values {
		diff := value - mean
		squaredDiffSum += diff * diff
	}
	return math.Sqrt(squaredDiffSum / float64(len(values)))
}

func calculateCoV(stdDev, mean float64) float64 {
	if mean <= 0 {
		return 0
	}
	return stdDev / mean
}

func classifySpread(cov float64) string {
	switch {
	case cov < 0.15:
		return "Tight"
	case cov < 0.35:
		return "Moderate"
	default:
		return "Wide"
	}
}

func calculateHistogramBins(prices []float64) []HistogramBin {
	if len(prices) == 0 {
		return nil
	}

	sortedPrices := append([]float64(nil), prices...)
	sort.Float64s(sortedPrices)

	minPrice := sortedPrices[0]
	maxPrice := sortedPrices[len(sortedPrices)-1]
	if minPrice == maxPrice {
		return []HistogramBin{
			{
				Label:    formatHistogramLabel(minPrice, maxPrice),
				Count:    len(sortedPrices),
				MinPrice: minPrice,
				MaxPrice: maxPrice,
			},
		}
	}

	binCount := sturgesBinCount(len(sortedPrices))
	if binCount > len(sortedPrices) {
		binCount = len(sortedPrices)
	}
	if binCount < 1 {
		binCount = 1
	}

	width := (maxPrice - minPrice) / float64(binCount)
	bins := make([]HistogramBin, binCount)

	for i := 0; i < binCount; i++ {
		start := minPrice + float64(i)*width
		end := start + width
		if i == binCount-1 {
			end = maxPrice
		}
		bins[i] = HistogramBin{
			Label:    formatHistogramLabel(start, end),
			MinPrice: start,
			MaxPrice: end,
		}
	}

	for _, price := range sortedPrices {
		idx := int((price - minPrice) / width)
		if idx >= binCount {
			idx = binCount - 1
		}
		bins[idx].Count++
	}

	return bins
}

func sturgesBinCount(n int) int {
	if n <= 1 {
		return 1
	}
	return int(math.Ceil(1 + 3.322*math.Log10(float64(n))))
}

func formatHistogramLabel(minPrice, maxPrice float64) string {
	left := math.Floor(minPrice)
	right := math.Ceil(maxPrice)
	if right < left {
		right = left
	}
	return fmt.Sprintf("$%.0f-$%.0f", left, right)
}

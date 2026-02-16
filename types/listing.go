package types

import (
	"math"
	"sort"
)

// Listing represents a single price listing from a marketplace
type Listing struct {
	Platform  string  // "eBay", "Mercari", "Amazon", "Facebook"
	Price     float64 // Price in USD
	Condition string  // "New", "Used", "Good", "Fair"
	Status    string  // "Sold", "Active"
	URL       string  // Link to the listing
	Title     string  // Item title/description
}

// Statistics holds calculated price statistics
type Statistics struct {
	Count   int
	Min     float64
	Max     float64
	Average float64
	Median  float64
}

// CalculateStats computes statistics from a slice of listings
func CalculateStats(listings []Listing) Statistics {
	if len(listings) == 0 {
		return Statistics{}
	}

	prices := make([]float64, len(listings))
	var sum float64

	for i, l := range listings {
		prices[i] = l.Price
		sum += l.Price
	}

	sort.Float64s(prices)

	stats := Statistics{
		Count:   len(listings),
		Min:     prices[0],
		Max:     prices[len(prices)-1],
		Average: sum / float64(len(prices)),
	}

	// Calculate median
	mid := len(prices) / 2
	if len(prices)%2 == 0 {
		stats.Median = (prices[mid-1] + prices[mid]) / 2
	} else {
		stats.Median = prices[mid]
	}

	return stats
}

// ProfitCalculation holds profit margin calculations
type ProfitCalculation struct {
	Cost          float64
	SellPrice     float64
	Profit        float64
	ProfitPercent float64
}

// CalculateProfit computes profit metrics for a given cost and sell price
func CalculateProfit(cost, sellPrice float64) ProfitCalculation {
	profit := sellPrice - cost
	var profitPercent float64
	if cost > 0 {
		profitPercent = (profit / cost) * 100
	} else if profit > 0 {
		profitPercent = math.Inf(1)
	} else if profit < 0 {
		profitPercent = math.Inf(-1)
	}
	return ProfitCalculation{
		Cost:          cost,
		SellPrice:     sellPrice,
		Profit:        profit,
		ProfitPercent: profitPercent,
	}
}

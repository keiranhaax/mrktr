package idea

import "fmt"

// RenderSummaryBody renders the summary-tab lines below the tab bar.
// The caller controls reveal animation and panel framing.
func RenderSummaryBody(stats ExtendedStatistics, sparkline string, width int) []string {
	if width < 20 {
		width = 20
	}

	resultSpread := fmt.Sprintf("Results: %d  Spread: %s", stats.Count, RenderSpreadValue(stats.Spread))
	avgMed := fmt.Sprintf("Avg: %s  Med: %s", formatPrice(stats.Average), formatPrice(stats.Median))

	if width < 44 {
		return []string{
			resultSpread,
			"Trend: " + sparkline,
			fmt.Sprintf("Min: %s  Max: %s", formatPrice(stats.Min), formatPrice(stats.Max)),
			avgMed,
			fmt.Sprintf("P25: %s  P75: %s", formatPrice(stats.P25), formatPrice(stats.P75)),
		}
	}

	return []string{
		resultSpread,
		"Trend: " + sparkline,
		fmt.Sprintf("Min: %s   P25: %s", formatPrice(stats.Min), formatPrice(stats.P25)),
		fmt.Sprintf("Max: %s   P75: %s", formatPrice(stats.Max), formatPrice(stats.P75)),
		avgMed,
	}
}

func formatPrice(v float64) string {
	return fmt.Sprintf("$%.2f", v)
}

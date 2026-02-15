package idea

import "github.com/charmbracelet/lipgloss"

var (
	statsTabActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#EA80FC"))

	statsTabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#667085"))

	spreadTightStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#12B76A"))

	spreadModerateStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#F79009"))

	spreadWideStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F97066"))
)

func RenderStatsTabs(mode StatsViewMode) string {
	sum := "[1:Sum]"
	dist := "[2:Dist]"
	mkt := "[3:Mkt]"

	if mode == StatsViewSummary {
		sum = statsTabActiveStyle.Render(sum)
	} else {
		sum = statsTabInactiveStyle.Render(sum)
	}
	if mode == StatsViewDistribution {
		dist = statsTabActiveStyle.Render(dist)
	} else {
		dist = statsTabInactiveStyle.Render(dist)
	}
	if mode == StatsViewMarket {
		mkt = statsTabActiveStyle.Render(mkt)
	} else {
		mkt = statsTabInactiveStyle.Render(mkt)
	}

	return sum + " " + dist + " " + mkt
}

func RenderSpreadValue(spread string) string {
	switch spread {
	case "Tight":
		return spreadTightStyle.Render(spread)
	case "Moderate":
		return spreadModerateStyle.Render(spread)
	case "Wide":
		return spreadWideStyle.Render(spread)
	default:
		return statsTabInactiveStyle.Render(spread)
	}
}

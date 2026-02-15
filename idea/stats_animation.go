package idea

import (
	"strings"
)

var skeletonFrames = []string{
	"░▒▓█▓▒░",
	"▒▓█▓▒░░",
	"▓█▓▒░░▒",
	"█▓▒░░▒▓",
}

// RenderStatsSkeleton returns compact animated placeholder lines.
func RenderStatsSkeleton(frame int, width int) []string {
	if width < 16 {
		width = 16
	}
	if len(skeletonFrames) == 0 {
		return []string{"loading..."}
	}

	idx := frame % len(skeletonFrames)
	if idx < 0 {
		idx = 0
	}
	base := skeletonFrames[idx]

	fill := func(n int) string {
		if n <= 0 {
			return ""
		}
		repeats := n/len(base) + 1
		block := strings.Repeat(base, repeats)
		if len(block) > n {
			block = block[:n]
		}
		return block
	}

	long := width - 8
	mid := width - 12
	short := width - 16
	if short < 6 {
		short = 6
	}
	if mid < short {
		mid = short + 2
	}
	if long < mid {
		long = mid + 2
	}

	return []string{
		"Results: " + fill(long),
		"Trend:   " + fill(mid),
		"Range:   " + fill(short),
		"Average: " + fill(mid),
		"Status:  " + fill(short),
	}
}

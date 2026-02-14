package main

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Pre-rendered ASCII art for "m r k t r" (figlet starwars font).
// Each line is the same visual width - pure ASCII for consistent alignment.
var introArtLines = []string{
	`.___  ___.    .______          __  ___    .___________.   .______`,
	`|   \/   |    |   _  \        |  |/  /    |           |   |   _  \`,
	`|  \  /  |    |  |_)  |       |  '  /     ` + "`" + `---|  |----` + "`" + `   |  |_)  |`,
	`|  |\/|  |    |      /        |    <          |  |        |      /`,
	`|  |  |  |    |  |\  \----.   |  .  \         |  |        |  |\  \----.`,
	`|__|  |__|    | _| ` + "`" + `._____|   |__|\__\        |__|        | _| ` + "`" + `._____|`,
}

func (m Model) renderIntro() string {
	maxWidth := 0
	for _, line := range introArtLines {
		if w := len([]rune(line)); w > maxWidth {
			maxWidth = w
		}
	}
	artLines := make([]string, len(introArtLines))
	for i, line := range introArtLines {
		runes := []rune(line)
		if len(runes) < maxWidth {
			line += strings.Repeat(" ", maxWidth-len(runes))
		}
		artLines[i] = line
	}

	totalChars := 0
	for _, line := range artLines {
		totalChars += len([]rune(line))
	}

	var revealFraction float64
	var glowPosition float64
	var fadeOpacity float64

	switch m.intro.Phase {
	case 0: // Reveal phase
		revealFraction = float64(m.intro.Tick) / float64(introRevealTicks)
		if revealFraction > 1 {
			revealFraction = 1
		}
	case 1: // Glow sweep phase
		revealFraction = 1
		glowProgress := float64(m.intro.Tick-introRevealTicks) / float64(introGlowTicks)
		if glowProgress > 1 {
			glowProgress = 1
		}
		glowPosition = glowProgress
	case 2: // Fade out phase
		revealFraction = 1
		glowPosition = 1
		fadeProgress := float64(m.intro.Tick-introRevealTicks-introGlowTicks) / float64(introFadeTicks)
		if fadeProgress > 1 {
			fadeProgress = 1
		}
		fadeOpacity = fadeProgress
	}

	_ = glowPosition
	_ = fadeOpacity

	var rendered []string
	charIndex := 0

	for _, line := range artLines {
		runes := []rune(line)
		var lineBuilder strings.Builder

		for _, r := range runes {
			charProgress := float64(charIndex) / float64(max(1, totalChars))
			charIndex++

			if charProgress > revealFraction {
				lineBuilder.WriteRune(' ')
				continue
			}

			var color string
			t := charProgress

			if m.intro.Phase >= 1 {
				dist := math.Abs(t - glowPosition)
				glowWidth := 0.15
				if dist < glowWidth {
					glowIntensity := 1.0 - (dist / glowWidth)
					baseColor := interpolateHexColor("#7D56F4", "#EA80FC", t)
					color = interpolateHexColor(baseColor, "#FFFFFF", glowIntensity*0.7)
				} else {
					color = interpolateHexColor("#7D56F4", "#EA80FC", t)
				}
			} else {
				appearProgress := 0.0
				revealPoint := charProgress
				if revealFraction > revealPoint {
					appearProgress = math.Min(1.0, (revealFraction-revealPoint)*3.0)
				}
				targetColor := interpolateHexColor("#7D56F4", "#EA80FC", t)
				color = interpolateHexColor("#1D1D2E", targetColor, appearProgress)
			}

			if m.intro.Phase == 2 {
				color = interpolateHexColor(color, "#1D1D2E", fadeOpacity)
			}

			lineBuilder.WriteString(
				lipgloss.NewStyle().
					Foreground(lipgloss.Color(color)).
					Render(string(r)),
			)
		}

		rendered = append(rendered, lineBuilder.String())
	}

	subtitle := ""
	if m.intro.Phase >= 1 {
		subText := "reseller price research"
		subOpacity := 1.0
		if m.intro.Phase == 1 {
			subOpacity = float64(m.intro.Tick-introRevealTicks) / float64(introGlowTicks)
		}
		if m.intro.Phase == 2 {
			subOpacity = 1.0 - fadeOpacity
		}
		subColor := interpolateHexColor("#1D1D2E", "#667085", min(1, subOpacity))
		subtitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(subColor)).
			Render(subText)
	}

	skipHint := ""
	if m.intro.Tick > 5 {
		hintOpacity := min(1.0, float64(m.intro.Tick-5)/10.0)
		if m.intro.Phase == 2 {
			hintOpacity = max(0, hintOpacity*(1.0-fadeOpacity))
		}
		hintColor := interpolateHexColor("#1D1D2E", "#475467", hintOpacity)
		skipHint = lipgloss.NewStyle().
			Foreground(lipgloss.Color(hintColor)).
			Italic(true).
			Render("press any key to skip")
	}

	artBlock := strings.Join(rendered, "\n")
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		artBlock,
		"",
		subtitle,
		"",
		skipHint,
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

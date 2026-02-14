package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	colorful "github.com/lucasb-eyer/go-colorful"
)

// Color palette
var (
	colorPrimary      = lipgloss.AdaptiveColor{Light: "#7928CA", Dark: "#7D56F4"}
	colorSecondary    = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#848484"}
	colorSuccess      = lipgloss.AdaptiveColor{Light: "#12B76A", Dark: "#73F59F"}
	colorWarning      = lipgloss.AdaptiveColor{Light: "#DC6803", Dark: "#F79009"}
	colorDanger       = lipgloss.AdaptiveColor{Light: "#D92D20", Dark: "#F97066"}
	colorMuted        = lipgloss.AdaptiveColor{Light: "#98A2B3", Dark: "#667085"}
	colorText         = lipgloss.AdaptiveColor{Light: "#1D2939", Dark: "#F2F4F7"}
	colorHighlight    = lipgloss.AdaptiveColor{Light: "#E040FB", Dark: "#EA80FC"}
	colorBorder       = lipgloss.AdaptiveColor{Light: "#D0D5DD", Dark: "#475467"}
	colorBorderActive = lipgloss.AdaptiveColor{Light: "#7928CA", Dark: "#7D56F4"}
	colorRowAlt       = lipgloss.AdaptiveColor{Light: "#F9FAFB", Dark: "#1D2939"}
	colorSubtle       = lipgloss.AdaptiveColor{Light: "#EAECF0", Dark: "#344054"}
)

// Platform styles
var (
	eBayPlatformStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#E53238"))
	mercariPlatformStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#4DC9F6"))
	amazonPlatformStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9900"))
	facebookPlatformStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1877F2"))
	defaultPlatformStyle  = lipgloss.NewStyle().Foreground(colorSecondary)
)

// Panel styles
var (
	// Base panel style with border
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Foreground(colorText).
			Padding(0, 1)

	// Active (focused) panel style
	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(colorBorderActive).
				Foreground(colorText).
				Padding(0, 1)

	// Panel title style
	titleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Active title style
	activeTitleStyle = lipgloss.NewStyle().
				Foreground(colorHighlight).
				Bold(true)

	// Highlight style for panel icons
	highlightStyle = lipgloss.NewStyle().
			Foreground(colorHighlight).
			Bold(true)
)

// Text styles
var (
	// Normal text
	textStyle = lipgloss.NewStyle().
			Foreground(colorText)

	// Muted/help text
	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Success/profit text
	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// Warning text
	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	// Danger/negative text
	dangerStyle = lipgloss.NewStyle().
			Foreground(colorDanger)

	// Highlighted/selected text
	selectedStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorPrimary).
			Padding(0, 1)

	// Price style
	priceStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	// Status: Sold
	soldStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Status: Active
	activeStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// Base row style
	rowStyle = lipgloss.NewStyle().
			Foreground(colorText)

	// Alternate row style
	rowAltStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorRowAlt)

	// Key badge style
	keyStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorSubtle).
			Bold(true).
			Padding(0, 1)

	// Key description style
	keyDescStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Label style
	labelStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Value style
	valueStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true)

	// Separator style
	separatorStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	// Empty state style
	emptyStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	// Spinner style
	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)

	// History item style
	historyItemStyle = lipgloss.NewStyle().
				Foreground(colorText)

	// Selected history item style
	historySelectedStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Background(colorPrimary).
				Padding(0, 1)

	// App header style
	appHeaderStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Scroll indicator style
	scrollInfoStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)
)

// Table header style
var headerStyle = lipgloss.NewStyle().
	Foreground(colorSecondary).
	Bold(true).
	BorderBottom(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colorBorder)

// Help bar style
var helpStyle = lipgloss.NewStyle().
	Foreground(colorMuted).
	MarginTop(1)

func platformStyleFor(name string) lipgloss.Style {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "ebay":
		return eBayPlatformStyle
	case "mercari":
		return mercariPlatformStyle
	case "amazon":
		return amazonPlatformStyle
	case "facebook", "facebook marketplace":
		return facebookPlatformStyle
	default:
		return defaultPlatformStyle
	}
}

func renderGradientText(text, colorA, colorB string) string {
	if text == "" {
		return ""
	}

	runes := []rune(text)
	if len(runes) == 1 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colorA)).Render(text)
	}

	var b strings.Builder
	for i, r := range runes {
		t := float64(i) / float64(len(runes)-1)
		color := interpolateHexColor(colorA, colorB, t)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(string(r)))
	}
	return b.String()
}

func interpolateHexColor(colorA, colorB string, t float64) string {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	start, err := colorful.Hex(colorA)
	if err != nil {
		return colorA
	}
	end, err := colorful.Hex(colorB)
	if err != nil {
		return colorB
	}
	return start.BlendHsv(end, t).Clamped().Hex()
}

// renderPanel creates a styled panel with title.
func renderPanel(icon, label, content string, width, height int, active, flashActive bool) string {
	style := panelStyle
	tStyle := titleStyle
	borderColor := colorBorder
	leftCorner := "╭─"
	rightCorner := "╮"
	dash := "─"

	if active {
		style = activePanelStyle
		tStyle = activeTitleStyle
		borderColor = colorBorderActive
		leftCorner = "┏━"
		rightCorner = "┓"
		dash = "━"
	}
	if flashActive && active {
		borderColor = colorHighlight
	}

	if width < 4 {
		width = 4
	}
	if height < 1 {
		height = 1
	}

	maxTitleWidth := width - 2
	if maxTitleWidth < 1 {
		maxTitleWidth = 1
	}

	styledTitle := renderPanelTitle(icon, label, tStyle, maxTitleWidth)
	titleWidth := lipgloss.Width(styledTitle)
	rightDashCount := max(0, width-titleWidth-1)
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	topBorder := borderStyle.Render(leftCorner) + styledTitle + borderStyle.Render(strings.Repeat(dash, rightDashCount)+rightCorner)

	contentBox := style.
		BorderForeground(borderColor).
		Width(width).
		Height(height).
		BorderTop(false).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, topBorder, contentBox)
}

func renderPanelTitle(icon, label string, titleTextStyle lipgloss.Style, maxTitleWidth int) string {
	icon = strings.TrimSpace(icon)
	if maxTitleWidth < 1 {
		return ""
	}

	if icon == "" {
		return fitStyledTitle(label, titleTextStyle, maxTitleWidth)
	}

	iconWidth := lipgloss.Width(icon)
	if iconWidth >= maxTitleWidth {
		return highlightStyle.Render(truncate(icon, maxTitleWidth))
	}

	labelWidth := maxTitleWidth - iconWidth - 1
	if labelWidth <= 0 {
		return highlightStyle.Render(icon)
	}

	rawLabel := truncate(label, labelWidth)
	styledTitle := highlightStyle.Render(icon) + " " + titleTextStyle.Render(rawLabel)
	for lipgloss.Width(styledTitle) > maxTitleWidth && labelWidth > 0 {
		labelWidth--
		rawLabel = truncate(label, labelWidth)
		if rawLabel == "" {
			styledTitle = highlightStyle.Render(icon)
			break
		}
		styledTitle = highlightStyle.Render(icon) + " " + titleTextStyle.Render(rawLabel)
	}

	if lipgloss.Width(styledTitle) > maxTitleWidth {
		return highlightStyle.Render(truncate(icon, maxTitleWidth))
	}

	return styledTitle
}

func fitStyledTitle(raw string, s lipgloss.Style, maxWidth int) string {
	if maxWidth < 1 {
		return ""
	}

	allowed := maxWidth
	trimmed := truncate(raw, allowed)
	styled := s.Render(trimmed)
	for lipgloss.Width(styled) > maxWidth && allowed > 0 {
		allowed--
		trimmed = truncate(raw, allowed)
		styled = s.Render(trimmed)
	}

	if lipgloss.Width(styled) > maxWidth {
		return s.Render("…")
	}
	return styled
}

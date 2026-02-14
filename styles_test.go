package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderGradientText(t *testing.T) {
	input := "m r k t r"
	output := renderGradientText(input, "#7D56F4", "#EA80FC")

	if output == "" {
		t.Fatal("expected gradient output to be non-empty")
	}
	if got, want := lipgloss.Width(output), lipgloss.Width(input); got != want {
		t.Fatalf("expected visual width %d, got %d", want, got)
	}
}

func TestPlatformColorStyles(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  lipgloss.TerminalColor
	}{
		{name: "ebay", input: "eBay", want: lipgloss.Color("#E53238")},
		{name: "mercari", input: "Mercari", want: lipgloss.Color("#4DC9F6")},
		{name: "amazon", input: "Amazon", want: lipgloss.Color("#FF9900")},
		{name: "facebook", input: "Facebook", want: lipgloss.Color("#1877F2")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := platformStyleFor(tt.input)
			if got := style.GetForeground(); got != tt.want {
				t.Fatalf("expected foreground %v, got %v", tt.want, got)
			}

			rendered := style.Render("platform")
			if strings.TrimSpace(rendered) == "" {
				t.Fatalf("expected rendered platform text for %s to be non-empty", tt.input)
			}
		})
	}

	unknown := platformStyleFor("Unknown")
	if got, want := unknown.GetForeground(), defaultPlatformStyle.GetForeground(); got != want {
		t.Fatalf("expected unknown platform fallback foreground %v, got %v", want, got)
	}
}

func TestRenderPanelTitleTruncationSafety(t *testing.T) {
	panelWidth := 14
	panel := renderPanel("/", "超長いタイトルで切り詰めを確認する", "content", panelWidth, 1, false, false)
	lines := strings.Split(panel, "\n")
	if len(lines) == 0 {
		t.Fatal("expected rendered panel lines")
	}
	if got, want := lipgloss.Width(lines[0]), panelWidth+2; got != want {
		t.Fatalf("expected top border width %d, got %d", want, got)
	}
	if !strings.Contains(lines[0], "/") {
		t.Fatalf("expected icon to remain visible in title, got %q", lines[0])
	}
}

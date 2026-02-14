package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit       key.Binding
	ForceQuit  key.Binding
	Tab        key.Binding
	ShiftTab   key.Binding
	Search     key.Binding
	Calculator key.Binding
	Escape     key.Binding
	Enter      key.Binding
	Down       key.Binding
	Up         key.Binding
	HistNext   key.Binding
	HistPrev   key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next/accept"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev panel"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Calculator: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "cost"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "results"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "run/open"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "navigate"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "navigate"),
		),
		HistNext: key.NewBinding(
			key.WithKeys("j", "right"),
			key.WithHelp("j/right", "next history"),
		),
		HistPrev: key.NewBinding(
			key.WithKeys("k", "left"),
			key.WithHelp("k/left", "prev history"),
		),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Search, k.Enter, k.Down, k.Tab, k.Calculator, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Search, k.Enter, k.Escape},
		{k.Down, k.Up, k.HistNext, k.HistPrev},
		{k.Tab, k.ShiftTab, k.Calculator, k.Quit, k.ForceQuit},
	}
}

package main

import "testing"

func TestParseBoolishEnv(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "1", want: true},
		{input: "true", want: true},
		{input: "TRUE", want: true},
		{input: "yes", want: true},
		{input: "on", want: true},
		{input: "0", want: false},
		{input: "false", want: false},
		{input: "", want: false},
		{input: "off", want: false},
	}

	for _, tc := range tests {
		got := parseBoolishEnv(tc.input)
		if got != tc.want {
			t.Fatalf("parseBoolishEnv(%q): expected %v, got %v", tc.input, tc.want, got)
		}
	}
}

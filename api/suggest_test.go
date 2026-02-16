package api

import (
	"strings"
	"testing"
)

func TestProductIndexExpandUsesHighConfidenceMatch(t *testing.T) {
	idx := newProductIndexFromEntries([]ProductEntry{
		{
			Name:     "PlayStation 5 Console",
			Category: "Gaming",
			Synonyms: []string{"ps5", "playstation 5"},
		},
		{
			Name:     "Nintendo Switch OLED",
			Category: "Gaming",
			Synonyms: []string{"switch"},
		},
	})

	if got := idx.Expand("ps5"); got != "PlayStation 5 Console" {
		t.Fatalf("expected query expansion to %q, got %q", "PlayStation 5 Console", got)
	}
}

func TestProductIndexExpandKeepsOriginalWhenAmbiguous(t *testing.T) {
	idx := newProductIndexFromEntries([]ProductEntry{
		{
			Name:     "Nintendo Switch OLED",
			Category: "Gaming",
			Synonyms: []string{"switch"},
		},
		{
			Name:     "Nintendo Switch Lite",
			Category: "Gaming",
			Synonyms: []string{"switch"},
		},
	})

	if got := idx.Expand("switch"); got != "switch" {
		t.Fatalf("expected ambiguous query to remain unchanged, got %q", got)
	}
}

func TestProductIndexExpandHandlesEmptyAndUnknownQueries(t *testing.T) {
	idx := newProductIndexFromEntries([]ProductEntry{
		{
			Name:     "AirPods Pro 2",
			Category: "Audio",
			Synonyms: []string{"airpods pro"},
		},
	})

	cases := []string{"", "   ", "totally unknown item"}
	for _, tc := range cases {
		want := strings.TrimSpace(tc)
		if got := idx.Expand(tc); got != want {
			t.Fatalf("expected query %q to remain unchanged as %q, got %q", tc, want, got)
		}
	}
}

func TestProductIndexSuggestReturnsTopPrefixMatches(t *testing.T) {
	idx := newProductIndexFromEntries([]ProductEntry{
		{
			Name:     "Nintendo Switch OLED",
			Category: "Gaming",
			Synonyms: []string{"switch oled", "nintendo switch"},
		},
		{
			Name:     "Nintendo Switch Lite",
			Category: "Gaming",
			Synonyms: []string{"switch lite"},
		},
		{
			Name:     "PlayStation 5 Console",
			Category: "Gaming",
			Synonyms: []string{"ps5"},
		},
	})

	got := idx.Suggest("swi")
	if len(got) < 2 {
		t.Fatalf("expected at least two switch suggestions, got %v", got)
	}
	if !strings.HasPrefix(strings.ToLower(got[0]), "swi") {
		t.Fatalf("expected top suggestion to keep typed prefix, got %q", got[0])
	}
}

func TestProductIndexSuggestMatchesTokenPrefixes(t *testing.T) {
	idx := newProductIndexFromEntries([]ProductEntry{
		{
			Name:     "Nintendo Switch OLED",
			Category: "Gaming",
			Synonyms: []string{"switch oled", "nintendo switch"},
		},
	})

	got := idx.Suggest("ole")
	if len(got) == 0 {
		t.Fatal("expected token-prefix suggestion for inner token")
	}

	found := false
	for _, suggestion := range got {
		if strings.Contains(strings.ToLower(suggestion), "oled") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected OLED token match in suggestions, got %v", got)
	}
}

func TestNewProductIndexLoadsEmbeddedCatalog(t *testing.T) {
	idx := NewProductIndex()
	if idx == nil {
		t.Fatal("expected product index to be initialized")
	}
	got := idx.Suggest("ps5")
	if len(got) == 0 {
		t.Fatal("expected embedded catalog to provide suggestions")
	}
}

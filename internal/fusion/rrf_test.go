package fusion

import (
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com/path/", "https://example.com/path"},
		{"https://EXAMPLE.COM/path#section1", "https://example.com/path"},
		{"http://go.dev/doc/?q=1#foo", "http://go.dev/doc?q=1"},
		{"", ""},
	}

	for _, tc := range tests {
		got := NormalizeURL(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizeURL(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestReciprocalRankFusion(t *testing.T) {
	list1 := []Item{
		{Title: "Go 1", URL: "https://go.dev/1", Content: "c1", Engine: "google"},
		{Title: "Go 2", URL: "https://go.dev/2", Content: "c2", Engine: "google"},
	}
	list2 := []Item{
		{Title: "Go 2 Alt", URL: "https://go.dev/2#anchor", Content: "c2 longer content", Engine: "exa"},
		{Title: "Go 3", URL: "https://go.dev/3", Content: "c3", Engine: "exa"},
	}

	fused := ReciprocalRankFusion([][]Item{list1, list2}, 60)

	if len(fused) != 3 {
		t.Fatalf("expected 3 items after deduplication, got %d", len(fused))
	}

	// Go 2 appeared in both lists, so it should have the highest RRF score!
	if fused[0].URL != "https://go.dev/2" {
		t.Errorf("expected top item to be https://go.dev/2, got %s", fused[0].URL)
	}

	if len(fused[0].Engines) != 2 {
		t.Errorf("expected 2 engines for fused item, got %v", fused[0].Engines)
	}

	if fused[0].Content != "c2 longer content" {
		t.Errorf("expected longer content preserved, got %s", fused[0].Content)
	}
}

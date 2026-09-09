package searxng

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientSearchSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "golang" {
			t.Fatalf("unexpected query: %s", r.URL.Query().Get("q"))
		}
		raw := rawSearxResponse{
			Results: []struct {
				Title         string   `json:"title"`
				URL           string   `json:"url"`
				Content       string   `json:"content"`
				Engine        string   `json:"engine"`
				Engines       []string `json:"engines"`
				Score         float64  `json:"score"`
				Category      string   `json:"category"`
				PublishedDate string   `json:"publishedDate"`
			}{
				{
					Title:   "Go Programming Language",
					URL:     "https://go.dev",
					Content: "Build simple, secure, scalable systems with Go",
					Engine:  "google",
					Score:   1.5,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(raw)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, 2*time.Second)
	res, err := client.Search(context.Background(), SearchParams{
		Query:      "golang",
		MaxResults: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Count != 1 {
		t.Errorf("expected 1 result, got %d", res.Count)
	}
	if res.Results[0].Title != "Go Programming Language" {
		t.Errorf("expected title 'Go Programming Language', got '%s'", res.Results[0].Title)
	}
	if res.Results[0].URL != "https://go.dev" {
		t.Errorf("expected URL 'https://go.dev', got '%s'", res.Results[0].URL)
	}
}

func TestClientSearchError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Rate limited", http.StatusTooManyRequests)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, 2*time.Second)
	res, err := client.Search(context.Background(), SearchParams{Query: "test"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !res.Degraded {
		t.Errorf("expected degraded=true")
	}
	if res.Count != 0 {
		t.Errorf("expected count=0, got %d", res.Count)
	}
}

func TestClientHealth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(rawSearxResponse{})
		} else if r.URL.Path == "/config" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"version": "2026.08.21"}`))
		}
	}))
	defer ts.Close()

	client := NewClient(ts.URL, 2*time.Second)
	info := client.Health(context.Background())
	if info.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", info.Status)
	}
	if !info.Reachable {
		t.Errorf("expected reachable=true")
	}
	if info.Version != "2026.08.21" {
		t.Errorf("expected version 2026.08.21, got '%s'", info.Version)
	}
}

package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigDefaults(t *testing.T) {
	os.Unsetenv("SEARXNG_URL")
	os.Unsetenv("SEARXNG_DEFAULT_MAX")
	os.Unsetenv("CASCADE_TIMEOUT")

	cfg := Load()
	if cfg.SearXNGURL != "http://127.0.0.1:8889" {
		t.Errorf("Expected SearXNGURL http://127.0.0.1:8889, got %s", cfg.SearXNGURL)
	}
	if cfg.DefaultMaxResults != 10 {
		t.Errorf("Expected DefaultMaxResults 10, got %d", cfg.DefaultMaxResults)
	}
	if cfg.MaxAllowedResults != 50 {
		t.Errorf("Expected MaxAllowedResults 50, got %d", cfg.MaxAllowedResults)
	}
	if cfg.RRFK != 60 {
		t.Errorf("Expected RRFK 60, got %d", cfg.RRFK)
	}
	if cfg.CascadeTimeout != 6*time.Second {
		t.Errorf("Expected CascadeTimeout 6s, got %v", cfg.CascadeTimeout)
	}
}

func TestConfigEnvOverrides(t *testing.T) {
	t.Setenv("SEARXNG_URL", "http://searxng.local:8080")
	t.Setenv("SEARXNG_DEFAULT_MAX", "25")
	t.Setenv("CASCADE_TIMEOUT", "12")
	t.Setenv("RRF_K", "100")

	cfg := Load()
	if cfg.SearXNGURL != "http://searxng.local:8080" {
		t.Errorf("Expected SearXNGURL http://searxng.local:8080, got %s", cfg.SearXNGURL)
	}
	if cfg.DefaultMaxResults != 25 {
		t.Errorf("Expected DefaultMaxResults 25, got %d", cfg.DefaultMaxResults)
	}
	if cfg.CascadeTimeout != 12*time.Second {
		t.Errorf("Expected CascadeTimeout 12s, got %v", cfg.CascadeTimeout)
	}
	if cfg.RRFK != 100 {
		t.Errorf("Expected RRFK 100, got %d", cfg.RRFK)
	}
}

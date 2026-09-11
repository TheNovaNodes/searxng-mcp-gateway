package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration for searxng-mcp-gateway.
type Config struct {
	SearXNGURL         string
	DefaultMaxResults  int
	MaxAllowedResults  int
	DefaultLanguage    string
	DefaultSafeSearch  int
	SearchTimeout      time.Duration
	CascadeTimeout     time.Duration
	RRFK               int
	VaultDir           string
	AllowPrivateScrape bool
}

// Load loads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		SearXNGURL:         getEnv("SEARXNG_URL", "http://127.0.0.1:8889"),
		DefaultMaxResults:  getEnvInt("SEARXNG_DEFAULT_MAX", 10),
		MaxAllowedResults:  getEnvInt("SEARXNG_MAX_ALLOWED", 50),
		DefaultLanguage:    getEnv("SEARXNG_DEFAULT_LANG", "auto"),
		DefaultSafeSearch:  getEnvInt("SEARXNG_SAFESEARCH", 0),
		SearchTimeout:      time.Duration(getEnvInt("SEARXNG_TIMEOUT", 10)) * time.Second,
		CascadeTimeout:     time.Duration(getEnvInt("CASCADE_TIMEOUT", 6)) * time.Second,
		RRFK:               getEnvInt("RRF_K", 60),
		VaultDir:           getEnv("AGENT_VAULT_DIR", "/dev/shm/agent_vault"),
		AllowPrivateScrape: getEnvBool("SEARXNG_ALLOW_PRIVATE_SCRAPE", false),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i > 0 {
			return i
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		lower := strings.ToLower(strings.TrimSpace(val))
		return lower == "1" || lower == "true" || lower == "yes"
	}
	return defaultVal
}

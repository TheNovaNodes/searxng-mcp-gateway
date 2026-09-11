package vault

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DefaultPointers map provider names to their vault filenames.
var DefaultPointers = map[string]string{
	"tavily":    "gsAYQxw",
	"firecrawl": "x01eFQ",
	"exa":       "H0bWdb",
	"olostep":   "Ib50He",
	"tinyfish":  "u7jMwl",
}

// DefaultEnvVars map provider names to standard environment variable fallbacks.
var DefaultEnvVars = map[string][]string{
	"tavily":    {"TAVILY_API_KEY"},
	"firecrawl": {"FIRECRAWL_API_KEY"},
	"olostep":   {"OLOSTEP_API_KEY"},
	"exa":       {"EXA_API_KEY"},
	"tinyfish":  {"TINYFISH_API_KEY"},
}

var (
	reTavily    = regexp.MustCompile(`tvly-dev-[a-zA-Z0-9_-]+`)
	reFirecrawl = regexp.MustCompile(`fc-[a-zA-Z0-9_-]+`)
	reOlostep   = regexp.MustCompile(`olostep_[a-zA-Z0-9_-]+`)
	reExa       = regexp.MustCompile(`(?:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|[A-Za-z0-9_-]{32,})`)
)

// Vault manages extraction of credentials from the shared memory vault or environment variables.
type Vault struct {
	dir string
}

// NewVault creates a new Vault manager.
func NewVault(dir string) *Vault {
	if dir == "" {
		dir = "/dev/shm/agent_vault"
	}
	return &Vault{dir: dir}
}

func parseKeysFromContent(provider, content string) []string {
	switch strings.ToLower(provider) {
	case "tavily":
		return reTavily.FindAllString(content, -1)
	case "firecrawl":
		return reFirecrawl.FindAllString(content, -1)
	case "olostep":
		return reOlostep.FindAllString(content, -1)
	case "exa":
		raw := reExa.FindAllString(content, -1)
		var valid []string
		for _, k := range raw {
			if len(k) >= 32 && !strings.EqualFold(k, "exa") {
				valid = append(valid, k)
			}
		}
		return valid
	default:
		return nil
	}
}

// GetKeys extracts valid API keys for the given provider from SHM vault or environment variables.
func (v *Vault) GetKeys(provider string) []string {
	prov := strings.ToLower(provider)
	var keys []string

	// 1. Try SHM vault directory first
	if pointer, ok := DefaultPointers[prov]; ok {
		filePath := filepath.Join(v.dir, pointer)
		if contentBytes, err := os.ReadFile(filePath); err == nil {
			keys = append(keys, parseKeysFromContent(prov, string(contentBytes))...)
		}
	}

	// 2. Fallback to environment variables if no keys were found in vault
	if len(keys) == 0 {
		if envVars, ok := DefaultEnvVars[prov]; ok {
			for _, envName := range envVars {
				if val := os.Getenv(envName); val != "" {
					parsed := parseKeysFromContent(prov, val)
					if len(parsed) == 0 {
						trimmed := strings.TrimSpace(val)
						if len(trimmed) > 8 {
							parsed = []string{trimmed}
						}
					}
					keys = append(keys, parsed...)
				}
			}
		}
	}

	// Deduplicate keys
	seen := make(map[string]bool)
	var unique []string
	for _, k := range keys {
		if !seen[k] && k != "" {
			seen[k] = true
			unique = append(unique, k)
		}
	}
	return unique
}

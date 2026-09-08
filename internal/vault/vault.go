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

var (
	reTavily   = regexp.MustCompile(`tvly-dev-[a-zA-Z0-9_-]+`)
	reFirecrawl = regexp.MustCompile(`fc-[a-zA-Z0-9_-]+`)
	reOlostep  = regexp.MustCompile(`olostep_[a-zA-Z0-9_-]+`)
	reExa      = regexp.MustCompile(`(?:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|[A-Za-z0-9_-]{32,})`)
)

// Vault manages extraction of credentials from the shared memory vault.
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

// GetKeys extracts valid API keys for the given provider.
func (v *Vault) GetKeys(provider string) []string {
	pointer, ok := DefaultPointers[strings.ToLower(provider)]
	if !ok {
		return nil
	}

	filePath := filepath.Join(v.dir, pointer)
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	content := string(contentBytes)

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

package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVaultExtraction(t *testing.T) {
	tempDir := t.TempDir()

	// Tavily
	_ = os.WriteFile(filepath.Join(tempDir, "gsAYQxw"), []byte("tvly-dev-testkey1234567890abcdef\nother tvly-dev-secondkey9876543210"), 0600)
	// Firecrawl
	_ = os.WriteFile(filepath.Join(tempDir, "x01eFQ"), []byte("key: fc-samplefirecrawlkey123456"), 0600)
	// Olostep
	_ = os.WriteFile(filepath.Join(tempDir, "Ib50He"), []byte("olostep_mysecretkey_abcdef123"), 0600)
	// Exa
	_ = os.WriteFile(filepath.Join(tempDir, "H0bWdb"), []byte("exa 12345678-1234-1234-1234-1234567890ab\nshort a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"), 0600)

	v := NewVault(tempDir)

	tavilyKeys := v.GetKeys("tavily")
	if len(tavilyKeys) != 2 {
		t.Fatalf("expected 2 tavily keys, got %d", len(tavilyKeys))
	}
	if tavilyKeys[0] != "tvly-dev-testkey1234567890abcdef" {
		t.Errorf("unexpected key: %s", tavilyKeys[0])
	}

	fcKeys := v.GetKeys("firecrawl")
	if len(fcKeys) != 1 || fcKeys[0] != "fc-samplefirecrawlkey123456" {
		t.Fatalf("unexpected firecrawl keys: %v", fcKeys)
	}

	oloKeys := v.GetKeys("olostep")
	if len(oloKeys) != 1 || oloKeys[0] != "olostep_mysecretkey_abcdef123" {
		t.Fatalf("unexpected olostep keys: %v", oloKeys)
	}

	exaKeys := v.GetKeys("exa")
	if len(exaKeys) != 2 {
		t.Fatalf("expected 2 exa keys, got %d", len(exaKeys))
	}
}

func TestVaultLiveSHM(t *testing.T) {
	// Test against live /dev/shm/agent_vault if it exists
	if _, err := os.Stat("/dev/shm/agent_vault"); os.IsNotExist(err) {
		t.Skip("/dev/shm/agent_vault not present")
	}

	v := NewVault("/dev/shm/agent_vault")
	exaKeys := v.GetKeys("exa")
	if len(exaKeys) == 0 {
		t.Log("Note: No Exa keys in live vault")
	} else {
		t.Logf("Found %d live Exa keys", len(exaKeys))
	}
}

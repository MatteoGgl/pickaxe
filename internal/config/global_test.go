package config_test

import (
	"path/filepath"
	"testing"

	"github.com/matteo/pickaxe/internal/config"
	"github.com/matteo/pickaxe/internal/testutil"
)

func TestGlobalConfig_RoundTrip(t *testing.T) {
	dir := testutil.TempDir(t)
	path := filepath.Join(dir, "config.json")

	cfg := &config.GlobalConfig{VaultRoot: "/Users/matteo/vault"}
	if err := config.WriteGlobalConfig(path, cfg); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := config.ReadGlobalConfig(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.VaultRoot != cfg.VaultRoot {
		t.Errorf("vault_root: got %q, want %q", got.VaultRoot, cfg.VaultRoot)
	}
}

func TestGlobalConfig_NotFound(t *testing.T) {
	_, err := config.ReadGlobalConfig("/nonexistent/config.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestGlobalConfig_CreatesParentDirs(t *testing.T) {
	dir := testutil.TempDir(t)
	path := filepath.Join(dir, "nested", "deep", "config.json")

	cfg := &config.GlobalConfig{VaultRoot: "/vault"}
	if err := config.WriteGlobalConfig(path, cfg); err != nil {
		t.Fatalf("write: %v", err)
	}
}

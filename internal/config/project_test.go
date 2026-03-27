package config_test

import (
	"path/filepath"
	"testing"

	"github.com/matteo/pickaxe/internal/config"
	"github.com/matteo/pickaxe/internal/testutil"
)

func TestProjectConfig_RoundTrip(t *testing.T) {
	dir := testutil.TempDir(t)
	path := filepath.Join(dir, ".pickaxe.json")

	cfg := &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: "/vault/adrs.md", Name: "adrs"},
			{Type: config.EntryTypeDir, Path: "/vault/foo/", Name: "foo", Recursive: false},
		},
	}

	if err := config.WriteProjectConfig(path, cfg); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := config.ReadProjectConfig(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got.Entries) != 2 {
		t.Fatalf("entries: got %d, want 2", len(got.Entries))
	}
	if got.Entries[0].Name != "adrs" {
		t.Errorf("entry[0].name: got %q, want %q", got.Entries[0].Name, "adrs")
	}
	if got.Entries[1].Type != config.EntryTypeDir {
		t.Errorf("entry[1].type: got %q, want %q", got.Entries[1].Type, config.EntryTypeDir)
	}
}

func TestProjectConfig_NotFound(t *testing.T) {
	_, err := config.ReadProjectConfig("/nonexistent/.pickaxe.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestProjectConfig_EmptyEntries(t *testing.T) {
	dir := testutil.TempDir(t)
	path := filepath.Join(dir, ".pickaxe.json")

	cfg := &config.ProjectConfig{Version: 1, Entries: []config.Entry{}}
	if err := config.WriteProjectConfig(path, cfg); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := config.ReadProjectConfig(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(got.Entries))
	}
}

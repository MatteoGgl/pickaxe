package registry_test

import (
	"path/filepath"
	"testing"

	"github.com/matteo/pickaxe/internal/config"
	"github.com/matteo/pickaxe/internal/registry"
	"github.com/matteo/pickaxe/internal/testutil"
)

func newConfig() *config.ProjectConfig {
	return &config.ProjectConfig{Version: 1, Entries: []config.Entry{}}
}

// --- DefaultName ---

func TestDefaultName_File(t *testing.T) {
	got := registry.DefaultName("/vault/projects/foo/adrs.md")
	if got != "adrs" {
		t.Errorf("got %q, want %q", got, "adrs")
	}
}

func TestDefaultName_DirTrailingSlash(t *testing.T) {
	got := registry.DefaultName("/vault/projects/foo/")
	if got != "foo" {
		t.Errorf("got %q, want %q", got, "foo")
	}
}

// --- AddFile ---

func TestAddFile_Success(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "# ADRs")

	cfg := newConfig()
	if err := registry.AddFile(cfg, filePath, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(cfg.Entries))
	}
	if cfg.Entries[0].Name != "adrs" {
		t.Errorf("name: got %q, want %q", cfg.Entries[0].Name, "adrs")
	}
	if cfg.Entries[0].Type != config.EntryTypeFile {
		t.Errorf("type: got %q, want %q", cfg.Entries[0].Type, config.EntryTypeFile)
	}
}

func TestAddFile_CustomName(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "# ADRs")

	cfg := newConfig()
	if err := registry.AddFile(cfg, filePath, "my-adrs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Entries[0].Name != "my-adrs" {
		t.Errorf("name: got %q, want %q", cfg.Entries[0].Name, "my-adrs")
	}
}

func TestAddFile_NonExistentPath(t *testing.T) {
	cfg := newConfig()
	err := registry.AddFile(cfg, "/nonexistent/file.md", "")
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

func TestAddFile_NameCollision(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "")

	cfg := newConfig()
	if err := registry.AddFile(cfg, filePath, ""); err != nil {
		t.Fatal(err)
	}

	if err := registry.AddFile(cfg, filePath, ""); err == nil {
		t.Fatal("expected name collision error, got nil")
	}
}

// --- AddDir ---

func TestAddDir_Flat(t *testing.T) {
	dir := testutil.TempDir(t)
	testutil.WriteFile(t, filepath.Join(dir, "a.md"), "")
	testutil.WriteFile(t, filepath.Join(dir, "b.md"), "")
	testutil.WriteFile(t, filepath.Join(dir, "sub", "c.md"), "")

	cfg := newConfig()
	if err := registry.AddDir(cfg, dir, "", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Entries) != 1 {
		t.Fatalf("expected 1 dir entry, got %d", len(cfg.Entries))
	}
	if cfg.Entries[0].Type != config.EntryTypeDir {
		t.Errorf("type: got %q, want %q", cfg.Entries[0].Type, config.EntryTypeDir)
	}
	if cfg.Entries[0].Recursive {
		t.Error("expected recursive=false")
	}
}

func TestAddDir_Recursive(t *testing.T) {
	dir := testutil.TempDir(t)
	testutil.WriteFile(t, filepath.Join(dir, "a.md"), "")
	testutil.WriteFile(t, filepath.Join(dir, "sub", "b.md"), "")

	cfg := newConfig()
	if err := registry.AddDir(cfg, dir, "", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Entries[0].Recursive {
		t.Error("expected recursive=true")
	}
}

func TestAddDir_NonExistentPath(t *testing.T) {
	cfg := newConfig()
	if err := registry.AddDir(cfg, "/nonexistent/", "", false); err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

// --- Remove ---

func TestRemove_Success(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "")

	cfg := newConfig()
	_ = registry.AddFile(cfg, filePath, "")

	if err := registry.Remove(cfg, "adrs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Entries) != 0 {
		t.Errorf("expected 0 entries after remove, got %d", len(cfg.Entries))
	}
}

func TestRemove_NotFound(t *testing.T) {
	cfg := newConfig()
	if err := registry.Remove(cfg, "nonexistent"); err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

// --- EnumerateFiles ---

func TestEnumerateFiles_Flat(t *testing.T) {
	dir := testutil.TempDir(t)
	testutil.WriteFile(t, filepath.Join(dir, "a.md"), "content a")
	testutil.WriteFile(t, filepath.Join(dir, "b.md"), "content b")
	testutil.WriteFile(t, filepath.Join(dir, "sub", "c.md"), "content c")
	testutil.WriteFile(t, filepath.Join(dir, "notes.txt"), "")

	entry := config.Entry{Type: config.EntryTypeDir, Path: dir, Name: "myfolder", Recursive: false}
	files, err := registry.EnumerateFiles(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files (flat .md only), got %d", len(files))
	}
	for _, f := range files {
		if f.Name[:len("myfolder/")] != "myfolder/" {
			t.Errorf("name %q not prefixed with myfolder/", f.Name)
		}
	}
}

func TestEnumerateFiles_Recursive(t *testing.T) {
	dir := testutil.TempDir(t)
	testutil.WriteFile(t, filepath.Join(dir, "a.md"), "")
	testutil.WriteFile(t, filepath.Join(dir, "sub", "b.md"), "")

	entry := config.Entry{Type: config.EntryTypeDir, Path: dir, Name: "myfolder", Recursive: true}
	files, err := registry.EnumerateFiles(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files (recursive .md), got %d", len(files))
	}
}

func TestEnumerateFiles_SingleFile(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "# ADRs")

	entry := config.Entry{Type: config.EntryTypeFile, Path: filePath, Name: "adrs"}
	files, err := registry.EnumerateFiles(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Name != "adrs" {
		t.Errorf("name: got %q, want %q", files[0].Name, "adrs")
	}
}

package cmd_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/matteoggl/pickaxe/cmd"
	"github.com/matteoggl/pickaxe/internal/testutil"
	"github.com/matteoggl/pickaxe/internal/tui"
	"github.com/matteoggl/pickaxe/internal/vault"
)

func initVault(t *testing.T) (*vault.Vault, string) {
	t.Helper()
	dir := testutil.TempDir(t)
	v, err := vault.Init(dir)
	if err != nil {
		t.Fatal(err)
	}
	return v, dir
}

func fakePicker(result tui.PickerResult, err error) cmd.PickerFunc {
	return func(root string, preSelected map[string]bool) (tui.PickerResult, error) {
		return result, err
	}
}

// --- AddDirect tests ---

func TestAddDirect_AddFile(t *testing.T) {
	v, dir := initVault(t)
	filePath := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, filePath, "hello")

	var buf bytes.Buffer
	if err := cmd.AddDirect(v, filePath, false, "", false, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := v.Entries()
	if len(entries) != 1 || entries[0].Type != vault.EntryTypeFile {
		t.Errorf("expected 1 file entry, got %v", entries)
	}
}

func TestAddDirect_AddDir(t *testing.T) {
	v, dir := initVault(t)
	subDir := filepath.Join(dir, "docs")
	testutil.WriteFile(t, filepath.Join(subDir, "x.md"), "x")

	var buf bytes.Buffer
	if err := cmd.AddDirect(v, subDir, true, "", true, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := v.Entries()
	if len(entries) != 1 || entries[0].Type != vault.EntryTypeDir {
		t.Errorf("expected 1 dir entry, got %v", entries)
	}
	if !entries[0].Recursive {
		t.Error("expected recursive=true")
	}
}

func TestAddDirect_WithAlias(t *testing.T) {
	v, dir := initVault(t)
	filePath := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, filePath, "hello")

	var buf bytes.Buffer
	if err := cmd.AddDirect(v, filePath, false, "my-note", false, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Entries()[0].Name != "my-note" {
		t.Errorf("expected name 'my-note', got %q", v.Entries()[0].Name)
	}
}

func TestAddDirect_NotFound(t *testing.T) {
	v, _ := initVault(t)

	var buf bytes.Buffer
	err := cmd.AddDirect(v, "/nonexistent/file.md", false, "", false, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
}

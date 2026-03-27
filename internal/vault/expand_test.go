package vault_test

import (
	"path/filepath"
	"testing"

	"github.com/matteo/pickaxe/internal/testutil"
	"github.com/matteo/pickaxe/internal/vault"
)

func TestExpandEntries_FileEntry(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, filePath, "content")

	entries := []vault.Entry{
		{Type: vault.EntryTypeFile, Path: filePath, Name: "note"},
	}

	result, err := vault.ExpandEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result))
	}
	if !result[filePath] {
		t.Errorf("expected %q in result", filePath)
	}
}

func TestExpandEntries_FlatDirEntry(t *testing.T) {
	dir := testutil.TempDir(t)
	docsDir := filepath.Join(dir, "docs")
	testutil.WriteFile(t, filepath.Join(docsDir, "a.md"), "")
	testutil.WriteFile(t, filepath.Join(docsDir, "b.md"), "")
	testutil.WriteFile(t, filepath.Join(docsDir, "sub", "c.md"), "")
	testutil.WriteFile(t, filepath.Join(docsDir, "notes.txt"), "")

	entries := []vault.Entry{
		{Type: vault.EntryTypeDir, Path: docsDir, Name: "docs", Recursive: false},
	}

	result, err := vault.ExpandEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 files (flat .md only), got %d", len(result))
	}

	if !result[filepath.Join(docsDir, "a.md")] {
		t.Error("expected a.md in result")
	}
	if !result[filepath.Join(docsDir, "b.md")] {
		t.Error("expected b.md in result")
	}
	if result[filepath.Join(docsDir, "sub", "c.md")] {
		t.Error("expected sub/c.md NOT in result (flat only)")
	}
}

func TestExpandEntries_RecursiveDirEntry(t *testing.T) {
	dir := testutil.TempDir(t)
	docsDir := filepath.Join(dir, "docs")
	testutil.WriteFile(t, filepath.Join(docsDir, "a.md"), "")
	testutil.WriteFile(t, filepath.Join(docsDir, "sub", "b.md"), "")
	testutil.WriteFile(t, filepath.Join(docsDir, "sub", "deep", "c.md"), "")

	entries := []vault.Entry{
		{Type: vault.EntryTypeDir, Path: docsDir, Name: "docs", Recursive: true},
	}

	result, err := vault.ExpandEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 files (recursive), got %d", len(result))
	}

	if !result[filepath.Join(docsDir, "a.md")] {
		t.Error("expected a.md in result")
	}
	if !result[filepath.Join(docsDir, "sub", "b.md")] {
		t.Error("expected sub/b.md in result")
	}
	if !result[filepath.Join(docsDir, "sub", "deep", "c.md")] {
		t.Error("expected sub/deep/c.md in result")
	}
}

func TestExpandEntries_Mixed(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, filePath, "")
	docsDir := filepath.Join(dir, "docs")
	testutil.WriteFile(t, filepath.Join(docsDir, "a.md"), "")
	testutil.WriteFile(t, filepath.Join(docsDir, "b.md"), "")

	entries := []vault.Entry{
		{Type: vault.EntryTypeFile, Path: filePath, Name: "note"},
		{Type: vault.EntryTypeDir, Path: docsDir, Name: "docs", Recursive: false},
	}

	result, err := vault.ExpandEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 files (file + dir), got %d", len(result))
	}

	if !result[filePath] {
		t.Error("expected note.md in result")
	}
	if !result[filepath.Join(docsDir, "a.md")] {
		t.Error("expected docs/a.md in result")
	}
	if !result[filepath.Join(docsDir, "b.md")] {
		t.Error("expected docs/b.md in result")
	}
}

func TestExpandEntries_Empty(t *testing.T) {
	entries := []vault.Entry{}

	result, err := vault.ExpandEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Error("expected empty map, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 files, got %d", len(result))
	}
}

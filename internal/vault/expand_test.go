package vault_test

import (
	"path/filepath"
	"testing"

	"github.com/matteoggl/pickaxe/internal/testutil"
	"github.com/matteoggl/pickaxe/internal/vault"
)

func TestExpandEntries_FileEntry(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, filePath, "content")

	entries := []vault.Entry{
		{Type: vault.EntryTypeFile, Path: filePath, Name: "note"},
	}

	paths, _, err := vault.ExpandEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(paths) != 1 {
		t.Fatalf("expected 1 file, got %d", len(paths))
	}
	if !paths[filePath] {
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

	paths, _, err := vault.ExpandEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 files (flat .md only), got %d", len(paths))
	}

	if !paths[filepath.Join(docsDir, "a.md")] {
		t.Error("expected a.md in result")
	}
	if !paths[filepath.Join(docsDir, "b.md")] {
		t.Error("expected b.md in result")
	}
	if paths[filepath.Join(docsDir, "sub", "c.md")] {
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

	paths, _, err := vault.ExpandEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(paths) != 3 {
		t.Fatalf("expected 3 files (recursive), got %d", len(paths))
	}

	if !paths[filepath.Join(docsDir, "a.md")] {
		t.Error("expected a.md in result")
	}
	if !paths[filepath.Join(docsDir, "sub", "b.md")] {
		t.Error("expected sub/b.md in result")
	}
	if !paths[filepath.Join(docsDir, "sub", "deep", "c.md")] {
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

	paths, _, err := vault.ExpandEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(paths) != 3 {
		t.Fatalf("expected 3 files (file + dir), got %d", len(paths))
	}

	if !paths[filePath] {
		t.Error("expected note.md in result")
	}
	if !paths[filepath.Join(docsDir, "a.md")] {
		t.Error("expected docs/a.md in result")
	}
	if !paths[filepath.Join(docsDir, "b.md")] {
		t.Error("expected docs/b.md in result")
	}
}

func TestExpandEntries_Empty(t *testing.T) {
	entries := []vault.Entry{}

	paths, writable, err := vault.ExpandEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if paths == nil {
		t.Error("expected empty map, got nil")
	}
	if len(paths) != 0 {
		t.Errorf("expected 0 files, got %d", len(paths))
	}
	if len(writable) != 0 {
		t.Errorf("expected 0 writable files, got %d", len(writable))
	}
}

func TestExpandEntries_Writable(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, filePath, "content")
	roPath := filepath.Join(dir, "ro.md")
	testutil.WriteFile(t, roPath, "content")

	entries := []vault.Entry{
		{Type: vault.EntryTypeFile, Path: filePath, Name: "note", Writable: true},
		{Type: vault.EntryTypeFile, Path: roPath, Name: "ro", Writable: false},
	}

	paths, writable, err := vault.ExpandEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !paths[filePath] || !paths[roPath] {
		t.Error("expected both paths in result")
	}
	if !writable[filePath] {
		t.Errorf("expected %q to be writable", filePath)
	}
	if writable[roPath] {
		t.Errorf("expected %q to be read-only", roPath)
	}
}

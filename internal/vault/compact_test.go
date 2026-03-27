package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/matteo/pickaxe/internal/testutil"
	"github.com/matteo/pickaxe/internal/vault"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCompact_Empty(t *testing.T) {
	entries, err := vault.Compact(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty, got %v", entries)
	}

	entries, err = vault.Compact([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty, got %v", entries)
	}
}

func TestCompact_SingleFile(t *testing.T) {
	dir := testutil.TempDir(t)
	f := filepath.Join(dir, "notes.md")
	writeFile(t, f)

	entries, err := vault.Compact([]string{f})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d: %v", len(entries), entries)
	}
	e := entries[0]
	if e.Type != vault.EntryTypeFile {
		t.Errorf("expected file entry, got %s", e.Type)
	}
	if e.Path != f {
		t.Errorf("expected path %s, got %s", f, e.Path)
	}
	if e.Name != "notes" {
		t.Errorf("expected name 'notes', got %s", e.Name)
	}
}

func TestCompact_AllFilesInFlatDir(t *testing.T) {
	dir := testutil.TempDir(t)
	f1 := filepath.Join(dir, "a.md")
	f2 := filepath.Join(dir, "b.md")
	writeFile(t, f1)
	writeFile(t, f2)

	entries, err := vault.Compact([]string{f1, f2})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 dir entry, got %d: %v", len(entries), entries)
	}
	e := entries[0]
	if e.Type != vault.EntryTypeDir {
		t.Errorf("expected dir entry, got %s", e.Type)
	}
	if e.Path != dir {
		t.Errorf("expected path %s, got %s", dir, e.Path)
	}
	if e.Recursive {
		t.Error("expected non-recursive (no subdirs)")
	}
}

func TestCompact_PartialDirSelection(t *testing.T) {
	dir := testutil.TempDir(t)
	f1 := filepath.Join(dir, "a.md")
	f2 := filepath.Join(dir, "b.md")
	f3 := filepath.Join(dir, "c.md")
	writeFile(t, f1)
	writeFile(t, f2)
	writeFile(t, f3)

	// Only select 2 of 3
	entries, err := vault.Compact([]string{f1, f2})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 file entries, got %d: %v", len(entries), entries)
	}
	for _, e := range entries {
		if e.Type != vault.EntryTypeFile {
			t.Errorf("expected file entry, got %s", e.Type)
		}
	}
}

func TestCompact_RecursiveDir(t *testing.T) {
	dir := testutil.TempDir(t)
	sub := filepath.Join(dir, "sub")
	f1 := filepath.Join(dir, "top.md")
	f2 := filepath.Join(sub, "deep.md")
	writeFile(t, f1)
	writeFile(t, f2)

	entries, err := vault.Compact([]string{f1, f2})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 recursive dir entry, got %d: %v", len(entries), entries)
	}
	e := entries[0]
	if e.Type != vault.EntryTypeDir {
		t.Errorf("expected dir, got %s", e.Type)
	}
	if e.Path != dir {
		t.Errorf("expected path %s, got %s", dir, e.Path)
	}
	if !e.Recursive {
		t.Error("expected recursive=true")
	}
}

func TestCompact_RecursiveDir_Partial(t *testing.T) {
	dir := testutil.TempDir(t)
	sub := filepath.Join(dir, "sub")
	f1 := filepath.Join(dir, "top.md")
	f2 := filepath.Join(sub, "deep1.md")
	f3 := filepath.Join(sub, "deep2.md")
	writeFile(t, f1)
	writeFile(t, f2)
	writeFile(t, f3)

	// Select top-level file and only one subdir file
	entries, err := vault.Compact([]string{f1, f2})
	if err != nil {
		t.Fatal(err)
	}
	// Expect: flat dir entry for top dir (f1 is the only .md there, fully covered)
	// + individual file entry for f2 (subdir partially covered)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d: %v", len(entries), entries)
	}

	var dirEntry, fileEntry *vault.Entry
	for i := range entries {
		switch entries[i].Type {
		case vault.EntryTypeDir:
			dirEntry = &entries[i]
		case vault.EntryTypeFile:
			fileEntry = &entries[i]
		}
	}
	if dirEntry == nil {
		t.Fatal("expected a dir entry")
	}
	if fileEntry == nil {
		t.Fatal("expected a file entry")
	}
	if dirEntry.Path != dir {
		t.Errorf("dir entry path: want %s got %s", dir, dirEntry.Path)
	}
	if dirEntry.Recursive {
		t.Error("top-level dir should not be recursive")
	}
	if fileEntry.Path != f2 {
		t.Errorf("file entry path: want %s got %s", f2, fileEntry.Path)
	}
}

func TestCompact_Mixed(t *testing.T) {
	root := testutil.TempDir(t)
	// dirA: fully selected → dir entry
	dirA := filepath.Join(root, "a")
	a1 := filepath.Join(dirA, "x.md")
	a2 := filepath.Join(dirA, "y.md")
	writeFile(t, a1)
	writeFile(t, a2)

	// dirB: partially selected → individual files
	dirB := filepath.Join(root, "b")
	b1 := filepath.Join(dirB, "p.md")
	b2 := filepath.Join(dirB, "q.md")
	writeFile(t, b1)
	writeFile(t, b2)

	entries, err := vault.Compact([]string{a1, a2, b1}) // b2 not selected
	if err != nil {
		t.Fatal(err)
	}
	// Expect: 1 dir entry for dirA + 1 file entry for b1
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d: %v", len(entries), entries)
	}

	var dirCount, fileCount int
	for _, e := range entries {
		switch e.Type {
		case vault.EntryTypeDir:
			dirCount++
			if e.Path != dirA {
				t.Errorf("unexpected dir entry path: %s", e.Path)
			}
		case vault.EntryTypeFile:
			fileCount++
			if e.Path != b1 {
				t.Errorf("unexpected file entry path: %s", e.Path)
			}
		}
	}
	if dirCount != 1 || fileCount != 1 {
		t.Errorf("expected 1 dir + 1 file, got %d dir + %d file", dirCount, fileCount)
	}
}

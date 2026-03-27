package vault_test

import (
	"path/filepath"
	"testing"

	"github.com/matteoggl/pickaxe/internal/testutil"
	"github.com/matteoggl/pickaxe/internal/vault"
)

// TestReader_SatisfiesInterface verifies that *vault.Reader is assignable to VaultReader
// (compile-time check via local interface mirror).
type vaultReader interface {
	ListFiles() ([]vault.ResolvedFile, error)
	ReadFile(name string) (string, error)
}

var _ vaultReader = (*vault.Reader)(nil)

func TestReader_RoundTrip(t *testing.T) {
	dir := testutil.TempDir(t)
	v, err := vault.Init(dir)
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, filePath, "hello world")
	if err := v.AddFile(filePath, "note"); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	r := vault.NewReader(dir)

	files, err := r.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if len(files) != 1 || files[0].Name != "note" {
		t.Errorf("ListFiles: got %v", files)
	}

	content, err := r.ReadFile("note")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if content != "hello world" {
		t.Errorf("ReadFile: got %q, want %q", content, "hello world")
	}
}

func TestReader_ReloadsAfterChange(t *testing.T) {
	dir := testutil.TempDir(t)
	if _, err := vault.Init(dir); err != nil {
		t.Fatal(err)
	}
	r := vault.NewReader(dir)

	files, err := r.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles initially: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected 0 files initially, got %d", len(files))
	}

	// Add a file to the vault on disk
	v, err := vault.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, filePath, "hello")
	if err := v.AddFile(filePath, "note"); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	// Reader should see the new entry on next call
	files, err = r.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles after change: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file after change, got %d", len(files))
	}
}

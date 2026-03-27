package vault_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/matteo/pickaxe/internal/testutil"
	"github.com/matteo/pickaxe/internal/vault"
)

// --- Open / Init ---

func TestInit_CreatesFile(t *testing.T) {
	dir := testutil.TempDir(t)
	v, err := vault.Init(dir)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if len(v.Entries()) != 0 {
		t.Errorf("expected 0 entries, got %d", len(v.Entries()))
	}
}

func TestInit_AlreadyExists(t *testing.T) {
	dir := testutil.TempDir(t)
	if _, err := vault.Init(dir); err != nil {
		t.Fatal(err)
	}
	_, err := vault.Init(dir)
	if !errors.Is(err, vault.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestOpen_NotInitialized(t *testing.T) {
	dir := testutil.TempDir(t)
	_, err := vault.Open(dir)
	if !errors.Is(err, vault.ErrNotInitialized) {
		t.Errorf("expected ErrNotInitialized, got %v", err)
	}
}

func TestOpen_CorruptJSON(t *testing.T) {
	dir := testutil.TempDir(t)
	testutil.WriteFile(t, filepath.Join(dir, ".pickaxe.json"), "{not valid json")
	_, err := vault.Open(dir)
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}
	if errors.Is(err, vault.ErrNotInitialized) {
		t.Error("should not be ErrNotInitialized for corrupt JSON")
	}
}

func TestInit_Open_RoundTrip(t *testing.T) {
	dir := testutil.TempDir(t)
	v, err := vault.Init(dir)
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

	v2, err := vault.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	entries := v2.Entries()
	if len(entries) != 1 || entries[0].Name != "note" {
		t.Errorf("round-trip: got %v", entries)
	}
}

// --- DefaultName ---

func TestDefaultName_File(t *testing.T) {
	got := vault.DefaultName("/vault/projects/foo/adrs.md")
	if got != "adrs" {
		t.Errorf("got %q, want %q", got, "adrs")
	}
}

func TestDefaultName_DirTrailingSlash(t *testing.T) {
	got := vault.DefaultName("/vault/projects/foo/")
	if got != "foo" {
		t.Errorf("got %q, want %q", got, "foo")
	}
}

// --- AddFile ---

func TestAddFile_Success(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "# ADRs")

	v, _ := vault.Init(dir)
	if err := v.AddFile(filePath, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := v.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "adrs" {
		t.Errorf("name: got %q, want %q", entries[0].Name, "adrs")
	}
	if entries[0].Type != vault.EntryTypeFile {
		t.Errorf("type: got %q, want %q", entries[0].Type, vault.EntryTypeFile)
	}
}

func TestAddFile_CustomName(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "# ADRs")

	v, _ := vault.Init(dir)
	if err := v.AddFile(filePath, "my-adrs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Entries()[0].Name != "my-adrs" {
		t.Errorf("name: got %q, want %q", v.Entries()[0].Name, "my-adrs")
	}
}

func TestAddFile_NonExistentPath(t *testing.T) {
	dir := testutil.TempDir(t)
	v, _ := vault.Init(dir)
	if err := v.AddFile("/nonexistent/file.md", ""); err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

func TestAddFile_NameCollision(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "")

	v, _ := vault.Init(dir)
	if err := v.AddFile(filePath, ""); err != nil {
		t.Fatal(err)
	}
	err := v.AddFile(filePath, "")
	if !errors.Is(err, vault.ErrNameCollision) {
		t.Errorf("expected ErrNameCollision, got %v", err)
	}
}

func TestAddFile_NameCollision_AcrossTypes(t *testing.T) {
	dir := testutil.TempDir(t)
	subDir := filepath.Join(dir, "notes")
	testutil.WriteFile(t, filepath.Join(subDir, "a.md"), "")
	filePath := filepath.Join(dir, "notes.md")
	testutil.WriteFile(t, filePath, "")

	v, _ := vault.Init(dir)
	if err := v.AddDir(subDir, "notes", false); err != nil {
		t.Fatal(err)
	}
	err := v.AddFile(filePath, "notes")
	if !errors.Is(err, vault.ErrNameCollision) {
		t.Errorf("expected ErrNameCollision across types, got %v", err)
	}
}

// --- AddDir ---

func TestAddDir_Flat(t *testing.T) {
	dir := testutil.TempDir(t)
	testutil.WriteFile(t, filepath.Join(dir, "a.md"), "")
	testutil.WriteFile(t, filepath.Join(dir, "b.md"), "")
	testutil.WriteFile(t, filepath.Join(dir, "sub", "c.md"), "")

	v, _ := vault.Init(dir)
	subDir := filepath.Join(dir, "docs")
	testutil.WriteFile(t, filepath.Join(subDir, "x.md"), "")
	if err := v.AddDir(subDir, "", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Entries()[0].Type != vault.EntryTypeDir {
		t.Errorf("type: got %q, want %q", v.Entries()[0].Type, vault.EntryTypeDir)
	}
	if v.Entries()[0].Recursive {
		t.Error("expected recursive=false")
	}
}

func TestAddDir_Recursive(t *testing.T) {
	dir := testutil.TempDir(t)
	subDir := filepath.Join(dir, "notes")
	testutil.WriteFile(t, filepath.Join(subDir, "a.md"), "")
	testutil.WriteFile(t, filepath.Join(subDir, "sub", "b.md"), "")

	v, _ := vault.Init(dir)
	if err := v.AddDir(subDir, "", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !v.Entries()[0].Recursive {
		t.Error("expected recursive=true")
	}
}

func TestAddDir_NonExistentPath(t *testing.T) {
	dir := testutil.TempDir(t)
	v, _ := vault.Init(dir)
	if err := v.AddDir("/nonexistent/", "", false); err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

// --- Remove ---

func TestRemove_Success(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "")

	v, _ := vault.Init(dir)
	_ = v.AddFile(filePath, "")
	if err := v.Remove("adrs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(v.Entries()) != 0 {
		t.Errorf("expected 0 entries after remove, got %d", len(v.Entries()))
	}
}

func TestRemove_NotFound(t *testing.T) {
	dir := testutil.TempDir(t)
	v, _ := vault.Init(dir)
	err := v.Remove("nonexistent")
	if !errors.Is(err, vault.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- ListFiles / ReadFile ---

func TestListFiles_Flat(t *testing.T) {
	dir := testutil.TempDir(t)
	docsDir := filepath.Join(dir, "docs")
	testutil.WriteFile(t, filepath.Join(docsDir, "a.md"), "content a")
	testutil.WriteFile(t, filepath.Join(docsDir, "b.md"), "content b")
	testutil.WriteFile(t, filepath.Join(docsDir, "sub", "c.md"), "content c")
	testutil.WriteFile(t, filepath.Join(docsDir, "notes.txt"), "")

	v, _ := vault.Init(dir)
	_ = v.AddDir(docsDir, "myfolder", false)
	_ = v.Save()

	files, err := vault.ListFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files (flat .md only), got %d", len(files))
	}
	for _, f := range files {
		if !hasPrefix(f.Name, "myfolder/") {
			t.Errorf("name %q not prefixed with myfolder/", f.Name)
		}
	}
}

func TestListFiles_Recursive(t *testing.T) {
	dir := testutil.TempDir(t)
	docsDir := filepath.Join(dir, "docs")
	testutil.WriteFile(t, filepath.Join(docsDir, "a.md"), "")
	testutil.WriteFile(t, filepath.Join(docsDir, "sub", "b.md"), "")

	v, _ := vault.Init(dir)
	_ = v.AddDir(docsDir, "myfolder", true)
	_ = v.Save()

	files, err := vault.ListFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
}

func TestReadFile_Found(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, filePath, "hello world")

	v, _ := vault.Init(dir)
	_ = v.AddFile(filePath, "note")
	_ = v.Save()

	content, err := vault.ReadFile(dir, "note")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "hello world" {
		t.Errorf("got %q, want %q", content, "hello world")
	}
}

func TestReadFile_NotFound(t *testing.T) {
	dir := testutil.TempDir(t)
	v, _ := vault.Init(dir)
	_ = v.Save()

	_, err := vault.ReadFile(dir, "nonexistent")
	if !errors.Is(err, vault.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestReadFile_Unavailable(t *testing.T) {
	dir := testutil.TempDir(t)
	v, _ := vault.Init(dir)
	// Manually add entry pointing to non-existent file via Init+Save pattern
	_ = v.Save()
	// Write a pickaxe.json with a ghost entry
	testutil.WriteFile(t, filepath.Join(dir, ".pickaxe.json"),
		`{"version":1,"entries":[{"type":"file","path":"/nonexistent/ghost.md","name":"ghost"}]}`)

	_, err := vault.ReadFile(dir, "ghost")
	if err == nil {
		t.Fatal("expected error for unavailable file")
	}
}

func TestReadFile_NotInitialized(t *testing.T) {
	dir := testutil.TempDir(t)
	_, err := vault.ReadFile(dir, "anything")
	if !errors.Is(err, vault.ErrNotInitialized) {
		t.Errorf("expected ErrNotInitialized, got %v", err)
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

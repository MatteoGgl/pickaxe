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

// --- RemoveByPath ---

func TestRemoveByPath_Success(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "")

	v, _ := vault.Init(dir)
	_ = v.AddFile(filePath, "")
	if err := v.RemoveByPath(filePath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(v.Entries()) != 0 {
		t.Errorf("expected 0 entries after RemoveByPath, got %d", len(v.Entries()))
	}
}

func TestRemoveByPath_NotFound(t *testing.T) {
	dir := testutil.TempDir(t)
	v, _ := vault.Init(dir)
	err := v.RemoveByPath("/nonexistent/path.md")
	if !errors.Is(err, vault.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- ReplaceEntries ---

func TestReplaceEntries_ReplacesAll(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath1 := filepath.Join(dir, "file1.md")
	filePath2 := filepath.Join(dir, "file2.md")
	testutil.WriteFile(t, filePath1, "")
	testutil.WriteFile(t, filePath2, "")

	v, _ := vault.Init(dir)
	_ = v.AddFile(filePath1, "file1")
	_ = v.AddFile(filePath2, "file2")

	newEntries := []vault.Entry{
		{Type: vault.EntryTypeFile, Path: filePath1, Name: "file1"},
	}
	v.ReplaceEntries(newEntries)

	if len(v.Entries()) != 1 {
		t.Errorf("expected 1 entry after ReplaceEntries, got %d", len(v.Entries()))
	}
	if v.Entries()[0].Name != "file1" {
		t.Errorf("expected name %q, got %q", "file1", v.Entries()[0].Name)
	}
}

func TestReplaceEntries_Empty(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "file.md")
	testutil.WriteFile(t, filePath, "")

	v, _ := vault.Init(dir)
	_ = v.AddFile(filePath, "file")

	v.ReplaceEntries([]vault.Entry{})

	if len(v.Entries()) != 0 {
		t.Errorf("expected 0 entries after ReplaceEntries with empty slice, got %d", len(v.Entries()))
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

func TestReadFile_StripsFrontmatter(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, filePath, "---\ntitle: secret\ntags: [a, b]\n---\n\n# Hello")

	v, _ := vault.Init(dir)
	_ = v.AddFile(filePath, "note")
	_ = v.Save()

	content, err := vault.ReadFile(dir, "note")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "\n# Hello" {
		t.Errorf("frontmatter not stripped: got %q", content)
	}
}

func TestReadFile_KeepsFrontmatterWhenDisabled(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "note.md")
	raw := "---\ntitle: secret\n---\n\n# Hello"
	testutil.WriteFile(t, filePath, raw)

	testutil.WriteFile(t, filepath.Join(dir, ".pickaxe.json"),
		`{"version":1,"strip_frontmatter":false,"entries":[{"type":"file","path":"`+filePath+`","name":"note"}]}`)

	content, err := vault.ReadFile(dir, "note")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != raw {
		t.Errorf("frontmatter should be preserved: got %q", content)
	}
}

// --- ResolveIdentifier ---

func TestResolveIdentifier_ByName(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "")
	v, _ := vault.Init(dir)
	_ = v.AddFile(filePath, "adrs")

	name, err := v.ResolveIdentifier("adrs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "adrs" {
		t.Errorf("expected %q, got %q", "adrs", name)
	}
}

func TestResolveIdentifier_ByHashPrefix(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "")
	v, _ := vault.Init(dir)
	_ = v.AddFile(filePath, "adrs")

	hash := vault.HashEntry("adrs")
	name, err := v.ResolveIdentifier(hash[:4])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "adrs" {
		t.Errorf("expected %q, got %q", "adrs", name)
	}
}

func TestResolveIdentifier_NotFound(t *testing.T) {
	dir := testutil.TempDir(t)
	v, _ := vault.Init(dir)

	_, err := v.ResolveIdentifier("zzzzzzz")
	if !errors.Is(err, vault.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestResolveIdentifier_NameTakesPrecedence(t *testing.T) {
	dir := testutil.TempDir(t)
	// Create an entry whose name happens to look like a hash prefix of another.
	// Use two files: "alpha" and an entry named after alpha's hash prefix.
	filePath1 := filepath.Join(dir, "alpha.md")
	testutil.WriteFile(t, filePath1, "")
	hashPrefix := vault.HashEntry("alpha")[:4]
	filePath2 := filepath.Join(dir, hashPrefix+".md")
	testutil.WriteFile(t, filePath2, "")

	v, _ := vault.Init(dir)
	_ = v.AddFile(filePath1, "alpha")
	_ = v.AddFile(filePath2, hashPrefix)

	// Resolving by the hash prefix as a name should find the entry named hashPrefix exactly.
	name, err := v.ResolveIdentifier(hashPrefix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != hashPrefix {
		t.Errorf("expected exact name match %q, got %q", hashPrefix, name)
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

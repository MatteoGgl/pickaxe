package cmd_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matteo/pickaxe/cmd"
	"github.com/matteo/pickaxe/internal/testutil"
	"github.com/matteo/pickaxe/internal/tui"
	"github.com/matteo/pickaxe/internal/vault"
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

// --- AddFromPicker tests ---

func TestAddFromPicker_AddsNewSelections(t *testing.T) {
	v, dir := initVault(t)
	f1 := filepath.Join(dir, "a.md")
	f2 := filepath.Join(dir, "b.md")
	testutil.WriteFile(t, f1, "a")
	testutil.WriteFile(t, f2, "b")

	picker := fakePicker(tui.PickerResult{
		Confirmed:  true,
		Selections: []tui.Selection{{Path: f1}, {Path: f2}},
	}, nil)

	var buf bytes.Buffer
	if err := cmd.AddFromPicker(v, dir, picker, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(v.Entries()) != 2 {
		t.Errorf("expected 2 entries, got %d", len(v.Entries()))
	}
}

func TestAddFromPicker_SkipsExisting(t *testing.T) {
	v, dir := initVault(t)
	f1 := filepath.Join(dir, "a.md")
	testutil.WriteFile(t, f1, "a")
	if err := v.AddFile(f1, "a"); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	picker := fakePicker(tui.PickerResult{
		Confirmed:  true,
		Selections: []tui.Selection{{Path: f1}},
	}, nil)

	var buf bytes.Buffer
	if err := cmd.AddFromPicker(v, dir, picker, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(v.Entries()) != 1 {
		t.Errorf("expected still 1 entry, got %d", len(v.Entries()))
	}
	if !strings.Contains(buf.String(), "no new entries added") {
		t.Errorf("expected 'no new entries added' message, got: %q", buf.String())
	}
}

func TestAddFromPicker_Cancelled(t *testing.T) {
	v, dir := initVault(t)

	picker := fakePicker(tui.PickerResult{Confirmed: false}, nil)

	var buf bytes.Buffer
	if err := cmd.AddFromPicker(v, dir, picker, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(v.Entries()) != 0 {
		t.Errorf("expected 0 entries after cancel, got %d", len(v.Entries()))
	}
	if !strings.Contains(buf.String(), "no changes made") {
		t.Errorf("expected 'no changes made' message, got: %q", buf.String())
	}
}

func TestAddFromPicker_PickerError(t *testing.T) {
	v, dir := initVault(t)
	picker := fakePicker(tui.PickerResult{}, errors.New("tui exploded"))

	var buf bytes.Buffer
	err := cmd.AddFromPicker(v, dir, picker, &buf)
	if err == nil {
		t.Fatal("expected error from picker, got nil")
	}
}

func TestAddFromPicker_EmptySelections(t *testing.T) {
	v, dir := initVault(t)
	picker := fakePicker(tui.PickerResult{Confirmed: true, Selections: nil}, nil)

	var buf bytes.Buffer
	if err := cmd.AddFromPicker(v, dir, picker, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no changes made") {
		t.Errorf("expected 'no changes made' message, got: %q", buf.String())
	}
}

func TestAddFromPicker_MixedNewAndExisting(t *testing.T) {
	v, dir := initVault(t)
	f1 := filepath.Join(dir, "a.md")
	f2 := filepath.Join(dir, "b.md")
	testutil.WriteFile(t, f1, "a")
	testutil.WriteFile(t, f2, "b")
	if err := v.AddFile(f1, "a"); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	picker := fakePicker(tui.PickerResult{
		Confirmed:  true,
		Selections: []tui.Selection{{Path: f1}, {Path: f2}},
	}, nil)

	var buf bytes.Buffer
	if err := cmd.AddFromPicker(v, dir, picker, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(v.Entries()) != 2 {
		t.Errorf("expected 2 entries, got %d", len(v.Entries()))
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

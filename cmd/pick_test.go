package cmd_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matteoggl/pickaxe/cmd"
	"github.com/matteoggl/pickaxe/internal/testutil"
	"github.com/matteoggl/pickaxe/internal/tui"
	"github.com/matteoggl/pickaxe/internal/vault"
)

func TestPickFromVault_AddsNewFile(t *testing.T) {
	v, dir := initVault(t)
	f1 := filepath.Join(dir, "a.md")
	testutil.WriteFile(t, f1, "a")

	picker := fakePicker(tui.PickerResult{
		Confirmed:  true,
		Selections: []tui.Selection{{Path: f1}},
	}, nil)

	var buf bytes.Buffer
	if err := cmd.PickFromVault(v, dir, picker, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(v.Entries()) != 1 {
		t.Errorf("expected 1 entry, got %d", len(v.Entries()))
	}
	if !strings.Contains(buf.String(), "added:") {
		t.Errorf("expected 'added:' in output, got: %q", buf.String())
	}
}

func TestPickFromVault_RemovesEntry(t *testing.T) {
	v, dir := initVault(t)
	f1 := filepath.Join(dir, "a.md")
	testutil.WriteFile(t, f1, "a")
	if err := v.AddFile(f1, ""); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	picker := fakePicker(tui.PickerResult{
		Confirmed:  true,
		Selections: nil,
	}, nil)

	var buf bytes.Buffer
	if err := cmd.PickFromVault(v, dir, picker, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(v.Entries()) != 0 {
		t.Errorf("expected 0 entries, got %d", len(v.Entries()))
	}
	if !strings.Contains(buf.String(), "removed:") {
		t.Errorf("expected 'removed:' in output, got: %q", buf.String())
	}
}

func TestPickFromVault_MixedAddRemove(t *testing.T) {
	v, dir := initVault(t)
	f1 := filepath.Join(dir, "a.md")
	f2 := filepath.Join(dir, "b.md")
	testutil.WriteFile(t, f1, "a")
	testutil.WriteFile(t, f2, "b")
	if err := v.AddFile(f1, ""); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	picker := fakePicker(tui.PickerResult{
		Confirmed:  true,
		Selections: []tui.Selection{{Path: f2}},
	}, nil)

	var buf bytes.Buffer
	if err := cmd.PickFromVault(v, dir, picker, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := v.Entries()
	if len(entries) != 1 || entries[0].Path != f2 {
		t.Errorf("expected only f2, got %v", entries)
	}
	out := buf.String()
	if !strings.Contains(out, "added:") || !strings.Contains(out, "removed:") {
		t.Errorf("expected both added and removed in output, got: %q", out)
	}
}

func TestPickFromVault_NoChanges(t *testing.T) {
	v, dir := initVault(t)
	f1 := filepath.Join(dir, "a.md")
	testutil.WriteFile(t, f1, "a")
	if err := v.AddFile(f1, ""); err != nil {
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
	if err := cmd.PickFromVault(v, dir, picker, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no changes") {
		t.Errorf("expected 'no changes' in output, got: %q", buf.String())
	}
}

func TestPickFromVault_Cancelled(t *testing.T) {
	v, dir := initVault(t)
	f1 := filepath.Join(dir, "a.md")
	testutil.WriteFile(t, f1, "a")
	if err := v.AddFile(f1, ""); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}

	picker := fakePicker(tui.PickerResult{Confirmed: false}, nil)

	var buf bytes.Buffer
	if err := cmd.PickFromVault(v, dir, picker, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(v.Entries()) != 1 {
		t.Errorf("expected vault unchanged (1 entry), got %d", len(v.Entries()))
	}
	if !strings.Contains(buf.String(), "no changes made") {
		t.Errorf("expected 'no changes made' in output, got: %q", buf.String())
	}
}

func TestPickFromVault_PickerError(t *testing.T) {
	v, dir := initVault(t)

	picker := fakePicker(tui.PickerResult{}, errors.New("tui exploded"))

	var buf bytes.Buffer
	err := cmd.PickFromVault(v, dir, picker, &buf)
	if err == nil {
		t.Fatal("expected error from picker, got nil")
	}
}

func TestPickFromVault_DirCollapses(t *testing.T) {
	v, dir := initVault(t)
	subDir := filepath.Join(dir, "notes")
	f1 := filepath.Join(subDir, "a.md")
	f2 := filepath.Join(subDir, "b.md")
	testutil.WriteFile(t, f1, "a")
	testutil.WriteFile(t, f2, "b")

	picker := fakePicker(tui.PickerResult{
		Confirmed: true,
		Selections: []tui.Selection{
			{Path: f1},
			{Path: f2},
		},
	}, nil)

	var buf bytes.Buffer
	if err := cmd.PickFromVault(v, dir, picker, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := v.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 dir entry after compaction, got %d: %v", len(entries), entries)
	}
	if entries[0].Type != vault.EntryTypeDir {
		t.Errorf("expected dir entry, got %v", entries[0].Type)
	}
	if entries[0].Path != subDir {
		t.Errorf("expected path %q, got %q", subDir, entries[0].Path)
	}
}

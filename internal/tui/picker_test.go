package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func pressKey(m *Model, key tea.KeyMsg) *Model {
	updated, _ := m.Update(key)
	return updated.(*Model)
}

func pressSpace(m *Model) *Model {
	return pressKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
}

func pressEnter(m *Model) *Model {
	return pressKey(m, tea.KeyMsg{Type: tea.KeyEnter})
}

func pressDown(m *Model) *Model {
	return pressKey(m, tea.KeyMsg{Type: tea.KeyDown})
}

func pressCtrlD(m *Model) *Model {
	return pressKey(m, tea.KeyMsg{Type: tea.KeyCtrlD})
}

// setupNestedDir creates:
//
//	root/
//	  parent/
//	    child/
//	      deep.md
//	    shallow.md
func setupNestedDir(t *testing.T) (root, parentDir, shallowFile, deepFile string) {
	t.Helper()
	root = t.TempDir()
	parentDir = filepath.Join(root, "parent")
	childDir := filepath.Join(parentDir, "child")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatal(err)
	}
	shallowFile = filepath.Join(parentDir, "shallow.md")
	deepFile = filepath.Join(childDir, "deep.md")
	if err := os.WriteFile(shallowFile, []byte("# shallow"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(deepFile, []byte("# deep"), 0644); err != nil {
		t.Fatal(err)
	}
	return root, parentDir, shallowFile, deepFile
}

// setupDirs creates:
//
//	root/
//	  dirA/
//	    fileA.md
//	  dirB/
//	    fileB.md
func setupDirs(t *testing.T) (root, fileA, fileB string) {
	t.Helper()
	root = t.TempDir()
	dirA := filepath.Join(root, "dirA")
	dirB := filepath.Join(root, "dirB")
	if err := os.MkdirAll(dirA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dirB, 0755); err != nil {
		t.Fatal(err)
	}
	fileA = filepath.Join(dirA, "fileA.md")
	fileB = filepath.Join(dirB, "fileB.md")
	if err := os.WriteFile(fileA, []byte("# A"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fileB, []byte("# B"), 0644); err != nil {
		t.Fatal(err)
	}
	return root, fileA, fileB
}

// TestPicker_SelectionPersistsAcrossNav selects a file in dirA, navigates to
// dirB and back, then verifies the file in dirA is still checked.
func TestPicker_SelectionPersistsAcrossNav(t *testing.T) {
	root, fileA, _ := setupDirs(t)

	m := NewPicker(root, nil, nil)

	// Root shows: dirA/, dirB/ (sorted alphabetically, dirs first)
	// cursor=0 → dirA
	m = pressEnter(m) // navigate into dirA

	// dirA shows: ../, fileA.md
	// cursor=0 → "..", move down to fileA.md
	m = pressDown(m)  // cursor=1 → fileA.md
	m = pressSpace(m) // select fileA.md

	if !m.selected[fileA] {
		t.Fatalf("expected fileA to be selected, selected map: %v", m.selected)
	}

	// Navigate back to root
	m.cursor = 0
	m = pressEnter(m) // navigate to ".."

	// Navigate into dirB
	// cursor=0 → dirA, cursor=1 → dirB
	m = pressDown(m)  // cursor=1 → dirB
	m = pressEnter(m) // navigate into dirB

	// Navigate back to root
	m.cursor = 0
	m = pressEnter(m) // navigate to ".."

	// Navigate into dirA again
	m.cursor = 0
	m = pressEnter(m) // navigate into dirA

	// fileA.md should still be checked
	m = pressDown(m) // cursor=1 → fileA.md
	visible := m.visibleItems()
	if m.cursor >= len(visible) {
		t.Fatalf("cursor out of range: cursor=%d len=%d", m.cursor, len(visible))
	}
	if !visible[m.cursor].checked {
		t.Errorf("fileA.md should still be checked after navigating away and back")
	}
}

// TestPicker_CtrlDCollectsFromAllDirs selects files in two different
// directories and verifies both appear in the ctrl+d result.
func TestPicker_CtrlDCollectsFromAllDirs(t *testing.T) {
	root, fileA, fileB := setupDirs(t)

	m := NewPicker(root, nil, nil)

	// Navigate into dirA, select fileA
	m = pressEnter(m) // into dirA
	m = pressDown(m)  // cursor → fileA.md
	m = pressSpace(m) // select
	m.cursor = 0
	m = pressEnter(m) // back to root

	// Navigate into dirB, select fileB
	m = pressDown(m)  // cursor → dirB
	m = pressEnter(m) // into dirB
	m = pressDown(m)  // cursor → fileB.md
	m = pressSpace(m) // select

	// Confirm
	m = pressCtrlD(m)

	result := m.Result()
	if !result.Confirmed {
		t.Fatal("expected Confirmed=true")
	}

	paths := make(map[string]bool)
	for _, sel := range result.Selections {
		paths[sel.Path] = true
	}

	if !paths[fileA] {
		t.Errorf("expected fileA in result, got %v", result.Selections)
	}
	if !paths[fileB] {
		t.Errorf("expected fileB in result, got %v", result.Selections)
	}
}

// TestPicker_DirToggleSelectsAllFiles verifies that pressing space on a dir
// adds all its .md files to m.selected and marks the dir checked.
func TestPicker_DirToggleSelectsAllFiles(t *testing.T) {
	root, fileA, _ := setupDirs(t)
	m := NewPicker(root, nil, nil)

	// cursor=0 → dirA
	m = pressSpace(m)

	if !m.selected[fileA] {
		t.Errorf("expected fileA in selected after toggling dirA, selected=%v", m.selected)
	}
	// dir item should be checked, not partial
	if !m.items[m.cursor].checked {
		t.Errorf("expected dirA item to be checked")
	}
	if m.items[m.cursor].partial {
		t.Errorf("expected dirA item to not be partial")
	}
}

// TestPicker_DirToggleDeselectsAllFiles verifies that pressing space on a
// fully-selected dir removes all its .md files from m.selected.
func TestPicker_DirToggleDeselectsAllFiles(t *testing.T) {
	root, fileA, _ := setupDirs(t)
	m := NewPicker(root, nil, nil)

	// Select dirA first
	m = pressSpace(m)
	if !m.selected[fileA] {
		t.Fatalf("precondition: fileA should be selected")
	}

	// Deselect dirA
	m = pressSpace(m)

	if m.selected[fileA] {
		t.Errorf("expected fileA to be removed from selected after toggling dirA off")
	}
	if m.items[m.cursor].checked {
		t.Errorf("expected dirA item to be unchecked")
	}
	if m.items[m.cursor].partial {
		t.Errorf("expected dirA item to not be partial")
	}
}

// TestPicker_DirPartialState selects one file in dirA (which has one file),
// then uses a multi-file dir to verify partial state.
func TestPicker_DirPartialState(t *testing.T) {
	root := t.TempDir()
	dirA := filepath.Join(root, "dirA")
	if err := os.MkdirAll(dirA, 0755); err != nil {
		t.Fatal(err)
	}
	file1 := filepath.Join(dirA, "file1.md")
	file2 := filepath.Join(dirA, "file2.md")
	for _, f := range []string{file1, file2} {
		if err := os.WriteFile(f, []byte("# x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Pre-select only file1
	m := NewPicker(root, map[string]bool{file1: true}, nil)

	// At root, cursor=0 → dirA
	item := m.items[0]
	if item.checked {
		t.Errorf("expected dirA not fully checked (only 1 of 2 files selected)")
	}
	if !item.partial {
		t.Errorf("expected dirA to be partial (1 of 2 files selected)")
	}
}

// TestPicker_PartialDirTogglesToFull verifies that a partial dir becomes
// fully selected (not deselected) when toggled.
func TestPicker_PartialDirTogglesToFull(t *testing.T) {
	root := t.TempDir()
	dirA := filepath.Join(root, "dirA")
	if err := os.MkdirAll(dirA, 0755); err != nil {
		t.Fatal(err)
	}
	file1 := filepath.Join(dirA, "file1.md")
	file2 := filepath.Join(dirA, "file2.md")
	for _, f := range []string{file1, file2} {
		if err := os.WriteFile(f, []byte("# x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Pre-select only file1 → dirA is partial
	m := NewPicker(root, map[string]bool{file1: true}, nil)

	// cursor=0 → dirA (partial); press space → should become fully selected
	m = pressSpace(m)

	if !m.selected[file1] {
		t.Errorf("expected file1 still selected after toggle-to-full")
	}
	if !m.selected[file2] {
		t.Errorf("expected file2 selected after toggle-to-full on partial dir")
	}
	if !m.items[0].checked {
		t.Errorf("expected dirA to be checked after toggle-to-full")
	}
	if m.items[0].partial {
		t.Errorf("expected dirA not partial after toggle-to-full")
	}
}

// TestPicker_CtrlDOnlyEmitsFiles verifies that confirming with a selected dir
// produces file paths (not dir paths) in the result.
func TestPicker_CtrlDOnlyEmitsFiles(t *testing.T) {
	root, fileA, _ := setupDirs(t)
	m := NewPicker(root, nil, nil)

	// Select dirA (cursor=0)
	m = pressSpace(m)
	m = pressCtrlD(m)

	result := m.Result()
	if !result.Confirmed {
		t.Fatal("expected Confirmed=true")
	}

	for _, sel := range result.Selections {
		if sel.IsDir {
			t.Errorf("result should not contain dir selections, got %v", sel)
		}
		info, err := os.Stat(sel.Path)
		if err != nil {
			t.Errorf("selection path does not exist: %v", sel.Path)
			continue
		}
		if info.IsDir() {
			t.Errorf("selection path is a directory: %v", sel.Path)
		}
	}

	paths := make(map[string]bool)
	for _, sel := range result.Selections {
		paths[sel.Path] = true
	}
	if !paths[fileA] {
		t.Errorf("expected fileA in result, got %v", result.Selections)
	}
}

// TestPicker_NestedDirToggle verifies that toggling a parent dir selects all
// .md files at all depths.
func TestPicker_NestedDirToggle(t *testing.T) {
	root, parentDir, shallowFile, deepFile := setupNestedDir(t)
	_ = parentDir
	m := NewPicker(root, nil, nil)

	// cursor=0 → parent/
	m = pressSpace(m)

	if !m.selected[shallowFile] {
		t.Errorf("expected shallowFile selected, selected=%v", m.selected)
	}
	if !m.selected[deepFile] {
		t.Errorf("expected deepFile (nested) selected, selected=%v", m.selected)
	}
	if !m.items[0].checked {
		t.Errorf("expected parent dir item to be checked")
	}
}

func pressW(m *Model) *Model {
	return pressKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
}

// TestPicker_WritableToggle_File checks that pressing w on a checked file
// toggles writable on then off, and ctrl+d reflects the final state.
func TestPicker_WritableToggle_File(t *testing.T) {
	root, fileA, _ := setupDirs(t)
	m := NewPicker(root, nil, nil)

	// Navigate into dirA, select fileA
	m = pressEnter(m) // into dirA
	m = pressDown(m)  // cursor → fileA.md
	m = pressSpace(m) // select

	if !m.selected[fileA] {
		t.Fatal("precondition: fileA should be selected")
	}

	// w → writable on
	m = pressW(m)
	if !m.writable[fileA] {
		t.Errorf("expected writable[fileA]=true after first w press")
	}

	// w again → writable off
	m = pressW(m)
	if m.writable[fileA] {
		t.Errorf("expected writable[fileA]=false after second w press")
	}

	// Turn writable back on and confirm
	m = pressW(m)
	m = pressCtrlD(m)

	result := m.Result()
	if !result.Confirmed {
		t.Fatal("expected Confirmed=true")
	}
	var found *Selection
	for i, sel := range result.Selections {
		if sel.Path == fileA {
			found = &result.Selections[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("fileA not in result: %v", result.Selections)
	}
	if !found.Writable {
		t.Errorf("expected Selection.Writable=true, got false")
	}
}

// TestPicker_WritableOnUnchecked_SelectsAndMarksWritable verifies that pressing
// w on an unchecked file selects it and marks it writable in one action.
func TestPicker_WritableOnUnchecked_SelectsAndMarksWritable(t *testing.T) {
	root, fileA, _ := setupDirs(t)
	m := NewPicker(root, nil, nil)

	// Navigate into dirA — fileA is unchecked
	m = pressEnter(m) // into dirA
	m = pressDown(m)  // cursor → fileA.md (unchecked)

	m = pressW(m)

	if !m.selected[fileA] {
		t.Errorf("expected fileA to be selected after w on unchecked file")
	}
	if !m.writable[fileA] {
		t.Errorf("expected writable[fileA]=true after w on unchecked file")
	}
}

// TestPicker_WritableOnUncheckedDir_SelectsAllAndMarksWritable verifies that
// pressing w on an unchecked dir selects all md files and marks them writable.
func TestPicker_WritableOnUncheckedDir_SelectsAllAndMarksWritable(t *testing.T) {
	root := t.TempDir()
	dirA := filepath.Join(root, "dirA")
	if err := os.MkdirAll(dirA, 0755); err != nil {
		t.Fatal(err)
	}
	file1 := filepath.Join(dirA, "file1.md")
	file2 := filepath.Join(dirA, "file2.md")
	for _, f := range []string{file1, file2} {
		if err := os.WriteFile(f, []byte("# x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	m := NewPicker(root, nil, nil)

	// cursor=0 → dirA (unchecked); press w → all files selected + writable
	m = pressW(m)

	if !m.selected[file1] {
		t.Errorf("expected file1 selected after w on unchecked dir")
	}
	if !m.selected[file2] {
		t.Errorf("expected file2 selected after w on unchecked dir")
	}
	if !m.writable[file1] {
		t.Errorf("expected writable[file1]=true after w on unchecked dir")
	}
	if !m.writable[file2] {
		t.Errorf("expected writable[file2]=true after w on unchecked dir")
	}
	if !m.items[0].checked {
		t.Errorf("expected dirA item to be checked")
	}
}

// TestPicker_WritableToggle_Dir verifies that pressing w on a dir toggles
// writable for all checked files within it.
func TestPicker_WritableToggle_Dir(t *testing.T) {
	root := t.TempDir()
	dirA := filepath.Join(root, "dirA")
	if err := os.MkdirAll(dirA, 0755); err != nil {
		t.Fatal(err)
	}
	file1 := filepath.Join(dirA, "file1.md")
	file2 := filepath.Join(dirA, "file2.md")
	for _, f := range []string{file1, file2} {
		if err := os.WriteFile(f, []byte("# x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Pre-select both files
	m := NewPicker(root, map[string]bool{file1: true, file2: true}, nil)

	// cursor=0 → dirA (checked); press w → all checked files become writable
	m = pressW(m)
	if !m.writable[file1] {
		t.Errorf("expected writable[file1]=true after w on dir")
	}
	if !m.writable[file2] {
		t.Errorf("expected writable[file2]=true after w on dir")
	}

	// Press w again → all become non-writable
	m = pressW(m)
	if m.writable[file1] {
		t.Errorf("expected writable[file1]=false after second w on dir")
	}
	if m.writable[file2] {
		t.Errorf("expected writable[file2]=false after second w on dir")
	}
}

// TestPicker_ParentDirNotToggleable verifies that pressing space on ".." does
// not toggle its checked state.
func TestPicker_ParentDirNotToggleable(t *testing.T) {
	root, _, _ := setupDirs(t)
	m := NewPicker(root, nil, nil)

	// Navigate into dirA so ".." appears at cursor=0
	m = pressEnter(m) // into dirA

	visible := m.visibleItems()
	if len(visible) == 0 || visible[0].name != ".." {
		t.Fatalf("expected '..' at cursor=0, got %v", visible)
	}
	// cursor=0 → ".."
	before := visible[0].checked
	m = pressSpace(m)
	visible = m.visibleItems()
	if visible[0].checked != before {
		t.Errorf("'..' checked state changed after space: was %v, now %v", before, visible[0].checked)
	}
}

// TestPicker_DeselectClearsWritable verifies that deselecting a file also
// clears its writable state so re-selecting starts clean.
func TestPicker_DeselectClearsWritable(t *testing.T) {
	root, fileA, _ := setupDirs(t)
	m := NewPicker(root, nil, nil)

	m = pressEnter(m) // into dirA
	m = pressDown(m)  // cursor → fileA.md
	m = pressSpace(m) // select
	m = pressW(m)     // mark writable

	if !m.writable[fileA] {
		t.Fatal("precondition: writable[fileA] should be true")
	}

	m = pressSpace(m) // deselect
	m = pressSpace(m) // re-select

	if m.writable[fileA] {
		t.Errorf("writable[fileA] should be false after deselect+reselect, got true")
	}
}

// TestPicker_DirDeselectClearsWritable verifies that toggling a dir off clears
// writable state for all its files.
func TestPicker_DirDeselectClearsWritable(t *testing.T) {
	root, fileA, _ := setupDirs(t)
	m := NewPicker(root, nil, nil)

	// cursor=0 → dirA; select all then mark writable
	m = pressSpace(m) // select all
	m = pressW(m)     // mark writable

	if !m.writable[fileA] {
		t.Fatal("precondition: writable[fileA] should be true")
	}

	m = pressSpace(m) // deselect all (dir toggle off)
	m = pressSpace(m) // re-select all

	if m.writable[fileA] {
		t.Errorf("writable[fileA] should be false after dir deselect+reselect, got true")
	}
}

// TestPicker_UntogglePreSelectedExcludesFromResult pre-selects a file, then
// untoggling it via space should remove it from the ctrl+d result.
func TestPicker_UntogglePreSelectedExcludesFromResult(t *testing.T) {
	root, fileA, _ := setupDirs(t)

	preSelected := map[string]bool{fileA: true}
	m := NewPicker(root, preSelected, nil)

	// Navigate into dirA
	m = pressEnter(m) // into dirA

	// fileA.md should be pre-checked; untoggle it
	m = pressDown(m)  // cursor → fileA.md
	m = pressSpace(m) // untoggle (was true → false)

	if m.selected[fileA] {
		t.Fatalf("expected fileA to be deselected, selected map: %v", m.selected)
	}

	// Confirm
	m = pressCtrlD(m)

	result := m.Result()
	if !result.Confirmed {
		t.Fatal("expected Confirmed=true")
	}

	for _, sel := range result.Selections {
		if sel.Path == fileA {
			t.Errorf("fileA should not appear in result after being untoggled")
		}
	}
}

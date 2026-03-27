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

	m := NewPicker(root, nil)

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

	m := NewPicker(root, nil)

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

// TestPicker_UntogglePreSelectedExcludesFromResult pre-selects a file, then
// untoggling it via space should remove it from the ctrl+d result.
func TestPicker_UntogglePreSelectedExcludesFromResult(t *testing.T) {
	root, fileA, _ := setupDirs(t)

	preSelected := map[string]bool{fileA: true}
	m := NewPicker(root, preSelected)

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

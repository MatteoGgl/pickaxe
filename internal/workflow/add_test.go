package workflow_test

import (
	"testing"

	"github.com/matteo/pickaxe/internal/vault"
	"github.com/matteo/pickaxe/internal/workflow"
)

func entries(paths ...string) []vault.Entry {
	result := make([]vault.Entry, len(paths))
	for i, p := range paths {
		result[i] = vault.Entry{Type: vault.EntryTypeFile, Path: p, Name: "e" + string(rune('0'+i))}
	}
	return result
}

func sels(paths ...string) []workflow.Selection {
	result := make([]workflow.Selection, len(paths))
	for i, p := range paths {
		result[i] = workflow.Selection{Path: p}
	}
	return result
}

func TestDiffSelections_AllNew(t *testing.T) {
	plan := workflow.DiffSelections(entries(), sels("/a", "/b"))
	if len(plan.Additions) != 2 {
		t.Errorf("expected 2 additions, got %d", len(plan.Additions))
	}
}

func TestDiffSelections_AllExisting(t *testing.T) {
	plan := workflow.DiffSelections(entries("/a", "/b"), sels("/a", "/b"))
	if len(plan.Additions) != 0 {
		t.Errorf("expected 0 additions, got %d", len(plan.Additions))
	}
}

func TestDiffSelections_Mixed(t *testing.T) {
	plan := workflow.DiffSelections(entries("/a"), sels("/a", "/b"))
	if len(plan.Additions) != 1 {
		t.Errorf("expected 1 addition, got %d", len(plan.Additions))
	}
	if plan.Additions[0].Path != "/b" {
		t.Errorf("expected /b, got %s", plan.Additions[0].Path)
	}
}

func TestDiffSelections_EmptySelections(t *testing.T) {
	plan := workflow.DiffSelections(entries("/a"), sels())
	if len(plan.Additions) != 0 {
		t.Errorf("expected 0 additions, got %d", len(plan.Additions))
	}
}

func TestDiffSelections_BothEmpty(t *testing.T) {
	plan := workflow.DiffSelections(entries(), sels())
	if len(plan.Additions) != 0 {
		t.Errorf("expected 0 additions, got %d", len(plan.Additions))
	}
}

func TestBuildPreSelected(t *testing.T) {
	e := entries("/a", "/b")
	m := workflow.BuildPreSelected(e)
	if !m["/a"] || !m["/b"] {
		t.Errorf("expected /a and /b in preSelected, got %v", m)
	}
	if m["/c"] {
		t.Error("unexpected /c in preSelected")
	}
}

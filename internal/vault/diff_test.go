package vault_test

import (
	"testing"

	"github.com/matteo/pickaxe/internal/vault"
)

func TestBuildPreSelected(t *testing.T) {
	entries := []vault.Entry{
		{Type: vault.EntryTypeFile, Path: "/a", Name: "a"},
		{Type: vault.EntryTypeFile, Path: "/b", Name: "b"},
	}
	m := vault.BuildPreSelected(entries)
	if !m["/a"] || !m["/b"] {
		t.Errorf("expected /a and /b in preSelected, got %v", m)
	}
	if m["/c"] {
		t.Error("unexpected /c in preSelected")
	}
}

func TestBuildPreSelected_Empty(t *testing.T) {
	m := vault.BuildPreSelected(nil)
	if len(m) != 0 {
		t.Errorf("expected empty map, got %v", m)
	}
}

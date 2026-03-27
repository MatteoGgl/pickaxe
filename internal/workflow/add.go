package workflow

import "github.com/matteo/pickaxe/internal/vault"

// Selection is a single item chosen by the user in the TUI picker.
type Selection struct {
	Path  string
	IsDir bool
}

// DiffPlan describes what should change given existing entries and new selections.
type DiffPlan struct {
	Additions []Selection
}

// BuildPreSelected builds a set of already-registered paths for use as TUI pre-selection.
func BuildPreSelected(entries []vault.Entry) map[string]bool {
	m := make(map[string]bool, len(entries))
	for _, e := range entries {
		m[e.Path] = true
	}
	return m
}

// DiffSelections computes what needs to be added given existing entries and new selections.
// Currently only computes additions (no removals on deselect — preserves current behaviour).
func DiffSelections(entries []vault.Entry, selections []Selection) DiffPlan {
	existing := BuildPreSelected(entries)
	var additions []Selection
	for _, sel := range selections {
		if !existing[sel.Path] {
			additions = append(additions, sel)
		}
	}
	return DiffPlan{Additions: additions}
}

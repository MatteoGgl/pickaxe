package vault

// BuildPreSelected builds a set of already-registered paths for TUI pre-selection.
func BuildPreSelected(entries []Entry) map[string]bool {
	m := make(map[string]bool, len(entries))
	for _, e := range entries {
		m[e.Path] = true
	}
	return m
}

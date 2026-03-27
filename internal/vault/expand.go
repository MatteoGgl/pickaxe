package vault

// ExpandEntries returns a map of all concrete .md file paths covered by the given entries.
// File entries contribute their path directly.
// Dir entries are expanded (flat or recursive depending on entry.Recursive).
// The boolean value is always true (the map is used as a set).
// Unavailable files are skipped.
func ExpandEntries(entries []Entry) (map[string]bool, error) {
	result := make(map[string]bool)

	for _, entry := range entries {
		files, err := enumerateFiles(entry)
		if err != nil {
			return nil, err
		}
		for _, rf := range files {
			if !rf.Unavailable {
				result[rf.Path] = true
			}
		}
	}

	return result, nil
}

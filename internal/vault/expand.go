package vault

// ExpandEntries returns two maps: paths (all concrete .md file paths) and
// writable (paths whose parent entry has Writable=true). Unavailable files are skipped.
func ExpandEntries(entries []Entry) (paths map[string]bool, writable map[string]bool, err error) {
	paths = make(map[string]bool)
	writable = make(map[string]bool)

	for _, entry := range entries {
		files, err := enumerateFiles(entry)
		if err != nil {
			return nil, nil, err
		}
		for _, rf := range files {
			if !rf.Unavailable {
				paths[rf.Path] = true
				if rf.Writable {
					writable[rf.Path] = true
				}
			}
		}
	}

	return paths, writable, nil
}

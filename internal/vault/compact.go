package vault

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Compact takes a set of selected file paths and an optional writablePaths map,
// and returns the most compact []Entry representation. For each directory, if ALL
// .md files are selected, collapse to a dir entry. If all subdirs are also fully
// selected, upgrade to recursive. Mixed writability within a directory prevents
// dir collapse — individual file entries are emitted instead.
func Compact(selectedPaths []string, writablePaths map[string]bool) ([]Entry, error) {
	if len(selectedPaths) == 0 {
		return []Entry{}, nil
	}

	selected := make(map[string]bool, len(selectedPaths))
	for _, p := range selectedPaths {
		selected[p] = true
	}

	// Group selected files by parent dir
	byDir := make(map[string][]string)
	for _, p := range selectedPaths {
		dir := filepath.Dir(p)
		byDir[dir] = append(byDir[dir], p)
	}

	// Collect unique dirs and sort shallowest first so parent dirs are
	// considered before their children, enabling top-down coverage marking.
	dirs := make([]string, 0, len(byDir))
	for d := range byDir {
		dirs = append(dirs, d)
	}
	sort.Slice(dirs, func(i, j int) bool {
		ci := strings.Count(dirs[i], string(filepath.Separator))
		cj := strings.Count(dirs[j], string(filepath.Separator))
		if ci != cj {
			return ci < cj // shallower first
		}
		return dirs[i] < dirs[j]
	})

	// Track which paths are already covered by a dir entry
	covered := make(map[string]bool)

	var entries []Entry

	for _, dir := range dirs {
		// Skip if this dir is already covered by a parent recursive entry
		alreadyCovered := false
		for covPath := range covered {
			if isUnder(dir, covPath) {
				alreadyCovered = true
				break
			}
		}
		if alreadyCovered {
			continue
		}

		allMd, err := mdFilesInDir(dir)
		if err != nil {
			// Can't read dir; fall back to individual files
			for _, p := range byDir[dir] {
				if !covered[p] {
					entries = append(entries, Entry{
						Type:     EntryTypeFile,
						Path:     p,
						Name:     DefaultName(p),
						Writable: writablePaths[p],
					})
				}
			}
			continue
		}

		subdirs, err := subdirsOf(dir)
		if err != nil {
			subdirs = nil
		}
		hasSubdirs := len(subdirs) > 0

		// Check if all selected files in this dir have consistent writable state.
		// Mixed writability prevents dir collapse.
		dirWritable := writablePaths[byDir[dir][0]]
		mixedWritable := false
		for _, p := range byDir[dir] {
			if writablePaths[p] != dirWritable {
				mixedWritable = true
				break
			}
		}

		// Collapse to dir entry only when there's genuine compaction benefit:
		// either multiple .md files or subdirectories involved.
		// Mixed writability also prevents collapse.
		if len(allMd) == 0 || !allSelected(allMd, selected) || (len(allMd) == 1 && !hasSubdirs) || mixedWritable {
			// Partial selection, single-file dir, or mixed writability: emit individual file entries
			for _, p := range byDir[dir] {
				if !covered[p] {
					entries = append(entries, Entry{
						Type:     EntryTypeFile,
						Path:     p,
						Name:     DefaultName(p),
						Writable: writablePaths[p],
					})
				}
			}
			continue
		}

		// All .md files in this dir are selected and there's real compaction value.
		// Check if we can go recursive (only meaningful if subdirs exist).
		fullyRecursive := false
		if hasSubdirs {
			fullyRecursive, err = allSubdirsFullyCovered(dir, selected)
			if err != nil {
				fullyRecursive = false
			}
		}

		// When recursive, verify all descendant selected files share the same writability.
		if fullyRecursive {
			for path := range selected {
				if isUnder(path, dir) && path != dir {
					if writablePaths[path] != dirWritable {
						mixedWritable = true
						break
					}
				}
			}
			if mixedWritable {
				for _, p := range byDir[dir] {
					if !covered[p] {
						entries = append(entries, Entry{
							Type:     EntryTypeFile,
							Path:     p,
							Name:     DefaultName(p),
							Writable: writablePaths[p],
						})
					}
				}
				continue
			}
		}

		entries = append(entries, Entry{
			Type:      EntryTypeDir,
			Path:      dir,
			Name:      DefaultName(dir),
			Recursive: fullyRecursive,
			Writable:  dirWritable,
		})

		if fullyRecursive {
			// Mark the entire subtree as covered.
			covered[dir] = true
		} else {
			// Flat dir entry: mark only the top-level .md files as covered.
			// Subdirs still need processing for their selected files.
			for _, p := range allMd {
				covered[p] = true
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	return entries, nil
}

// subdirsOf returns the immediate subdirectory paths of dir.
func subdirsOf(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, de := range dirEntries {
		if de.IsDir() {
			result = append(result, filepath.Join(dir, de.Name()))
		}
	}
	return result, nil
}

// mdFilesInDir returns all .md file paths directly inside dir (non-recursive).
func mdFilesInDir(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, de := range dirEntries {
		if !de.IsDir() && strings.HasSuffix(de.Name(), ".md") {
			result = append(result, filepath.Join(dir, de.Name()))
		}
	}
	return result, nil
}

// allSelected returns true if every path in paths is in selected.
func allSelected(paths []string, selected map[string]bool) bool {
	for _, p := range paths {
		if !selected[p] {
			return false
		}
	}
	return true
}

// allSubdirsFullyCovered checks recursively that every .md file at all depths
// under dir is present in selected.
func allSubdirsFullyCovered(dir string, selected map[string]bool) (bool, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, de := range dirEntries {
		if de.IsDir() {
			subdir := filepath.Join(dir, de.Name())
			// All .md files in this subdir must be selected
			mdFiles, err := mdFilesInDir(subdir)
			if err != nil {
				return false, err
			}
			if !allSelected(mdFiles, selected) {
				return false, nil
			}
			// And recurse deeper
			ok, err := allSubdirsFullyCovered(subdir, selected)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
	}
	return true, nil
}

// isUnder returns true if path is inside (or equal to) dir.
func isUnder(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

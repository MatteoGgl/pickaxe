package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/matteo/pickaxe/internal/config"
)

// ResolvedFile is a concrete file path with its display name and availability info.
type ResolvedFile struct {
	Name        string
	Path        string
	LastMod     time.Time
	Unavailable bool
}

// DefaultName derives the short name from a file or directory path.
// For files: strips extension. For directories: uses last non-empty path component.
func DefaultName(path string) string {
	clean := strings.TrimRight(path, "/")
	base := filepath.Base(clean)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func hasName(cfg *config.ProjectConfig, name string) bool {
	for _, e := range cfg.Entries {
		if e.Name == name {
			return true
		}
	}
	return false
}

// AddFile registers a single file in cfg. name may be empty (defaults to filename sans ext).
func AddFile(cfg *config.ProjectConfig, path, name string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("path %q not found: %w", path, err)
	}
	if name == "" {
		name = DefaultName(path)
	}
	if hasName(cfg, name) {
		return fmt.Errorf("name %q already in use; use --as to pick a different name", name)
	}
	cfg.Entries = append(cfg.Entries, config.Entry{
		Type: config.EntryTypeFile,
		Path: path,
		Name: name,
	})
	return nil
}

// AddDir registers a directory in cfg. name may be empty (defaults to dir base name).
func AddDir(cfg *config.ProjectConfig, path, name string, recursive bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path %q not found: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", path)
	}
	if name == "" {
		name = DefaultName(path)
	}
	if hasName(cfg, name) {
		return fmt.Errorf("name %q already in use; use --as to pick a different name", name)
	}
	cfg.Entries = append(cfg.Entries, config.Entry{
		Type:      config.EntryTypeDir,
		Path:      path,
		Name:      name,
		Recursive: recursive,
	})
	return nil
}

// Remove removes the entry with the given name from cfg.
func Remove(cfg *config.ProjectConfig, name string) error {
	for i, e := range cfg.Entries {
		if e.Name == name {
			cfg.Entries = append(cfg.Entries[:i], cfg.Entries[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no entry with name %q", name)
}

// EnumerateFiles expands an entry into concrete resolved files.
// Missing files are returned with Unavailable=true rather than as errors.
func EnumerateFiles(entry config.Entry) ([]ResolvedFile, error) {
	if entry.Type == config.EntryTypeFile {
		rf := ResolvedFile{Name: entry.Name, Path: entry.Path}
		info, err := os.Stat(entry.Path)
		if err != nil {
			rf.Unavailable = true
		} else {
			rf.LastMod = info.ModTime()
		}
		return []ResolvedFile{rf}, nil
	}

	var files []ResolvedFile
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if path != entry.Path && !entry.Recursive {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, _ := filepath.Rel(entry.Path, path)
		name := entry.Name + "/" + strings.TrimSuffix(rel, ".md")
		files = append(files, ResolvedFile{
			Name:    name,
			Path:    path,
			LastMod: info.ModTime(),
		})
		return nil
	}
	if err := filepath.Walk(entry.Path, walkFn); err != nil {
		return nil, err
	}
	return files, nil
}

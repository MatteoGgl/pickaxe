package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EntryType distinguishes file vs directory entries.
type EntryType string

const (
	EntryTypeFile EntryType = "file"
	EntryTypeDir  EntryType = "dir"
)

// Entry represents one registered file or directory.
type Entry struct {
	Type      EntryType `json:"type"`
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Recursive bool      `json:"recursive,omitempty"`
}

// ResolvedFile is a concrete file path with its display name and availability info.
type ResolvedFile struct {
	Name        string
	Path        string
	LastMod     time.Time
	Unavailable bool
}

// ConfigFilename is the name of the per-project config file.
const ConfigFilename = ".pickaxe.json"

var (
	ErrNotInitialized = errors.New("vault not initialized")
	ErrAlreadyExists  = errors.New(".pickaxe.json already exists")
	ErrNameCollision  = errors.New("name already in use")
	ErrNotFound       = errors.New("no entry with that name")
)

type projectConfig struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
}

// Vault owns a single .pickaxe.json and its mutations.
type Vault struct {
	path string
	cfg  *projectConfig
}

// Open reads an existing .pickaxe.json from dir. Returns ErrNotInitialized if absent.
func Open(dir string) (*Vault, error) {
	p := filepath.Join(dir, ConfigFilename)
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotInitialized
		}
		return nil, fmt.Errorf("read %s: %w", ConfigFilename, err)
	}
	var cfg projectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", ConfigFilename, err)
	}
	return &Vault{path: p, cfg: &cfg}, nil
}

// Init creates a new .pickaxe.json in dir. Returns ErrAlreadyExists if present.
func Init(dir string) (*Vault, error) {
	p := filepath.Join(dir, ConfigFilename)
	if _, err := os.Stat(p); err == nil {
		return nil, ErrAlreadyExists
	}
	v := &Vault{path: p, cfg: &projectConfig{Version: 1, Entries: []Entry{}}}
	if err := v.Save(); err != nil {
		return nil, err
	}
	return v, nil
}

// Entries returns a read-only copy of the registered entries.
func (v *Vault) Entries() []Entry {
	result := make([]Entry, len(v.cfg.Entries))
	copy(result, v.cfg.Entries)
	return result
}

// Save persists the current state to disk.
func (v *Vault) Save() error {
	if err := os.MkdirAll(filepath.Dir(v.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v.cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(v.path, data, 0o644)
}

func (v *Vault) hasName(name string) bool {
	for _, e := range v.cfg.Entries {
		if e.Name == name {
			return true
		}
	}
	return false
}

// AddFile registers a single file. name may be empty (defaults to filename sans ext).
func (v *Vault) AddFile(path, name string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("path %q not found: %w", path, err)
	}
	if name == "" {
		name = DefaultName(path)
	}
	if v.hasName(name) {
		return fmt.Errorf("%w: %q; use --as to pick a different name", ErrNameCollision, name)
	}
	v.cfg.Entries = append(v.cfg.Entries, Entry{
		Type: EntryTypeFile,
		Path: path,
		Name: name,
	})
	return nil
}

// AddDir registers a directory. name may be empty (defaults to dir base name).
func (v *Vault) AddDir(path, name string, recursive bool) error {
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
	if v.hasName(name) {
		return fmt.Errorf("%w: %q; use --as to pick a different name", ErrNameCollision, name)
	}
	v.cfg.Entries = append(v.cfg.Entries, Entry{
		Type:      EntryTypeDir,
		Path:      path,
		Name:      name,
		Recursive: recursive,
	})
	return nil
}

// Remove removes the entry with the given name.
func (v *Vault) Remove(name string) error {
	for i, e := range v.cfg.Entries {
		if e.Name == name {
			v.cfg.Entries = append(v.cfg.Entries[:i], v.cfg.Entries[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("%w: %q", ErrNotFound, name)
}

// DefaultName derives the short name from a file or directory path.
func DefaultName(path string) string {
	clean := strings.TrimRight(path, "/")
	base := filepath.Base(clean)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// enumerateFiles expands a single entry into concrete resolved files.
func enumerateFiles(entry Entry) ([]ResolvedFile, error) {
	if entry.Type == EntryTypeFile {
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

// ListFiles opens the vault at dir and returns all resolved files. Re-reads from disk every call.
func ListFiles(dir string) ([]ResolvedFile, error) {
	v, err := Open(dir)
	if err != nil {
		return nil, err
	}
	var all []ResolvedFile
	for _, entry := range v.cfg.Entries {
		files, err := enumerateFiles(entry)
		if err != nil {
			all = append(all, ResolvedFile{Name: entry.Name, Unavailable: true})
			continue
		}
		all = append(all, files...)
	}
	return all, nil
}

// ReadFile opens the vault at dir and returns the content of the named file.
// Uses an O(1) map lookup after a single enumeration pass.
func ReadFile(dir string, name string) (string, error) {
	v, err := Open(dir)
	if err != nil {
		return "", err
	}
	index := make(map[string]ResolvedFile)
	for _, entry := range v.cfg.Entries {
		files, err := enumerateFiles(entry)
		if err != nil {
			continue
		}
		for _, f := range files {
			index[f.Name] = f
		}
	}
	f, ok := index[name]
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrNotFound, name)
	}
	if f.Unavailable {
		return "", fmt.Errorf("file %q is currently unavailable", name)
	}
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return "", fmt.Errorf("could not read file %q", name)
	}
	return string(data), nil
}

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

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

// ProjectConfig is stored at <repo-root>/.pickaxe.json.
type ProjectConfig struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
}

// ProjectConfigFilename is the name of the per-project config file.
const ProjectConfigFilename = ".pickaxe.json"

// ReadProjectConfig reads and parses .pickaxe.json from path.
func ReadProjectConfig(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// WriteProjectConfig serializes cfg to path.
func WriteProjectConfig(path string, cfg *ProjectConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

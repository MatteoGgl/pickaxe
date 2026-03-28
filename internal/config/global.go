package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/matteoggl/pickaxe/internal/pathutil"
)

// GlobalConfig is stored at ~/.config/pickaxe/config.json.
type GlobalConfig struct {
	VaultRoot string `json:"vault_root"`
}

// DefaultGlobalConfigPath returns the default path for the global config.
func DefaultGlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "pickaxe", "config.json"), nil
}

// ReadGlobalConfig reads and parses the global config from path.
func ReadGlobalConfig(path string) (*GlobalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// WriteGlobalConfig serializes cfg to path atomically.
func WriteGlobalConfig(path string, cfg *GlobalConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return pathutil.AtomicWrite(path, data, 0o644)
}

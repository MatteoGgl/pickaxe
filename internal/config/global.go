package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// GlobalConfig is stored at ~/.config/pickaxe/config.json.
type GlobalConfig struct {
	VaultRoot string `json:"vault_root"`
}

// DefaultGlobalConfigPath returns the default path for the global config.
func DefaultGlobalConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pickaxe", "config.json")
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

// WriteGlobalConfig serializes cfg to path, creating parent directories as needed.
func WriteGlobalConfig(path string, cfg *GlobalConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

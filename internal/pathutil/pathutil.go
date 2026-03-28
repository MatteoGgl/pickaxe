package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWrite writes data to path using write-to-temp + rename.
// Cleans up any prior crash orphan at the temp path.
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	os.Remove(tmp)
	if err := os.WriteFile(tmp, data, perm); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// ExpandHome replaces a leading ~/ with the user's home directory.
func ExpandHome(raw string) (string, error) {
	if len(raw) >= 2 && raw[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home dir: %w", err)
		}
		return home + raw[1:], nil
	}
	return raw, nil
}

// ExpandAndResolve expands ~/ and resolves to an absolute path.
func ExpandAndResolve(raw string) (string, error) {
	expanded, err := ExpandHome(raw)
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	return abs, nil
}

package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
)

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

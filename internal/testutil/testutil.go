package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates a temp directory and registers cleanup.
func TempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "pickaxe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// WriteFile writes content to path, creating parent dirs as needed.
func WriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

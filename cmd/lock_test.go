package cmd_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matteoggl/pickaxe/cmd"
	"github.com/matteoggl/pickaxe/internal/testutil"
	"github.com/matteoggl/pickaxe/internal/vault"
)

func vaultWithFile(t *testing.T, writable bool) (*vault.Vault, string) {
	t.Helper()
	dir := testutil.TempDir(t)
	v, err := vault.Init(dir)
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(dir, "notes.md")
	testutil.WriteFile(t, filePath, "content")
	if err := v.AddFile(filePath, "notes", writable); err != nil {
		t.Fatal(err)
	}
	return v, dir
}

func TestLockCmd_Success(t *testing.T) {
	v, _ := vaultWithFile(t, true) // writable=true, lock it
	var buf bytes.Buffer
	if err := cmd.LockEntry(v, "notes", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Entries()[0].Writable {
		t.Error("expected Writable=false after lock")
	}
	if !strings.Contains(buf.String(), "read-only") {
		t.Errorf("expected 'read-only' in output, got %q", buf.String())
	}
}

func TestLockCmd_AlreadyLocked(t *testing.T) {
	v, _ := vaultWithFile(t, false) // already read-only
	var buf bytes.Buffer
	if err := cmd.LockEntry(v, "notes", &buf); err != nil {
		t.Fatalf("expected idempotent lock, got error: %v", err)
	}
	if v.Entries()[0].Writable {
		t.Error("expected Writable=false")
	}
}

func TestUnlockCmd_Success(t *testing.T) {
	v, _ := vaultWithFile(t, false) // read-only, unlock it
	var buf bytes.Buffer
	if err := cmd.UnlockEntry(v, "notes", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !v.Entries()[0].Writable {
		t.Error("expected Writable=true after unlock")
	}
	if !strings.Contains(buf.String(), "read-write") {
		t.Errorf("expected 'read-write' in output, got %q", buf.String())
	}
}

func TestUnlockCmd_AlreadyUnlocked(t *testing.T) {
	v, _ := vaultWithFile(t, true) // already writable
	var buf bytes.Buffer
	if err := cmd.UnlockEntry(v, "notes", &buf); err != nil {
		t.Fatalf("expected idempotent unlock, got error: %v", err)
	}
	if !v.Entries()[0].Writable {
		t.Error("expected Writable=true")
	}
}

func TestLockCmd_NotFound(t *testing.T) {
	v, _ := vaultWithFile(t, true)
	var buf bytes.Buffer
	err := cmd.LockEntry(v, "nonexistent", &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent entry, got nil")
	}
}

func TestUnlockCmd_NotFound(t *testing.T) {
	v, _ := vaultWithFile(t, false)
	var buf bytes.Buffer
	err := cmd.UnlockEntry(v, "nonexistent", &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent entry, got nil")
	}
}

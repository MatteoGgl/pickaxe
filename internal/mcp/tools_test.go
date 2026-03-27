package mcp_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matteo/pickaxe/internal/config"
	internalmcp "github.com/matteo/pickaxe/internal/mcp"
	"github.com/matteo/pickaxe/internal/testutil"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// writeRegistry writes a ProjectConfig as .pickaxe.json in dir and returns the path.
func writeRegistry(t *testing.T, dir string, cfg *config.ProjectConfig) string {
	t.Helper()
	path := filepath.Join(dir, config.ProjectConfigFilename)
	if err := config.WriteProjectConfig(path, cfg); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestListVaultFiles_Empty(t *testing.T) {
	dir := testutil.TempDir(t)
	path := writeRegistry(t, dir, &config.ProjectConfig{Version: 1, Entries: []config.Entry{}})
	handler := internalmcp.MakeListVaultFilesHandler(path)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected non-error result")
	}
	// Should return [] not null
	text := result.Content[0].(*sdkmcp.TextContent).Text
	var items []any
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		t.Fatalf("expected JSON array, got: %q (%v)", text, err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty array, got %d items", len(items))
	}
}

func TestListVaultFiles_WithFile(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "# ADRs")

	path := writeRegistry(t, dir, &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: filePath, Name: "adrs"},
		},
	})
	handler := internalmcp.MakeListVaultFilesHandler(path)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result")
	}
}

func TestListVaultFiles_UnavailableFile(t *testing.T) {
	dir := testutil.TempDir(t)
	path := writeRegistry(t, dir, &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: "/nonexistent/file.md", Name: "ghost"},
		},
	})
	handler := internalmcp.MakeListVaultFilesHandler(path)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected non-error result even for unavailable file")
	}
}

func TestListVaultFiles_NoRegistry(t *testing.T) {
	handler := internalmcp.MakeListVaultFilesHandler("/nonexistent/.pickaxe.json")

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("missing registry should return a message, not an error")
	}
}

func TestReadVaultFile_Success(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "# ADRs\nDecision 1")

	path := writeRegistry(t, dir, &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: filePath, Name: "adrs"},
		},
	})
	handler := internalmcp.MakeReadVaultFileHandler(path)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "adrs"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result")
	}
}

func TestReadVaultFile_NotRegistered(t *testing.T) {
	dir := testutil.TempDir(t)
	path := writeRegistry(t, dir, &config.ProjectConfig{Version: 1, Entries: []config.Entry{}})
	handler := internalmcp.MakeReadVaultFileHandler(path)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unregistered name")
	}
}

func TestReadVaultFile_UnavailableFile(t *testing.T) {
	dir := testutil.TempDir(t)
	path := writeRegistry(t, dir, &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: "/nonexistent/file.md", Name: "ghost"},
		},
	})
	handler := internalmcp.MakeReadVaultFileHandler(path)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "ghost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unavailable file")
	}
}

func TestReadVaultFile_NoRegistry(t *testing.T) {
	handler := internalmcp.MakeReadVaultFileHandler("/nonexistent/.pickaxe.json")

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "anything"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true with no registry")
	}
}

func TestListVaultFiles_NoPathInResponse(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, filePath, "hello")

	path := writeRegistry(t, dir, &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: filePath, Name: "note"},
		},
	})
	handler := internalmcp.MakeListVaultFilesHandler(path)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	var items []map[string]any
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		t.Fatalf("expected JSON array: %v", err)
	}
	for _, item := range items {
		if _, ok := item["path"]; ok {
			t.Errorf("path must not appear in list_vault_files response, got: %v", item)
		}
	}
}

func TestListVaultFiles_UnavailableNoPath(t *testing.T) {
	dir := testutil.TempDir(t)
	path := writeRegistry(t, dir, &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: "/nonexistent/secret/file.md", Name: "ghost"},
		},
	})
	handler := internalmcp.MakeListVaultFilesHandler(path)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	var items []map[string]any
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		t.Fatalf("expected JSON array: %v", err)
	}
	for _, item := range items {
		if _, ok := item["path"]; ok {
			t.Errorf("path must not appear for unavailable files, got: %v", item)
		}
	}
}

func TestReadVaultFile_UnavailableErrorNoPath(t *testing.T) {
	dir := testutil.TempDir(t)
	const secretPath = "/nonexistent/secret/private.md"
	path := writeRegistry(t, dir, &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: secretPath, Name: "secret"},
		},
	})
	handler := internalmcp.MakeReadVaultFileHandler(path)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if strings.Contains(text, secretPath) {
		t.Errorf("error message must not contain the file path, got: %q", text)
	}
}

func TestReadVaultFile_ReadErrorNoPath(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "locked.md")
	testutil.WriteFile(t, filePath, "secret content")
	if err := os.Chmod(filePath, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(filePath, 0o644) })

	path := writeRegistry(t, dir, &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: filePath, Name: "locked"},
		},
	})
	handler := internalmcp.MakeReadVaultFileHandler(path)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "locked"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unreadable file")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if strings.Contains(text, filePath) {
		t.Errorf("error message must not contain the file path, got: %q", text)
	}
}

func TestListVaultFiles_ReloadsAfterChange(t *testing.T) {
	dir := testutil.TempDir(t)
	path := writeRegistry(t, dir, &config.ProjectConfig{Version: 1, Entries: []config.Entry{}})
	handler := internalmcp.MakeListVaultFilesHandler(path)

	// First call: empty
	result, _, _ := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	text := result.Content[0].(*sdkmcp.TextContent).Text
	var items []any
	json.Unmarshal([]byte(text), &items)
	if len(items) != 0 {
		t.Fatalf("expected 0 items initially, got %d", len(items))
	}

	// Update registry on disk
	filePath := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, filePath, "hello")
	writeRegistry(t, dir, &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{{Type: config.EntryTypeFile, Path: filePath, Name: "note"}},
	})

	// Second call: should see the new entry
	result, _, _ = handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	text = result.Content[0].(*sdkmcp.TextContent).Text
	json.Unmarshal([]byte(text), &items)
	if len(items) != 1 {
		t.Fatalf("expected 1 item after registry update, got %d", len(items))
	}
}

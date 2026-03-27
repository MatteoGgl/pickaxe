package mcp_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/matteo/pickaxe/internal/config"
	internalmcp "github.com/matteo/pickaxe/internal/mcp"
	"github.com/matteo/pickaxe/internal/testutil"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestListVaultFiles_Empty(t *testing.T) {
	cfg := &config.ProjectConfig{Version: 1, Entries: []config.Entry{}}
	handler := internalmcp.MakeListVaultFilesHandler(cfg)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected non-error result")
	}
}

func TestListVaultFiles_WithFile(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "# ADRs")

	cfg := &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: filePath, Name: "adrs"},
		},
	}
	handler := internalmcp.MakeListVaultFilesHandler(cfg)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result")
	}
}

func TestListVaultFiles_UnavailableFile(t *testing.T) {
	cfg := &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: "/nonexistent/file.md", Name: "ghost"},
		},
	}
	handler := internalmcp.MakeListVaultFilesHandler(cfg)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected non-error result even for unavailable file")
	}
}

func TestReadVaultFile_Success(t *testing.T) {
	dir := testutil.TempDir(t)
	filePath := filepath.Join(dir, "adrs.md")
	testutil.WriteFile(t, filePath, "# ADRs\nDecision 1")

	cfg := &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: filePath, Name: "adrs"},
		},
	}
	handler := internalmcp.MakeReadVaultFileHandler(cfg)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "adrs"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result")
	}
}

func TestReadVaultFile_NotRegistered(t *testing.T) {
	cfg := &config.ProjectConfig{Version: 1, Entries: []config.Entry{}}
	handler := internalmcp.MakeReadVaultFileHandler(cfg)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unregistered name")
	}
}

func TestReadVaultFile_UnavailableFile(t *testing.T) {
	cfg := &config.ProjectConfig{
		Version: 1,
		Entries: []config.Entry{
			{Type: config.EntryTypeFile, Path: "/nonexistent/file.md", Name: "ghost"},
		},
	}
	handler := internalmcp.MakeReadVaultFileHandler(cfg)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "ghost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unavailable file")
	}
}

func TestReadVaultFile_NilConfig(t *testing.T) {
	handler := internalmcp.MakeReadVaultFileHandler(nil)

	result, _, err := handler(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "anything"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true with nil config")
	}
}

package mcp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	internalmcp "github.com/matteo/pickaxe/internal/mcp"
	"github.com/matteo/pickaxe/internal/vault"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type fakeVault struct {
	files    []vault.ResolvedFile
	listErr  error
	contents map[string]string
	readErr  error
}

func (f *fakeVault) ListFiles() ([]vault.ResolvedFile, error) {
	return f.files, f.listErr
}

func (f *fakeVault) ReadFile(name string) (string, error) {
	if f.readErr != nil {
		return "", f.readErr
	}
	c, ok := f.contents[name]
	if !ok {
		return "", vault.ErrNotFound
	}
	return c, nil
}

func listHandler(fv *fakeVault) func(context.Context, *sdkmcp.CallToolRequest, internalmcp.ListVaultFilesParams) (*sdkmcp.CallToolResult, any, error) {
	return internalmcp.MakeListVaultFilesHandler(fv)
}

func readHandler(fv *fakeVault) func(context.Context, *sdkmcp.CallToolRequest, internalmcp.ReadVaultFileParams) (*sdkmcp.CallToolResult, any, error) {
	return internalmcp.MakeReadVaultFileHandler(fv)
}

func TestListVaultFiles_Empty(t *testing.T) {
	fv := &fakeVault{}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected non-error result")
	}
	var items []any
	json.Unmarshal([]byte(result.Content[0].(*sdkmcp.TextContent).Text), &items)
	if len(items) != 0 {
		t.Errorf("expected empty array, got %d items", len(items))
	}
}

func TestListVaultFiles_WithFile(t *testing.T) {
	fv := &fakeVault{files: []vault.ResolvedFile{{Name: "adrs", LastMod: time.Now()}}}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("unexpected error result")
	}
	var items []map[string]any
	json.Unmarshal([]byte(result.Content[0].(*sdkmcp.TextContent).Text), &items)
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestListVaultFiles_UnavailableFile(t *testing.T) {
	fv := &fakeVault{files: []vault.ResolvedFile{{Name: "ghost", Unavailable: true}}}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected non-error result even for unavailable file")
	}
}

func TestListVaultFiles_NoRegistry(t *testing.T) {
	fv := &fakeVault{listErr: vault.ErrNotInitialized}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("missing registry should return a message, not an error")
	}
}

func TestReadVaultFile_Success(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"adrs": "# ADRs\nDecision 1"}}
	result, _, err := readHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "adrs"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result")
	}
	if result.Content[0].(*sdkmcp.TextContent).Text != "# ADRs\nDecision 1" {
		t.Errorf("unexpected content: %q", result.Content[0].(*sdkmcp.TextContent).Text)
	}
}

func TestReadVaultFile_NotRegistered(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{}}
	result, _, err := readHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unregistered name")
	}
}

func TestReadVaultFile_UnavailableFile(t *testing.T) {
	fv := &fakeVault{readErr: vault.ErrNotFound}
	result, _, err := readHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "ghost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unavailable file")
	}
}

func TestReadVaultFile_NoRegistry(t *testing.T) {
	fv := &fakeVault{readErr: vault.ErrNotInitialized}
	result, _, err := readHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "anything"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true with no registry")
	}
}

func TestListVaultFiles_NoPathInResponse(t *testing.T) {
	fv := &fakeVault{files: []vault.ResolvedFile{{Name: "note", LastMod: time.Now()}}}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var items []map[string]any
	json.Unmarshal([]byte(result.Content[0].(*sdkmcp.TextContent).Text), &items)
	for _, item := range items {
		if _, ok := item["path"]; ok {
			t.Errorf("path must not appear in list_vault_files response, got: %v", item)
		}
	}
}

func TestListVaultFiles_UnavailableNoPath(t *testing.T) {
	fv := &fakeVault{files: []vault.ResolvedFile{{Name: "ghost", Unavailable: true}}}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var items []map[string]any
	json.Unmarshal([]byte(result.Content[0].(*sdkmcp.TextContent).Text), &items)
	for _, item := range items {
		if _, ok := item["path"]; ok {
			t.Errorf("path must not appear for unavailable files, got: %v", item)
		}
	}
}

func TestReadVaultFile_UnavailableErrorNoPath(t *testing.T) {
	// vault.ReadFile returns "file %q is currently unavailable" (name, not path)
	fv := &fakeVault{readErr: fmt.Errorf("file %q is currently unavailable", "secret")}
	result, _, err := readHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	if strings.Contains(result.Content[0].(*sdkmcp.TextContent).Text, "/nonexistent/secret") {
		t.Errorf("error message must not contain file path")
	}
}

func TestReadVaultFile_ReadErrorNoPath(t *testing.T) {
	// vault.ReadFile returns "could not read file %q" (name, not path)
	fv := &fakeVault{readErr: fmt.Errorf("could not read file %q", "locked")}
	result, _, err := readHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "locked"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unreadable file")
	}
	if strings.Contains(result.Content[0].(*sdkmcp.TextContent).Text, "/tmp/") {
		t.Errorf("error message must not contain file path")
	}
}

func TestReadVaultFile_StripsFrontmatter(t *testing.T) {
	// vault.ReadFile strips frontmatter before returning; handler passes through as-is
	fv := &fakeVault{contents: map[string]string{"note": "\n# Hello"}}
	result, _, err := readHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "note"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "# Hello") {
		t.Errorf("body missing: got %q", text)
	}
}

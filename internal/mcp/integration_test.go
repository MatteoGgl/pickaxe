package mcp_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/matteoggl/pickaxe/internal/testutil"
)

// buildPickaxe compiles the binary to a temp file and returns its path.
func buildPickaxe(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "pickaxe")
	root, err := filepath.Abs(filepath.Join("..", "..", "."))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func connectToServer(t *testing.T, bin, dir string) (context.CancelFunc, *sdkmcp.ClientSession) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.Command(bin, "serve")
	cmd.Dir = dir

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	session, err := client.Connect(ctx, &sdkmcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		cancel()
		t.Fatalf("connect to server: %v", err)
	}
	t.Cleanup(func() {
		session.Close()
		cancel()
	})
	return cancel, session
}

func TestIntegration_ListVaultFiles_NoConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	bin := buildPickaxe(t)
	dir := testutil.TempDir(t) // no .pickaxe.json here

	_, session := connectToServer(t, bin, dir)

	result, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "list_vault_files",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("list_vault_files: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %v", result.Content)
	}
	// Should mention no registry found
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "pickaxe init") {
		t.Errorf("expected hint to run pickaxe init, got: %q", text)
	}
}

func TestIntegration_ReadVaultFile_Found(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	bin := buildPickaxe(t)
	dir := testutil.TempDir(t)

	vaultFile := filepath.Join(dir, "note.md")
	testutil.WriteFile(t, vaultFile, "# My Note\nContent here.")
	registry := fmt.Sprintf(`{"version":1,"entries":[{"type":"file","path":%q,"name":"note"}]}`, vaultFile)
	testutil.WriteFile(t, filepath.Join(dir, ".pickaxe.json"), registry)

	_, session := connectToServer(t, bin, dir)

	result, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "read_vault_file",
		Arguments: map[string]any{"name": "note"},
	})
	if err != nil {
		t.Fatalf("read_vault_file: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %v", result.Content)
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "My Note") {
		t.Errorf("expected file content, got: %q", text)
	}
}

func TestIntegration_ReadVaultFile_Missing(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	bin := buildPickaxe(t)
	dir := testutil.TempDir(t)

	// Register a file that doesn't exist
	registry := `{"version":1,"entries":[{"type":"file","path":"/nonexistent/ghost.md","name":"ghost"}]}`
	testutil.WriteFile(t, filepath.Join(dir, ".pickaxe.json"), registry)

	_, session := connectToServer(t, bin, dir)

	result, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "read_vault_file",
		Arguments: map[string]any{"name": "ghost"},
	})
	if err != nil {
		t.Fatalf("read_vault_file: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unavailable file")
	}
}

func TestIntegration_ListVaultFiles_MarksUnavailable(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	bin := buildPickaxe(t)
	dir := testutil.TempDir(t)

	registry := `{"version":1,"entries":[{"type":"file","path":"/nonexistent/ghost.md","name":"ghost"}]}`
	testutil.WriteFile(t, filepath.Join(dir, ".pickaxe.json"), registry)

	_, session := connectToServer(t, bin, dir)

	result, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "list_vault_files",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("list_vault_files: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result (unavailable should be in list, not an error)")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "unavailable") {
		t.Errorf("expected unavailable flag in response, got: %q", text)
	}
}

// Ensure os is used (for potential future use in this test file).
var _ = os.DevNull

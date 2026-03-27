package mcp

import (
	"github.com/matteoggl/pickaxe/internal/vault"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewServer creates a configured MCP server. registryPath is the path to
// .pickaxe.json; it is re-read on every tool call so changes from `pickaxe add`
// are visible without restarting the server.
func NewServer(registryPath string) *sdkmcp.Server {
	s := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "pickaxe", Version: "1.0.0"}, nil)
	vr := vault.NewReader(registryPath)

	sdkmcp.AddTool(s,
		&sdkmcp.Tool{
			Name:        "list_vault_files",
			Description: "List all vault files registered for this project. Returns name, last_modified, and unavailable status for each file.",
		},
		MakeListVaultFilesHandler(vr),
	)

	sdkmcp.AddTool(s,
		&sdkmcp.Tool{
			Name:        "read_vault_file",
			Description: "Read the full content of a registered vault file by name. Use list_vault_files first to see available names.",
		},
		MakeReadVaultFileHandler(vr),
	)

	return s
}

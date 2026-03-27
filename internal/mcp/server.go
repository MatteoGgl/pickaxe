package mcp

import (
	"github.com/matteo/pickaxe/internal/config"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewServer creates a configured MCP server. cfg may be nil.
func NewServer(cfg *config.ProjectConfig) *sdkmcp.Server {
	s := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "pickaxe", Version: "1.0.0"}, nil)

	sdkmcp.AddTool(s,
		&sdkmcp.Tool{
			Name:        "list_vault_files",
			Description: "List all vault files registered for this project. Returns name, path, last_modified, and unavailable status for each file.",
		},
		MakeListVaultFilesHandler(cfg),
	)

	sdkmcp.AddTool(s,
		&sdkmcp.Tool{
			Name:        "read_vault_file",
			Description: "Read the full content of a registered vault file by name. Use list_vault_files first to see available names.",
		},
		MakeReadVaultFileHandler(cfg),
	)

	return s
}

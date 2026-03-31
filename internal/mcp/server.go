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
	tracker := NewReadTracker()

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
			Description: "Read the content of a registered vault file by name. Returns numbered lines (cat -n format). Optional: 'offset' (1-based start line, default 1) and 'limit' (number of lines, default all). Use list_vault_files first to see available names.",
		},
		MakeReadVaultFileHandler(vr, tracker),
	)

	sdkmcp.AddTool(s,
		&sdkmcp.Tool{
			Name:        "update_vault_file",
			Description: "Update a registered vault file. Requires read_vault_file first. Two modes: (1) full replace with 'content', or (2) targeted edit with 'old_string' and 'new_string' (must match exactly once). Only works on writable files — use list_vault_files to check.",
		},
		MakeUpdateVaultFileHandler(vr, tracker),
	)

	return s
}

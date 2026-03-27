# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...
go build -o dist/pickaxe .

# Install to GOPATH/bin
go install .

# Test (short mode skips integration tests)
go test ./... -short
go test ./... -timeout 60s          # includes integration tests
go test ./internal/mcp/... -v -run TestIntegration -timeout 60s

# Single test
go test ./internal/registry/... -v -run TestEnumerateFiles_Flat

# Run
./dist/pickaxe --help
```

## Architecture

Single Go binary, two modes:

**CLI mode** (`cmd/`) — cobra commands that manage `.pickaxe.json` in the project root:
- `config set vault <path>` — writes `~/.config/pickaxe/config.json`
- `init` — creates `.pickaxe.json` with empty entries
- `add [path]` — no-arg opens TUI picker; with path registers file/dir directly
- `remove <name>`, `list` — mutate/read `.pickaxe.json`
- `serve` — starts the MCP server

**MCP server mode** (`internal/mcp/`) — stdio JSON-RPC server started by Claude Code. Exposes two tools: `list_vault_files` and `read_vault_file`. The registry is **re-read from disk on every tool call** so `pickaxe add` changes are visible immediately without restarting the server.

## Key design decisions

**Two config files:**
- Global: `~/.config/pickaxe/config.json` — stores `vault_root` for the TUI picker
- Per-project: `.pickaxe.json` — stores registered entries (version + entries array)

**Registry entries** are either `type: "file"` or `type: "dir"`. Directory entries are expanded dynamically by `registry.EnumerateFiles` at call time (flat or recursive, `.md` only). Names for dir children are prefixed: `dir-name/file-name`.

**MCP SDK:** Uses `github.com/modelcontextprotocol/go-sdk` (the official SDK), not `mark3labs/mcp-go`. Tool handlers use typed param structs (`ListVaultFilesParams`, `ReadVaultFileParams`). Integration tests use `sdkmcp.CommandTransport` to spin up the binary as a subprocess.

**TUI picker** (`internal/tui/picker.go`) uses bubbletea with `tea.WithAltScreen()` and viewport scrolling. Key bindings: space=toggle, enter=navigate into dir, ctrl+d=confirm, /=filter, q=cancel.

## Package layout

| Package | Responsibility |
|---------|---------------|
| `internal/config` | Read/write global config and `.pickaxe.json` |
| `internal/registry` | Business logic: add/remove/enumerate, name collision detection |
| `internal/mcp` | MCP server, tool handlers, integration tests |
| `internal/tui` | Bubbletea multi-select file picker |
| `internal/testutil` | Shared test helpers (temp dirs, file writes) |
| `cmd/` | Cobra CLI commands |

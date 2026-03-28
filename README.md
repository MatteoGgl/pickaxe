# pickaxe

Expose Obsidian vault files to Claude Code via MCP.

## Install

```sh
brew install matteoggl/pickaxe/pickaxe
```

Or with Go:

```sh
go install github.com/matteoggl/pickaxe@latest
```

Or build from source:

```sh
go build -o dist/pickaxe .
```

## Setup

```sh
pickaxe config set vault ~/your-vault
pickaxe init
```

`config set vault` writes the vault root to `~/.config/pickaxe/config.json` (used by the TUI picker). `init` creates `.pickaxe.json` in the current project directory.

## Adding files

```sh
pickaxe pick                 # open TUI picker
pickaxe add path/to/file.md  # register directly
pickaxe add path/to/dir      # register a directory
```

TUI key bindings: `space`=toggle, `enter`=navigate into dir, `ctrl+d`=confirm, `/`=filter, `q`=cancel.

Directories show a tri-state checkbox: `[x]` all files selected, `[~]` some files selected, `[ ]` none. Toggling a dir selects or deselects all `.md` files inside recursively. A partial dir (`[~]`) toggles to fully selected.

Flags: `--as <name>` to set a custom name, `--recursive` to include nested directories.

Directory entries expand to all `.md` files inside. Child names are prefixed as `dir-name/file-name`.

## MCP server

Add to `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "pickaxe": {
      "command": "pickaxe",
      "args": ["serve"]
    }
  }
}
```

Changes to `.pickaxe.json` are picked up immediately — no server restart needed.

## Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `config set vault <path>` | `c` | Set Obsidian vault root |
| `init` | `i` | Create `.pickaxe.json` in current dir |
| `add <path>` | `a` | Register a file or directory directly |
| `remove <name\|hash>` | `rm`, `r` | Unregister an entry by name or hash prefix |
| `list` | `ls`, `l` | Show registered entries with hash IDs |
| `serve` | `s` | Start MCP server (stdio) |

## Hash IDs

Each entry gets a deterministic 8-character hex ID derived from its name. `list` displays them with the shortest unique prefix highlighted:

```
  3a7f[c291] [file] my-notes ~/vault/my-notes.md ✓
```

`remove` accepts a hash prefix instead of a full name — any unambiguous prefix works:

```sh
pickaxe remove 3a7f
```

## Privacy

YAML frontmatter is stripped from `.md` files before sending content to Claude. Tags, properties, and other metadata stay in your vault. Only the note body is shared.

To opt out, add `"strip_frontmatter": false` to `.pickaxe.json`.

## How it works

`.pickaxe.json` stores registered entries (files and dirs). `pickaxe serve` runs an MCP stdio server exposing `list_vault_files` and `read_vault_file`. The registry is re-read on every tool call, so `pickaxe add` changes are visible immediately without restarting the server.

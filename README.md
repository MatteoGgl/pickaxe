# pickaxe

Expose Obsidian vault files to Claude Code via MCP.

## Install

You can download the compiled binaries from the [Releases](https://github.com/MatteoGgl/pickaxe/releases) page or install it directly with Go:

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
pickaxe pick                            # open TUI picker
pickaxe add path/to/file.md             # register directly (read-only)
pickaxe add path/to/file.md --writable  # register as read-write
pickaxe add path/to/dir                 # register a directory
```

TUI key bindings: `space`=toggle, `w`=mark writable, `enter`=navigate into dir, `ctrl+d`=confirm, `/`=filter, `q`=cancel. Pressing `w` on an unchecked item selects it and marks it writable in one action.

Directories show a tri-state checkbox: `[x]` all files selected, `[~]` some files selected, `[ ]` none. Toggling a dir selects or deselects all `.md` files inside recursively. A partial dir (`[~]`) toggles to fully selected.

Flags: `--as <name>` to set a custom name, `--recursive` to include nested directories.

Directory entries expand to all `.md` files inside. Child names are prefixed as `dir-name/file-name`.

## Permissions

Files are read-only by default. Mark entries writable to allow Claude to update them:

```sh
pickaxe add path/to/file.md --writable
pickaxe unlock notes          # by name or hash prefix
pickaxe lock notes            # revert to read-only
pickaxe list                  # shows rw/r- per entry
```

In the TUI picker, press `w` on any item to toggle writable. The permission column shows `rw` (writable) or `r-` (read-only).

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
| `pick` | `p` | Interactive TUI picker for vault entries |
| `add <path>` | `a` | Register a file or directory directly |
| `remove <name\|hash>` | `rm`, `r` | Unregister an entry by name or hash prefix |
| `list` | `ls`, `l` | Show registered entries with hash IDs |
| `lock <name\|hash>` | `lk` | Mark entry read-only |
| `unlock <name\|hash>` | `ul` | Mark entry writable |
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

`.pickaxe.json` stores registered entries (files and dirs). `pickaxe serve` runs an MCP stdio server exposing `list_vault_files`, `read_vault_file`, and `update_vault_file`. Files must be read before they can be updated — `update_vault_file` returns an error if `read_vault_file` has not been called first for that file. The registry is re-read on every tool call, so `pickaxe add` changes are visible immediately without restarting the server.

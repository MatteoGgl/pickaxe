package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/matteoggl/pickaxe/internal/vault"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ReadTracker records which vault files have been read in the current session.
// update_vault_file requires a prior read_vault_file call for the same file.
type ReadTracker struct {
	mu   sync.Mutex
	read map[string]bool
}

func NewReadTracker() *ReadTracker {
	return &ReadTracker{read: make(map[string]bool)}
}

func (t *ReadTracker) MarkRead(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.read[name] = true
}

func (t *ReadTracker) HasRead(name string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.read[name]
}

func (t *ReadTracker) Invalidate(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.read, name)
}

// VaultReader abstracts vault access for testability.
type VaultReader interface {
	ListFiles() ([]vault.ResolvedFile, error)
	ReadFile(name string) (string, error)
}

type vaultFileInfo struct {
	Name        string `json:"name"`
	LastMod     string `json:"last_modified,omitempty"`
	Unavailable bool   `json:"unavailable,omitempty"`
	Writable    bool   `json:"writable"`
}

// ListVaultFilesParams is the (empty) parameter struct for list_vault_files.
type ListVaultFilesParams struct{}

// ReadVaultFileParams is the parameter struct for read_vault_file.
type ReadVaultFileParams struct {
	Name   string `json:"name"`
	Offset int    `json:"offset,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// MakeListVaultFilesHandler returns the handler for the list_vault_files tool.
func MakeListVaultFilesHandler(vr VaultReader) func(context.Context, *sdkmcp.CallToolRequest, ListVaultFilesParams) (*sdkmcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, _ ListVaultFilesParams) (*sdkmcp.CallToolResult, any, error) {
		files, err := vr.ListFiles()
		if errors.Is(err, vault.ErrNotInitialized) {
			msg := "no .pickaxe.json found in this project; run 'pickaxe init' to set up a registry"
			return &sdkmcp.CallToolResult{
				Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: msg}},
			}, nil, nil
		}
		if err != nil {
			return nil, nil, fmt.Errorf("open vault: %w", err)
		}

		infos := make([]vaultFileInfo, 0, len(files))
		for _, f := range files {
			info := vaultFileInfo{
				Name:        f.Name,
				Unavailable: f.Unavailable,
				Writable:    f.Writable,
			}
			if !f.Unavailable {
				info.LastMod = f.LastMod.Format("2006-01-02T15:04:05Z07:00")
			}
			infos = append(infos, info)
		}

		data, err := json.MarshalIndent(infos, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("marshal: %w", err)
		}
		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

// VaultWriter abstracts vault write access for testability.
type VaultWriter interface {
	WriteFile(name string, content string) error
	ReplaceInFile(name, oldStr, newStr string) error
}

// UpdateVaultFileParams is the parameter struct for update_vault_file.
type UpdateVaultFileParams struct {
	Name      string  `json:"name"`
	Content   string  `json:"content,omitempty"`
	OldString *string `json:"old_string,omitempty"`
	NewString *string `json:"new_string,omitempty"`
}

// FormatLines formats content with line numbers (cat -n style).
// offset is 1-based (default 1). limit=0 means all remaining lines.
func FormatLines(content string, offset, limit int) string {
	lines := strings.Split(content, "\n")
	if offset <= 0 {
		offset = 1
	}
	if offset > len(lines) {
		return ""
	}
	remaining := lines[offset-1:]
	if limit > 0 && limit < len(remaining) {
		remaining = remaining[:limit]
	}
	out := make([]string, len(remaining))
	for i, line := range remaining {
		out[i] = fmt.Sprintf("%6d\t%s", offset+i, line)
	}
	return strings.Join(out, "\n")
}

func errResult(msg string) (*sdkmcp.CallToolResult, any, error) {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: msg}},
		IsError: true,
	}, nil, nil
}

// MakeReadVaultFileHandler returns the handler for the read_vault_file tool.
func MakeReadVaultFileHandler(vr VaultReader, tracker *ReadTracker) func(context.Context, *sdkmcp.CallToolRequest, ReadVaultFileParams) (*sdkmcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, args ReadVaultFileParams) (*sdkmcp.CallToolResult, any, error) {
		if args.Name == "" {
			return errResult("name parameter is required")
		}

		content, err := vr.ReadFile(args.Name)
		if errors.Is(err, vault.ErrNotInitialized) {
			return errResult("no .pickaxe.json found in this project")
		}
		if errors.Is(err, vault.ErrNotFound) {
			return errResult(fmt.Sprintf("no registered file with name %q; use list_vault_files to see available names", args.Name))
		}
		if err != nil {
			return errResult(err.Error())
		}
		content = FormatLines(content, args.Offset, args.Limit)
		tracker.MarkRead(args.Name)
		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: content}},
		}, nil, nil
	}
}

// MakeUpdateVaultFileHandler returns the handler for the update_vault_file tool.
func MakeUpdateVaultFileHandler(vw VaultWriter, tracker *ReadTracker) func(context.Context, *sdkmcp.CallToolRequest, UpdateVaultFileParams) (*sdkmcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, args UpdateVaultFileParams) (*sdkmcp.CallToolResult, any, error) {
		if args.Name == "" {
			return errResult("name parameter is required")
		}

		hasContent := args.Content != ""
		hasReplace := args.OldString != nil || args.NewString != nil

		if hasContent && hasReplace {
			return errResult("cannot provide both 'content' and 'old_string'/'new_string'")
		}
		if !hasContent && !hasReplace {
			return errResult("provide 'content' for full replace or 'old_string'+'new_string' for targeted edit")
		}
		if hasReplace && (args.OldString == nil || args.NewString == nil) {
			return errResult("both old_string and new_string are required")
		}
		if args.OldString != nil && *args.OldString == "" {
			return errResult("old_string cannot be empty")
		}

		if !tracker.HasRead(args.Name) {
			return errResult("You must read the file before updating it. Use read_vault_file first.")
		}

		var err error
		if hasContent {
			err = vw.WriteFile(args.Name, args.Content)
		} else {
			err = vw.ReplaceInFile(args.Name, *args.OldString, *args.NewString)
		}
		if errors.Is(err, vault.ErrReadOnly) {
			return errResult(fmt.Sprintf("This file is read-only. Ask the user to make it writable with 'pickaxe unlock %s'.", args.Name))
		}
		if errors.Is(err, vault.ErrNotFound) {
			return errResult(fmt.Sprintf("no registered file with name %q; use list_vault_files to see available names", args.Name))
		}
		if errors.Is(err, vault.ErrNoMatch) {
			return errResult("old_string not found in file")
		}
		if errors.Is(err, vault.ErrAmbiguousMatch) {
			return errResult("old_string matches multiple locations; use a more specific string")
		}
		if err != nil {
			return errResult(err.Error())
		}
		tracker.Invalidate(args.Name)
		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: fmt.Sprintf("Successfully updated %q", args.Name)}},
		}, nil, nil
	}
}

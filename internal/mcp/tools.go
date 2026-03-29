package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	Name string `json:"name"`
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
}

// UpdateVaultFileParams is the parameter struct for update_vault_file.
type UpdateVaultFileParams struct {
	Name    string `json:"name"`
	Content string `json:"content"`
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

		if !tracker.HasRead(args.Name) {
			return errResult("You must read the file before updating it. Use read_vault_file first.")
		}

		err := vw.WriteFile(args.Name, args.Content)
		if errors.Is(err, vault.ErrReadOnly) {
			return errResult(fmt.Sprintf("This file is read-only. Ask the user to make it writable with 'pickaxe unlock %s'.", args.Name))
		}
		if errors.Is(err, vault.ErrNotFound) {
			return errResult(fmt.Sprintf("no registered file with name %q; use list_vault_files to see available names", args.Name))
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

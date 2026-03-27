package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/matteo/pickaxe/internal/vault"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type vaultFileInfo struct {
	Name        string `json:"name"`
	LastMod     string `json:"last_modified,omitempty"`
	Unavailable bool   `json:"unavailable,omitempty"`
}

// ListVaultFilesParams is the (empty) parameter struct for list_vault_files.
type ListVaultFilesParams struct{}

// ReadVaultFileParams is the parameter struct for read_vault_file.
type ReadVaultFileParams struct {
	Name string `json:"name"`
}

// MakeListVaultFilesHandler returns the handler for the list_vault_files tool.
func MakeListVaultFilesHandler(registryPath string) func(context.Context, *sdkmcp.CallToolRequest, ListVaultFilesParams) (*sdkmcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, _ ListVaultFilesParams) (*sdkmcp.CallToolResult, any, error) {
		files, err := vault.ListFiles(registryPath)
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

// MakeReadVaultFileHandler returns the handler for the read_vault_file tool.
func MakeReadVaultFileHandler(registryPath string) func(context.Context, *sdkmcp.CallToolRequest, ReadVaultFileParams) (*sdkmcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, args ReadVaultFileParams) (*sdkmcp.CallToolResult, any, error) {
		errResult := func(msg string) (*sdkmcp.CallToolResult, any, error) {
			return &sdkmcp.CallToolResult{
				Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: msg}},
				IsError: true,
			}, nil, nil
		}

		if args.Name == "" {
			return errResult("name parameter is required")
		}

		content, err := vault.ReadFile(registryPath, args.Name)
		if errors.Is(err, vault.ErrNotInitialized) {
			return errResult("no .pickaxe.json found in this project")
		}
		if errors.Is(err, vault.ErrNotFound) {
			return errResult(fmt.Sprintf("no registered file with name %q; use list_vault_files to see available names", args.Name))
		}
		if err != nil {
			return errResult(err.Error())
		}
		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: content}},
		}, nil, nil
	}
}

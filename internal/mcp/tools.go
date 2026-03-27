package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/matteo/pickaxe/internal/config"
	"github.com/matteo/pickaxe/internal/registry"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type vaultFileInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	LastMod     string `json:"last_modified,omitempty"`
	Unavailable bool   `json:"unavailable,omitempty"`
}

// ListVaultFilesParams is the (empty) parameter struct for list_vault_files.
type ListVaultFilesParams struct{}

// ReadVaultFileParams is the parameter struct for read_vault_file.
type ReadVaultFileParams struct {
	Name string `json:"name"`
}

// loadRegistry reads .pickaxe.json from registryPath on each call so changes
// made by `pickaxe add` are visible without restarting the server.
// Returns nil if the file doesn't exist.
func loadRegistry(registryPath string) *config.ProjectConfig {
	cfg, err := config.ReadProjectConfig(registryPath)
	if err != nil {
		return nil
	}
	return cfg
}

// MakeListVaultFilesHandler returns the handler for the list_vault_files tool.
func MakeListVaultFilesHandler(registryPath string) func(context.Context, *sdkmcp.CallToolRequest, ListVaultFilesParams) (*sdkmcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, _ ListVaultFilesParams) (*sdkmcp.CallToolResult, any, error) {
		cfg := loadRegistry(registryPath)
		if cfg == nil {
			msg := "no .pickaxe.json found in this project; run 'pickaxe init' to set up a registry"
			return &sdkmcp.CallToolResult{
				Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: msg}},
			}, nil, nil
		}

		infos := make([]vaultFileInfo, 0, len(cfg.Entries))
		for _, entry := range cfg.Entries {
			files, err := registry.EnumerateFiles(entry)
			if err != nil {
				infos = append(infos, vaultFileInfo{
					Name:        entry.Name,
					Path:        entry.Path,
					Unavailable: true,
				})
				continue
			}
			for _, f := range files {
				info := vaultFileInfo{
					Name:        f.Name,
					Path:        f.Path,
					Unavailable: f.Unavailable,
				}
				if !f.Unavailable {
					info.LastMod = f.LastMod.Format("2006-01-02T15:04:05Z07:00")
				}
				infos = append(infos, info)
			}
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

		cfg := loadRegistry(registryPath)
		if cfg == nil {
			return errResult("no .pickaxe.json found in this project")
		}
		if args.Name == "" {
			return errResult("name parameter is required")
		}

		for _, entry := range cfg.Entries {
			files, err := registry.EnumerateFiles(entry)
			if err != nil {
				continue
			}
			for _, f := range files {
				if f.Name != args.Name {
					continue
				}
				if f.Unavailable {
					return errResult(fmt.Sprintf("file %q is currently unavailable (path: %s)", args.Name, f.Path))
				}
				data, err := os.ReadFile(f.Path)
				if err != nil {
					return errResult(fmt.Sprintf("read error: %v", err))
				}
				return &sdkmcp.CallToolResult{
					Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: string(data)}},
				}, nil, nil
			}
		}

		return errResult(fmt.Sprintf("no registered file with name %q; use list_vault_files to see available names", args.Name))
	}
}

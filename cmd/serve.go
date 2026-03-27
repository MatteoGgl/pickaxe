package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/matteo/pickaxe/internal/config"
	internalmcp "github.com/matteo/pickaxe/internal/mcp"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server (used by Claude Code)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		registryPath := filepath.Join(cwd, config.ProjectConfigFilename)

		var cfg *config.ProjectConfig
		if loaded, err := config.ReadProjectConfig(registryPath); err == nil {
			cfg = loaded
		} else {
			fmt.Fprintf(os.Stderr, "pickaxe: no .pickaxe.json found, serving empty registry\n")
		}

		s := internalmcp.NewServer(cfg)
		if err := s.Run(context.Background(), &sdkmcp.StdioTransport{}); err != nil && err != io.EOF {
			// Non-EOF errors are unexpected; EOF means client disconnected normally.
			fmt.Fprintf(os.Stderr, "pickaxe serve: %v\n", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

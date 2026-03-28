package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	internalmcp "github.com/matteoggl/pickaxe/internal/mcp"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"s"},
	Short:   "Start the MCP server (used by Claude Code)",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		s := internalmcp.NewServer(cwd)
		if err := s.Run(context.Background(), &sdkmcp.StdioTransport{}); err != nil && err != io.EOF {
			return fmt.Errorf("pickaxe serve: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

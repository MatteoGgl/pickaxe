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
	Use:   "serve",
	Short: "Start the MCP server (used by Claude Code)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		s := internalmcp.NewServer(cwd)
		if err := s.Run(context.Background(), &sdkmcp.StdioTransport{}); err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "pickaxe serve: %v\n", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

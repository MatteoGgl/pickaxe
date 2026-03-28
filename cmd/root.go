package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pickaxe",
	Short: "Expose Obsidian vault files to Claude Code via MCP",
	Long: `   ░▒▒▓▓██████████████████▓▓▒▒░
  ░▒▓█▀                    ░▀█▓░
  ▒▓█     P I C K A X E  ⛏   █▓▒
  ░▒▓█▄░  ·  .,░ ▒░ ▓▒░▓█▒▄█▓▒░
   ░▒▓▓██████████████████▓▓▒▒░

Expose Obsidian vault files to Claude Code via MCP.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

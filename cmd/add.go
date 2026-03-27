package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/matteo/pickaxe/internal/config"
	"github.com/matteo/pickaxe/internal/registry"
	"github.com/spf13/cobra"
)

var addAlias string
var addRecursive bool

var addCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Register a vault file or directory",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// TUI mode — implemented in Task 13
			return fmt.Errorf("interactive picker not yet implemented; provide a path")
		}

		rawPath := args[0]

		if len(rawPath) > 1 && rawPath[:2] == "~/" {
			home, _ := os.UserHomeDir()
			rawPath = home + rawPath[1:]
		}

		absPath, err := filepath.Abs(rawPath)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		registryPath := filepath.Join(cwd, config.ProjectConfigFilename)

		cfg, err := config.ReadProjectConfig(registryPath)
		if err != nil {
			return fmt.Errorf("no .pickaxe.json found; run 'pickaxe init' first")
		}

		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("%q not found", absPath)
		}

		if info.IsDir() || strings.HasSuffix(rawPath, "/") {
			if err := registry.AddDir(cfg, absPath, addAlias, addRecursive); err != nil {
				return err
			}
		} else {
			if err := registry.AddFile(cfg, absPath, addAlias); err != nil {
				return err
			}
		}

		if err := config.WriteProjectConfig(registryPath, cfg); err != nil {
			return fmt.Errorf("write .pickaxe.json: %w", err)
		}

		fmt.Printf("registered: %s\n", absPath)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addAlias, "as", "", "Override the short name for this entry")
	addCmd.Flags().BoolVar(&addRecursive, "recursive", false, "Register directory recursively")
	rootCmd.AddCommand(addCmd)
}

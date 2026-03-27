package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/matteo/pickaxe/internal/config"
	"github.com/matteo/pickaxe/internal/registry"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a registered vault entry by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		registryPath := filepath.Join(cwd, config.ProjectConfigFilename)

		cfg, err := config.ReadProjectConfig(registryPath)
		if err != nil {
			return fmt.Errorf("no .pickaxe.json found; run 'pickaxe init' first")
		}

		if err := registry.Remove(cfg, name); err != nil {
			return err
		}

		if err := config.WriteProjectConfig(registryPath, cfg); err != nil {
			return fmt.Errorf("write .pickaxe.json: %w", err)
		}

		fmt.Printf("removed: %s\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

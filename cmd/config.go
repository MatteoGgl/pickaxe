package cmd

import (
	"fmt"

	"github.com/matteoggl/pickaxe/internal/config"
	"github.com/matteoggl/pickaxe/internal/pathutil"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"c"},
	Short: "Manage global pickaxe configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		if key != "vault" {
			return fmt.Errorf("unknown config key %q; supported keys: vault", key)
		}

		expanded, err := pathutil.ExpandHome(value)
		if err != nil {
			return err
		}

		path := config.DefaultGlobalConfigPath()
		cfg, err := config.ReadGlobalConfig(path)
		if err != nil {
			cfg = &config.GlobalConfig{}
		}

		cfg.VaultRoot = expanded
		if err := config.WriteGlobalConfig(path, cfg); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
		fmt.Printf("vault root set to: %s\n", cfg.VaultRoot)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}

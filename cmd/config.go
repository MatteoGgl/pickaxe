package cmd

import (
	"fmt"
	"os"

	"github.com/matteo/pickaxe/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
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

		path := config.DefaultGlobalConfigPath()
		cfg, err := config.ReadGlobalConfig(path)
		if err != nil {
			cfg = &config.GlobalConfig{}
		}

		if len(value) > 1 && value[:2] == "~/" {
			home, _ := os.UserHomeDir()
			value = home + value[1:]
		}

		cfg.VaultRoot = value
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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/matteo/pickaxe/internal/config"
	"github.com/matteo/pickaxe/internal/registry"
	"github.com/matteo/pickaxe/internal/tui"
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
			globalCfgPath := config.DefaultGlobalConfigPath()
			globalCfg, err := config.ReadGlobalConfig(globalCfgPath)
			if err != nil {
				return fmt.Errorf("vault root not configured; run 'pickaxe config set vault <path>' first")
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

			preSelected := map[string]bool{}
			for _, entry := range cfg.Entries {
				preSelected[entry.Path] = true
			}

			result, err := tui.Run(globalCfg.VaultRoot, preSelected)
			if err != nil {
				return fmt.Errorf("picker error: %w", err)
			}
			if !result.Confirmed || len(result.Selections) == 0 {
				fmt.Println("no changes made")
				return nil
			}

			changed := false
			for _, sel := range result.Selections {
				if preSelected[sel.Path] {
					continue
				}
				if sel.IsDir {
					if err := registry.AddDir(cfg, sel.Path, "", false); err != nil {
						fmt.Fprintf(os.Stderr, "warning: could not add %s: %v\n", sel.Path, err)
						continue
					}
				} else {
					if err := registry.AddFile(cfg, sel.Path, ""); err != nil {
						fmt.Fprintf(os.Stderr, "warning: could not add %s: %v\n", sel.Path, err)
						continue
					}
				}
				fmt.Printf("registered: %s\n", sel.Path)
				changed = true
			}

			if !changed {
				fmt.Println("no new entries added")
				return nil
			}

			return config.WriteProjectConfig(registryPath, cfg)
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

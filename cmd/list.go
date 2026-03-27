package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/matteo/pickaxe/internal/config"
	"github.com/matteo/pickaxe/internal/registry"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered vault entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		registryPath := filepath.Join(cwd, config.ProjectConfigFilename)

		cfg, err := config.ReadProjectConfig(registryPath)
		if err != nil {
			return fmt.Errorf("no .pickaxe.json found; run 'pickaxe init' first")
		}

		if len(cfg.Entries) == 0 {
			fmt.Println("no entries registered")
			return nil
		}

		for _, entry := range cfg.Entries {
			files, err := registry.EnumerateFiles(entry)
			if err != nil {
				fmt.Printf("  [%s] %s — ERROR: %v\n", entry.Type, entry.Name, err)
				continue
			}
			if entry.Type == config.EntryTypeDir {
				fmt.Printf("  [dir] %s (%s", entry.Name, entry.Path)
				if entry.Recursive {
					fmt.Print(", recursive")
				}
				fmt.Printf(") — %d file(s)\n", len(files))
			} else {
				for _, f := range files {
					status := "ok"
					if f.Unavailable {
						status = "UNAVAILABLE"
					}
					fmt.Printf("  [file] %s (%s) — %s\n", f.Name, f.Path, status)
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

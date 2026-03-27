package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/matteo/pickaxe/internal/vault"
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

		v, err := vault.Open(cwd)
		if err != nil {
			return fmt.Errorf("no .pickaxe.json found; run 'pickaxe init' first")
		}

		entries := v.Entries()
		if len(entries) == 0 {
			fmt.Println("no entries registered")
			return nil
		}

		allFiles, listErr := vault.ListFiles(cwd)

		for _, entry := range entries {
			if listErr != nil {
				fmt.Printf("  [%s] %s — ERROR: %v\n", entry.Type, entry.Name, listErr)
				continue
			}
			if entry.Type == vault.EntryTypeDir {
				prefix := entry.Name + "/"
				count := 0
				for _, f := range allFiles {
					if strings.HasPrefix(f.Name, prefix) {
						count++
					}
				}
				fmt.Printf("  [dir] %s (%s", entry.Name, entry.Path)
				if entry.Recursive {
					fmt.Print(", recursive")
				}
				fmt.Printf(") — %d file(s)\n", count)
			} else {
				status := "ok"
				for _, f := range allFiles {
					if f.Name == entry.Name && f.Unavailable {
						status = "UNAVAILABLE"
					}
				}
				fmt.Printf("  [file] %s (%s) — %s\n", entry.Name, entry.Path, status)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

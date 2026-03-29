package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/matteoggl/pickaxe/internal/vault"
	"github.com/spf13/cobra"
)

var (
	styleBold = lipgloss.NewStyle().Bold(true)
	styleDim  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleFile = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // cyan
	styleDir  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
	styleOK   = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	styleBad  = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List all registered vault entries",
	Args:    cobra.NoArgs,
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
		ht := vault.NewHashTable(entries)

		for _, entry := range entries {
			shortHash := ht.ShortHash(entry.Name)
			plen := ht.ShortPrefixLen(entry.Name)
			styledHash := styleBold.Render(shortHash[:plen]) + styleDim.Render(shortHash[plen:])

			perm := styleDim.Render("r-")
			if entry.Writable {
				perm = styleOK.Render("rw")
			}

			if listErr != nil {
				badge := styleFile.Render("[file]")
				if entry.Type == vault.EntryTypeDir {
					badge = styleDir.Render("[dir]")
				}
				fmt.Printf("  %s %s %s %s — %s\n", styledHash, badge, perm, styleBold.Render(entry.Name), styleBad.Render("ERROR: "+listErr.Error()))
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
				meta := styleDim.Render(entry.Path)
				if entry.Recursive {
					meta += styleDim.Render(", recursive")
				}
				fileCount := styleDim.Render(fmt.Sprintf("%d files", count))
				fmt.Printf("  %s %s %s %s %s — %s\n", styledHash, styleDir.Render("[dir]"), perm, styleBold.Render(entry.Name), meta, fileCount)
			} else {
				unavailable := false
				for _, f := range allFiles {
					if f.Name == entry.Name && f.Unavailable {
						unavailable = true
					}
				}
				var status string
				if unavailable {
					status = styleBad.Render("unavailable")
				} else {
					status = styleOK.Render("✓")
				}
				fmt.Printf("  %s %s %s %s %s %s\n", styledHash, styleFile.Render("[file]"), perm, styleBold.Render(entry.Name), styleDim.Render(entry.Path), status)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

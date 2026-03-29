package cmd

import (
	"fmt"
	"io"
	"maps"
	"os"
	"sort"

	"github.com/matteoggl/pickaxe/internal/config"
	"github.com/matteoggl/pickaxe/internal/tui"
	"github.com/matteoggl/pickaxe/internal/vault"
	"github.com/spf13/cobra"
)

// PickerFunc is the signature for launching the TUI file picker.
type PickerFunc func(root string, preSelected map[string]bool, preWritable map[string]bool) (tui.PickerResult, error)

// PickFromVault runs the interactive pick flow: expand current entries,
// show picker, compact result, replace vault entries.
func PickFromVault(v *vault.Vault, vaultRoot string, picker PickerFunc, w io.Writer) error {
	expanded, expandedWritable, err := vault.ExpandEntries(v.Entries())
	if err != nil {
		return fmt.Errorf("expand entries: %w", err)
	}

	result, err := picker(vaultRoot, expanded, expandedWritable)
	if err != nil {
		return fmt.Errorf("picker error: %w", err)
	}
	if !result.Confirmed {
		fmt.Fprintln(w, "no changes made")
		return nil
	}

	var selectedPaths []string
	writablePaths := make(map[string]bool)
	for _, sel := range result.Selections {
		selectedPaths = append(selectedPaths, sel.Path)
		if sel.Writable {
			writablePaths[sel.Path] = true
		}
	}

	compacted, err := vault.Compact(selectedPaths, writablePaths)
	if err != nil {
		return fmt.Errorf("compact: %w", err)
	}

	oldPaths := entryPathSet(v.Entries())
	newPaths := entryPathSet(compacted)
	oldWritable := entryWritableSet(v.Entries())
	newWritable := entryWritableSet(compacted)
	if maps.Equal(oldPaths, newPaths) && maps.Equal(oldWritable, newWritable) {
		fmt.Fprintln(w, "no changes")
		return nil
	}

	v.ReplaceEntries(compacted)
	if err := v.Save(); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	added := sortedKeys(setDiff(newPaths, oldPaths))
	removed := sortedKeys(setDiff(oldPaths, newPaths))
	for _, p := range added {
		fmt.Fprintf(w, "added: %s\n", p)
	}
	for _, p := range removed {
		fmt.Fprintf(w, "removed: %s\n", p)
	}

	return nil
}

func entryPathSet(entries []vault.Entry) map[string]bool {
	m := make(map[string]bool, len(entries))
	for _, e := range entries {
		m[e.Path] = true
	}
	return m
}

func entryWritableSet(entries []vault.Entry) map[string]bool {
	m := make(map[string]bool, len(entries))
	for _, e := range entries {
		if e.Writable {
			m[e.Path] = true
		}
	}
	return m
}

func setDiff(a, b map[string]bool) map[string]bool {
	result := make(map[string]bool)
	for k := range a {
		if !b[k] {
			result[k] = true
		}
	}
	return result
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

var pickCmd = &cobra.Command{
	Use:     "pick",
	Aliases: []string{"p"},
	Short:   "Interactively add or remove vault entries",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		globalCfgPath, err := config.DefaultGlobalConfigPath()
		if err != nil {
			return err
		}
		globalCfg, err := config.ReadGlobalConfig(globalCfgPath)
		if err != nil || globalCfg.VaultRoot == "" {
			return fmt.Errorf("vault root not configured; run 'pickaxe config set vault <path>' first")
		}
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		v, err := vault.Open(cwd)
		if err != nil {
			return fmt.Errorf("no .pickaxe.json found; run 'pickaxe init' first")
		}
		return PickFromVault(v, globalCfg.VaultRoot, func(root string, preSelected, preWritable map[string]bool) (tui.PickerResult, error) {
			return tui.Run(root, preSelected, preWritable)
		}, os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(pickCmd)
}

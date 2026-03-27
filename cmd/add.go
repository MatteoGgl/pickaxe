package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/matteo/pickaxe/internal/config"
	"github.com/matteo/pickaxe/internal/pathutil"
	"github.com/matteo/pickaxe/internal/tui"
	"github.com/matteo/pickaxe/internal/vault"
	"github.com/spf13/cobra"
)

var addAlias string
var addRecursive bool

// AddFromPicker runs the interactive TUI add flow.
func AddFromPicker(v *vault.Vault, vaultRoot string, picker PickerFunc, w io.Writer) error {
	preSelected := vault.BuildPreSelected(v.Entries())

	result, err := picker(vaultRoot, preSelected)
	if err != nil {
		return fmt.Errorf("picker error: %w", err)
	}
	if !result.Confirmed || len(result.Selections) == 0 {
		fmt.Fprintln(w, "no changes made")
		return nil
	}

	var added int
	for _, sel := range result.Selections {
		if preSelected[sel.Path] {
			continue
		}
		if sel.IsDir {
			if err := v.AddDir(sel.Path, "", false); err != nil {
				fmt.Fprintf(w, "warning: could not add %s: %v\n", sel.Path, err)
				continue
			}
		} else {
			if err := v.AddFile(sel.Path, ""); err != nil {
				fmt.Fprintf(w, "warning: could not add %s: %v\n", sel.Path, err)
				continue
			}
		}
		fmt.Fprintf(w, "registered: %s\n", sel.Path)
		added++
	}

	if added == 0 {
		fmt.Fprintln(w, "no new entries added")
		return nil
	}

	return v.Save()
}

// AddDirect registers a single file or directory.
func AddDirect(v *vault.Vault, absPath string, isDir bool, alias string, recursive bool, w io.Writer) error {
	if isDir {
		if err := v.AddDir(absPath, alias, recursive); err != nil {
			return err
		}
	} else {
		if err := v.AddFile(absPath, alias); err != nil {
			return err
		}
	}
	if err := v.Save(); err != nil {
		return fmt.Errorf("write .pickaxe.json: %w", err)
	}
	fmt.Fprintf(w, "registered: %s\n", absPath)
	return nil
}

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

			v, err := vault.Open(cwd)
			if err != nil {
				return fmt.Errorf("no .pickaxe.json found; run 'pickaxe init' first")
			}

			return AddFromPicker(v, globalCfg.VaultRoot, tui.Run, os.Stdout)
		}

		rawPath := args[0]

		absPath, err := pathutil.ExpandAndResolve(rawPath)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		v, err := vault.Open(cwd)
		if err != nil {
			return fmt.Errorf("no .pickaxe.json found; run 'pickaxe init' first")
		}

		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("%q not found", absPath)
		}

		isDir := info.IsDir() || strings.HasSuffix(rawPath, "/")
		return AddDirect(v, absPath, isDir, addAlias, addRecursive, os.Stdout)
	},
}

func init() {
	addCmd.Flags().StringVar(&addAlias, "as", "", "Override the short name for this entry")
	addCmd.Flags().BoolVar(&addRecursive, "recursive", false, "Register directory recursively")
	rootCmd.AddCommand(addCmd)
}

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/matteo/pickaxe/internal/config"
	"github.com/matteo/pickaxe/internal/pathutil"
	"github.com/matteo/pickaxe/internal/tui"
	"github.com/matteo/pickaxe/internal/vault"
	"github.com/matteo/pickaxe/internal/workflow"
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

			v, err := vault.Open(cwd)
			if err != nil {
				return fmt.Errorf("no .pickaxe.json found; run 'pickaxe init' first")
			}

			preSelected := workflow.BuildPreSelected(v.Entries())

			result, err := tui.Run(globalCfg.VaultRoot, preSelected)
			if err != nil {
				return fmt.Errorf("picker error: %w", err)
			}
			if !result.Confirmed || len(result.Selections) == 0 {
				fmt.Println("no changes made")
				return nil
			}

			sels := make([]workflow.Selection, len(result.Selections))
			for i, s := range result.Selections {
				sels[i] = workflow.Selection{Path: s.Path, IsDir: s.IsDir}
			}
			plan := workflow.DiffSelections(v.Entries(), sels)

			if len(plan.Additions) == 0 {
				fmt.Println("no new entries added")
				return nil
			}

			for _, sel := range plan.Additions {
				if sel.IsDir {
					if err := v.AddDir(sel.Path, "", false); err != nil {
						fmt.Fprintf(os.Stderr, "warning: could not add %s: %v\n", sel.Path, err)
						continue
					}
				} else {
					if err := v.AddFile(sel.Path, ""); err != nil {
						fmt.Fprintf(os.Stderr, "warning: could not add %s: %v\n", sel.Path, err)
						continue
					}
				}
				fmt.Printf("registered: %s\n", sel.Path)
			}

			return v.Save()
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

		if info.IsDir() || strings.HasSuffix(rawPath, "/") {
			if err := v.AddDir(absPath, addAlias, addRecursive); err != nil {
				return err
			}
		} else {
			if err := v.AddFile(absPath, addAlias); err != nil {
				return err
			}
		}

		if err := v.Save(); err != nil {
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

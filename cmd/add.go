package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/matteoggl/pickaxe/internal/pathutil"
	"github.com/matteoggl/pickaxe/internal/vault"
	"github.com/spf13/cobra"
)

var addAlias string
var addRecursive bool

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
	Use:     "add <path>",
	Aliases: []string{"a"},
	Short: "Register a vault file or directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

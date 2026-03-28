package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/matteoggl/pickaxe/internal/vault"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Create a .pickaxe.json registry in the current directory",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		if _, err := vault.Init(cwd); errors.Is(err, vault.ErrAlreadyExists) {
			return errors.New(".pickaxe.json already exists")
		} else if err != nil {
			return fmt.Errorf("create .pickaxe.json: %w", err)
		}
		fmt.Printf("created %s\n", filepath.Join(cwd, vault.ConfigFilename))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

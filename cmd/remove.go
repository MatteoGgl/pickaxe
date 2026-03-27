package cmd

import (
	"fmt"
	"os"

	"github.com/matteo/pickaxe/internal/vault"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name|hash>",
	Short: "Remove a registered vault entry by name or hash prefix",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		v, err := vault.Open(cwd)
		if err != nil {
			return fmt.Errorf("no .pickaxe.json found; run 'pickaxe init' first")
		}

		resolved, err := v.ResolveIdentifier(id)
		if err != nil {
			return err
		}

		if err := v.Remove(resolved); err != nil {
			return err
		}

		if err := v.Save(); err != nil {
			return fmt.Errorf("write .pickaxe.json: %w", err)
		}

		fmt.Printf("removed: %s\n", resolved)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

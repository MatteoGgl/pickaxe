package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/matteoggl/pickaxe/internal/vault"
	"github.com/spf13/cobra"
)

func setEntryWritable(v *vault.Vault, id string, writable bool, w io.Writer) error {
	name, err := v.ResolveIdentifier(id)
	if err != nil {
		return err
	}
	if err := v.SetWritable(name, writable); err != nil {
		return err
	}
	if writable {
		fmt.Fprintf(w, "%q is now read-write\n", name)
	} else {
		fmt.Fprintf(w, "%q is now read-only\n", name)
	}
	return nil
}

// LockEntry marks the identified entry as read-only (writable=false).
func LockEntry(v *vault.Vault, id string, w io.Writer) error {
	return setEntryWritable(v, id, false, w)
}

// UnlockEntry marks the identified entry as read-write (writable=true).
func UnlockEntry(v *vault.Vault, id string, w io.Writer) error {
	return setEntryWritable(v, id, true, w)
}

var lockCmd = &cobra.Command{
	Use:     "lock <name|hash>",
	Aliases: []string{"lk"},
	Short: "Mark a vault entry as read-only",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		v, err := vault.Open(cwd)
		if err != nil {
			return fmt.Errorf("no .pickaxe.json found; run 'pickaxe init' first")
		}
		if err := LockEntry(v, args[0], cmd.OutOrStdout()); err != nil {
			return err
		}
		return v.Save()
	},
}

var unlockCmd = &cobra.Command{
	Use:     "unlock <name|hash>",
	Aliases: []string{"ul"},
	Short: "Mark a vault entry as read-write",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		v, err := vault.Open(cwd)
		if err != nil {
			return fmt.Errorf("no .pickaxe.json found; run 'pickaxe init' first")
		}
		if err := UnlockEntry(v, args[0], cmd.OutOrStdout()); err != nil {
			return err
		}
		return v.Save()
	},
}

func init() {
	rootCmd.AddCommand(lockCmd)
	rootCmd.AddCommand(unlockCmd)
}

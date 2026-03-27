package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/matteo/pickaxe/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a .pickaxe.json registry in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path := filepath.Join(cwd, config.ProjectConfigFilename)

		if _, err := os.Stat(path); err == nil {
			return errors.New(".pickaxe.json already exists")
		}

		cfg := &config.ProjectConfig{Version: 1, Entries: []config.Entry{}}
		if err := config.WriteProjectConfig(path, cfg); err != nil {
			return fmt.Errorf("create .pickaxe.json: %w", err)
		}
		fmt.Printf("created %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

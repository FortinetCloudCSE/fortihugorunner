package cmd

import (
	"fmt"
	"fortihugorunner/utilities"
	"github.com/spf13/cobra"
	"os"
)

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename to fortihugorunner; i.e. trim OS/arch suffix (e.g. fortihugorunner-linux-amd64 → fortihugorunner)",
	RunE: func(cmd *cobra.Command, args []string) error {
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("could not get executable path: %w", err)
		}

		err = utilities.RenameBinary(exePath)
		if err != nil {
			return fmt.Errorf("error renaming binary: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}

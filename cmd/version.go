package cmd

import (
	"fortihugorunner/version"
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print fortihugorunner version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nDate: %s\n", version.Version, version.Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

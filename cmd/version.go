package cmd

import (
	"fmt"
	"fortihugorunner/version"
	"github.com/spf13/cobra"
	"runtime"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print fortihugorunner version.",
	Run: func(cmd *cobra.Command, args []string) {
		osType := runtime.GOOS
		arch := runtime.GOARCH
		platform := osType + "/" + arch
		fmt.Printf("Version: %s\nDate: %s\nPlatform: %s\n", version.Version, version.Date, platform)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

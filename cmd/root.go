package cmd

import (
	"docker-run-go/version"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootVersion bool

var rootCmd = &cobra.Command{
	Use:   "docker-run-go",
	Short: "FortinetCloudCSE Workshop Docker development utility.",
	Long:  "Includes functions for facilitating Hugo app development with docker containers.",
	Run: func(cmd *cobra.Command, args []string) {
		if rootVersion {
			fmt.Printf("Version: %s\nDate: %s\n", version.Version, version.Date)
			os.Exit(0)
		}
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootVersion, "version", "v", false, "docker-run-go version information")
}

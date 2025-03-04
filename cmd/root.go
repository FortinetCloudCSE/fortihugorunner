package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "docker-run-go",
	Short: "FortinetCloudCSE Workshop Docker development utility.",
	Long:  "Includes functions for facilitating Hugo app development with docker containers.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(launchServerCmd)
	rootCmd.AddCommand(createContentCmd)
}

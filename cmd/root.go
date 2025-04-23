package cmd

import (
	"context"
	"docker-run-go/version"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"os"
)

var rootVersion bool

var rootCmd = &cobra.Command{
	Use:   "docker-run-go",
	Short: "FortinetCloudCSE Workshop Docker development utility.",
	Long:  "Includes functions for facilitating Hugo app development with docker containers.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := checkDockerRunning()
		if err != nil {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			fmt.Fprintf(os.Stderr, "\nReceived the error below. For troubleshooting help, head here: https://docs.docker.com/engine/daemon/troubleshoot/\n\n")
			return err
		}
		return nil

	},
	Run: func(cmd *cobra.Command, args []string) {
		if rootVersion {
			fmt.Printf("Version: %s\nDate: %s\n", version.Version, version.Date)
			os.Exit(0)
		}
		cmd.Help()
	},
}

func checkDockerRunning() error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not create Docker client: %w", err)
	}

	_, err = cli.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
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

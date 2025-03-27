package cmd

import (
        "context"
	"docker-run-go/version"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"os"
        "docker-run-go/dockerinternal"
)

var rootVersion bool

var dockerClient dockerinternal.DockerClient

var rootCmd = &cobra.Command{
	Use:   "docker-run-go",
	Short: "FortinetCloudCSE Workshop Docker development utility.",
	Long:  "Includes functions for facilitating Hugo app development with docker containers.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
                var err error
                dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	        if err != nil {
	        	return fmt.Errorf("could not create Docker client: %w", err)
	        }
                ctx := context.Background()

                //Check if Docker is running
                if realClient, ok := dockerClient.(*client.Client); ok {
                        _, err:= realClient.Ping(ctx)                 
                        if err != nil {
                            cmd.SilenceErrors = true
                            cmd.SilenceUsage = true
                            fmt.Fprintf(os.Stderr, "\nReceived the error below. For troubleshooting help, head here: https://docs.docker.com/engine/daemon/troubleshoot/\n\n")
                            return err
                        }
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

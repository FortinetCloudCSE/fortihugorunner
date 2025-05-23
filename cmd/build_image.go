package cmd

import (
	"fmt"
	"os"

	"docker-run-go/pkg/dockerinternal"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

// buildImageCmd represents the `build-image` command using Docker SDK
var buildImageCmd = &cobra.Command{
	Use:   "build-image",
	Short: "Builds a Docker image programmatically using the Docker SDK",
	Long: `Builds a Docker image with the specified environment.

Example:
  docker-run-go build-image --env author-dev
  docker-run-go build-image --env admin-dev --hugo-version 0.146.0
`,
	//Args: cobra.ExactArgs(1), // Require exactly one argument
	Run: func(cmd *cobra.Command, args []string) {
		//envArg := args[0]
		envArg, _ := cmd.Flags().GetString("env")
		hugoVersion, _ := cmd.Flags().GetString("hugo-version")

		// Map provided argument to actual Docker build target
		envMap := map[string]string{
			"author-dev": "prod",
			"admin-dev":  "dev",
		}
		env, exists := envMap[envArg]
		if !exists {
			fmt.Println("Error: env must be one of either author-dev or admin-dev.")
			os.Exit(1)
		}

		// Determine the corresponding container name
		containerMap := map[string]string{
			"prod": "fortinet-hugo",
			"dev":  "hugotester",
		}
		containerName := containerMap[env]

		// Initialize Docker client
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("Error creating Docker client: %v\n", err)
			os.Exit(1)
		}

		// Build the Docker image
		err = dockerinternal.BuildDockerImage(cli, containerName, env, envArg, hugoVersion)
		if err != nil {
			fmt.Printf("Error building Docker image: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("**** Built a %s container named: %s ****\n", envArg, containerName)
	},
}

func init() {
	rootCmd.AddCommand(buildImageCmd)
	buildImageCmd.Flags().String("env", "author-dev", "Environment. author-dev (prod) creates a fortinet-hugo image. admin-dev (dev) creates a hugotester image.")
	buildImageCmd.Flags().String("hugo-version", "std", "Hugo base image version Go will pull before proceeding to build the <env> image. This must match the hugomods/hugo tag referenced in your Dockerfile.")
}

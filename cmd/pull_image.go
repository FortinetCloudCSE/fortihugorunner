package cmd

import (
	"context"
	"fmt"
	"os"

	"fortihugorunner/dockerinternal"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

// buildImageCmd represents the `build-image` command using Docker SDK
var pullImageCmd = &cobra.Command{
	Use:   "pull-image",
	Short: "Pulls and re-tags the latest Hugo development images.",
	Long: `Pulls and re-tags the latest prebuilt Docker images for Hugo workshop development from our public ECR repositories."

Example:
  fortihugorunner pull-image --env author-dev
  fortihugorunner pull-image --env admin-dev
`,
	//Args: cobra.ExactArgs(1), // Require exactly one argument
	Run: func(cmd *cobra.Command, args []string) {

		//envArg := args[0]
		envArg, _ := cmd.Flags().GetString("env")
		ecrReg, _ := cmd.Flags().GetString("registry")

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

		// Pull the Docker image
		fullUri := ecrReg + containerName + ":latest"
		err = dockerinternal.EnsureImagePulled(cli, fullUri)
		if err != nil {
			fmt.Printf("Error pulling Docker image: %v\n", err)
			os.Exit(1)
		}

		// Tag the image
		err = cli.ImageTag(context.Background(), fullUri, containerName)
		if err != nil {
			fmt.Printf("Error re-tagging image: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("**** Image %s successfully pulled and tagged as: %s ****\n", fullUri, containerName)
	},
}

func init() {
	rootCmd.AddCommand(pullImageCmd)
	pullImageCmd.Flags().String("env", "author-dev", "Environment. author-dev (prod) creates a fortinet-hugo image. admin-dev (dev) creates a hugotester image.")
	pullImageCmd.Flags().String("registry", "public.ecr.aws/k4n6m5h8/", "ECR registry.")
}

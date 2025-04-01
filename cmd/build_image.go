package cmd

import (
	"fmt"
	//"os"

	"docker-run-go/dockerinternal"
	"github.com/spf13/cobra"
)

// buildImageCmd represents the `build-image` command using Docker SDK
func newBuildImageCmd(cli dockerinternal.DockerClient) *cobra.Command {
   return &cobra.Command{
	Use:   "build-image [author-dev | admin-dev]",
	Short: "Builds a Docker image programmatically using the Docker SDK",
	Long: `Builds a Docker image with the specified environment.

Example:
  docker-run-go build-image author-dev
  docker-run-go build-image admin-dev
`,
	Args: cobra.ExactArgs(1), // Require exactly one argument
	RunE: func(cmd *cobra.Command, args []string) error {
                
		envArg := args[0]

		// Map provided argument to actual Docker build target
		envMap := map[string]string{
			"author-dev": "prod",
			"admin-dev":  "dev",
		}
		env, exists := envMap[envArg]
		if !exists {
                        errString := fmt.Sprintf("Usage: docker-run-go build-image [author-dev | admin-dev]")
			cmd.Println(errString)
			return fmt.Errorf(errString)
		}

		// Determine the corresponding container name
		containerMap := map[string]string{
			"prod": "fortinet-hugo",
			"dev":  "hugotester",
		}
		containerName := containerMap[env]

		// Build the Docker image
		err := dockerinternal.BuildDockerImage(cli, containerName, env, envArg)
		if err != nil {
			cmd.Printf("Error building Docker image: %v\n", err)
			return err
		}

                printString := fmt.Sprintf("**** Built a %s container named: %s ****\n", envArg, containerName)
		cmd.Println(printString)
                return nil
	},
     }
}

func init() {
	rootCmd.AddCommand(newBuildImageCmd(dockerClient))
}

func NewTestableBuildImageCmd(cli dockerinternal.DockerClient) *cobra.Command {
	return newBuildImageCmd(cli)
}


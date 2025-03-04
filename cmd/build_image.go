package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// buildImageCmd represents the `build-image` command
var buildImageCmd = &cobra.Command{
	Use:   "build-image [admin-dev | shop-dev]",
	Short: "Builds a Docker image from a Dockerfile",
	Long: `Builds a Docker image with the specified environment.

Example:
  docker-run-go build-image admin-dev
  docker-run-go build-image shop-dev
`,
	Args: cobra.ExactArgs(1), // Require exactly one argument
	Run: func(cmd *cobra.Command, args []string) {
		env := args[0]

		// Define valid environments
		validEnvs := map[string]string{
			"admin-dev": "fortinet-hugo",
			"shop-dev":  "hugotester",
		}

		// Validate the provided environment
		containerName, exists := validEnvs[env]
		if !exists {
			fmt.Println("Usage: docker-run-go build-image [admin-dev | shop-dev]")
			os.Exit(1)
		}

		targetMap := map[string]string{
			"admin-dev": "prod",
			"shop-dev":  "dev",
		}
		target := targetMap[env]

		// Construct the docker build command
		dockerCmd := exec.Command("docker", "build", "-t", containerName, ".", "--target="+target)

		// Attach stdout and stderr to the command output
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr

		fmt.Printf("Building a %s container named: %s\n", env, containerName)

		// Execute the command
		err := dockerCmd.Run()
		if err != nil {
			fmt.Printf("Error building the image: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("**** Built a %s container named: %s ****\n", env, containerName)
	},
}

func init() {
	rootCmd.AddCommand(buildImageCmd)
}

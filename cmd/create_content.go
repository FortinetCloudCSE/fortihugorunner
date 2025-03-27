package cmd

import (
	"context"
	"fmt"
	"os"

	"docker-run-go/dockerinternal"
	"github.com/docker/docker/api/types/container"
	"github.com/spf13/cobra"
)

var createContentCmd = &cobra.Command{
	Use:   "create-content",
	Short: "Create content using a Docker container",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := dockerinternal.ContentConfig{
			DockerImage: getFlagString(cmd, "docker-image"),
		}
		ctx := context.Background()

		// Example: run a container with a different command.
		containerConfig := &container.Config{
			Image: cfg.DockerImage,
			Cmd:   []string{"create", "content"},
			Tty:   true,
		}
		created, err := dockerClient.ContainerCreate(ctx, containerConfig, nil, nil, nil, "")
		if err != nil {
			fmt.Printf("Error creating container: %v\n", err)
			os.Exit(1)
		}
		if err := dockerClient.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
			fmt.Printf("Error starting container: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Started content creation container: %s\n", created.ID)
	},
}

func init() {
	rootCmd.AddCommand(createContentCmd)
	createContentCmd.Flags().String("docker-image", "mycontent-image:latest", "Docker image to use for creating content")
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"fortihugorunner/dockerinternal"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

func getFlagString(cmd *cobra.Command, flagName string) string {
	value, _ := cmd.Flags().GetString(flagName)
	return value
}

func getFlagBool(cmd *cobra.Command, flagName string) bool {
	value, _ := cmd.Flags().GetBool(flagName)
	return value
}

var launchServerCmd = &cobra.Command{
	Use:   "launch-server",
	Short: "Launch the Hugo server container",
	Long: `Launch the Hugo server container based on specified image and other parameters.

Example:
  ./fortihugorunner launch-server \
      --docker-image fortinet-hugo:latest \
      --host-port 1313 \
      --container-port 1313 \
      --watch-dir . \
      --mount-toml
`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := dockerinternal.ServerConfig{
			DockerImage:   getFlagString(cmd, "docker-image"),
			HostPort:      getFlagString(cmd, "host-port"),
			ContainerPort: getFlagString(cmd, "container-port"),
			WatchDir:      getFlagString(cmd, "watch-dir"),
			MountToml:     getFlagBool(cmd, "mount-toml"),
		}

		// Ensure the watch directory is absolute.
		abs, err := filepath.Abs(cfg.WatchDir)
		if err == nil {
			cfg.WatchDir = abs
		}

		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("Error creating Docker client: %v\n", err)
			os.Exit(1)
		}

		containerID, err := dockerinternal.StartContainer(ctx, cli, cfg)
		if err != nil {
			fmt.Printf("Error starting container: %v\n", err)
			os.Exit(1)
		}

		if err := dockerinternal.AttachContainer(ctx, cli, containerID); err != nil {
			fmt.Printf("Error attaching container: %v\n", err)
			os.Exit(1)
		}

		// Setup signal handling.
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Println("\nReceived shutdown signal. Stopping container.")
			dockerinternal.StopAndRemoveContainer(cli, containerID)
			os.Exit(0)
		}()

		// Start file watcher.
		dockerinternal.WatchAndRestart(ctx, cli, cfg, &containerID)
	},
}

func init() {
	rootCmd.AddCommand(launchServerCmd)
	launchServerCmd.Flags().String("docker-image", "fortinet-hugo:latest", "Docker image to use")
	launchServerCmd.Flags().String("host-port", "1313", "Host port to expose")
	launchServerCmd.Flags().String("container-port", "1313", "Container port to expose")
	launchServerCmd.Flags().String("watch-dir", ".", "Directory to watch for file changes")
	launchServerCmd.Flags().Bool("mount-toml", false, "Mount the hugo.toml in your workshop directory and watch for updates. (Default is false)")
}

package dockerinternal

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type ServerConfig struct {
	DockerImage   string
	HostPort      string
	ContainerPort string
	WatchDir      string
}

type ContentConfig struct {
	DockerImage string
	// Add other flags as needed.
}

func StartContainer(ctx context.Context, cli *client.Client, cfg ServerConfig) (string, error) {
	// Adjust the path for mounting.
	userRepoPath := AdjustPathForDocker(cfg.WatchDir)
	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: userRepoPath,
			Target: "/home/UserRepo",
		},
	}

	// Mount the Hugo configuration file.
	configPath := filepath.Join(cfg.WatchDir, "hugo.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Warning: Hugo config file not found at %s. The container may exit if Hugo requires it.\n", configPath)
	}
	centralRepoPath := AdjustPathForDocker(configPath)
	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: centralRepoPath,
		Target: "/home/CentralRepo/hugo.toml",
	})

	containerConfig := &container.Config{
		Image: cfg.DockerImage,
		Cmd:   []string{"server", "--bind", "0.0.0.0", "--liveReload", "--disableFastRender", "--poll"},
		Tty:   true,
		ExposedPorts: nat.PortSet{
			nat.Port(cfg.ContainerPort + "/tcp"): struct{}{},
		},
	}
	hostConfig := &container.HostConfig{
		Mounts: mounts,
		PortBindings: nat.PortMap{
			nat.Port(cfg.ContainerPort + "/tcp"): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: cfg.HostPort,
				},
			},
		},
	}

	created, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("container create error: %w", err)
	}
	if err := cli.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("container start error: %w", err)
	}
	fmt.Printf("Started container: %s\n", created.ID)
	return created.ID, nil
}

func AttachContainer(ctx context.Context, cli *client.Client, containerID string) error {
	opts := container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
		Stdin:  true,
	}
	resp, err := cli.ContainerAttach(ctx, containerID, opts)
	if err != nil {
		return fmt.Errorf("container attach error: %w", err)
	}
	go func() {
		_, _ = io.Copy(os.Stdout, resp.Reader)
	}()
	go func() {
		_, _ = io.Copy(resp.Conn, os.Stdin)
	}()
	return nil
}

func StopAndRemoveContainer(cli *client.Client, containerID string) {
	fmt.Printf("Stopping container: %s\n", containerID)
	timeout := 10
	stopOpts := container.StopOptions{Timeout: &timeout}
	if err := cli.ContainerStop(context.Background(), containerID, stopOpts); err != nil {
		fmt.Printf("Error stopping container %s: %v\n", containerID, err)
	}
	if err := cli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true}); err != nil {
		fmt.Printf("Error removing container %s: %v\n", containerID, err)
	}
}

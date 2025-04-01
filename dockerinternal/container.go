package dockerinternal

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
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
}

func EnsureImagePulled(cli DockerClient, imageName string) error {
	ctx := context.Background()

	fmt.Printf("Ensuring frontend image %s is available...\n", imageName)
        
        fmt.Printf("Using client %T in EnsureImagePulled...\n", cli)
	out, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull frontend image: %w", err)
	}
	defer out.Close()

	// Stream output to the console
	_, err = io.Copy(os.Stdout, out)
	if err != nil {
		return fmt.Errorf("error reading image pull output: %w", err)
	}

	fmt.Println("Frontend image pulled successfully.")
	return nil
}

// buildDockerImage builds the Docker image using the SDK
func BuildDockerImage(cli DockerClient, imageName string, target string, envArg string) error {

        fmt.Printf("Using client %T in BuildDockerImage...\n", cli)
	branchMap := map[string]string{
		"admin-dev":  "prreviewJune23",
		"author-dev": "main",
	}

	branchWorking := branchMap[envArg]

	images := []string{
		"docker/dockerfile:1.5-labs",
		"docker.io/hugomods/hugo:std",
	}

	for _, img := range images {
		if err := EnsureImagePulled(cli, img); err != nil {
			//panic(err)
                        fmt.Printf("Error pulling image %s", imageName)
                        return fmt.Errorf("Couldn't pull required image, exiting....")
		}
	}
        
	// Create a tarball of the current directory (Docker build context)
	tarBuffer, err := CreateTarball(".")
	if err != nil {
		return fmt.Errorf("error creating tarball: %w", err)
	}

	// Define build options
	options := types.ImageBuildOptions{
		Tags:       []string{imageName},
		Dockerfile: "Dockerfile",
		Target:     target,
		Remove:     true,
		Version:    types.BuilderBuildKit,
		BuildArgs: map[string]*string{
			"BUILDKIT_INLINE_CACHE": strPtr("1"),
			"DOCKER_BUILDKIT":       strPtr("1"),
		},
	}

	// Execute Docker build
	ctx := context.Background()
	_, err = cli.BuildCachePrune(ctx, types.BuildCachePruneOptions{})

	response, err := cli.ImageBuild(ctx, tarBuffer, options)
	if err != nil {
		return fmt.Errorf("error building image: %w", err)
	}
	defer response.Body.Close()

	// Stream build output
	_, err = io.Copy(os.Stdout, response.Body)
	if err != nil {
		return fmt.Errorf("error reading build output: %w", err)
	}

	fmt.Println("Image built with CentralRepo branch: ", branchWorking)

	return nil
}

func strPtr(s string) *string {
	return &s
}

func StartContainer(ctx context.Context, cli DockerClient, cfg ServerConfig) (string, error) {
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
		Cmd:   []string{"server", "--bind", "0.0.0.0"},
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

func AttachContainer(ctx context.Context, cli DockerClient, containerID string) error {
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

func StopAndRemoveContainer(cli DockerClient, containerID string) {
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

// createTarball creates a tar archive of the given directory
func CreateTarball(dir string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := filepath.Walk(dir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if fi.IsDir() {
			return nil
		}

		// Open the file
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		// Write file header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}
		header.Name = file // Preserve file path
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Copy file data to the tar writer
		_, err = io.Copy(tw, f)
		return err
	})

	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}

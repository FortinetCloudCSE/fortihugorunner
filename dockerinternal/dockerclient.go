package dockerinternal
import (
    "context"
    "io"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/image"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/network"
    ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type DockerClient interface {
	ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)

	ImageBuild(ctx context.Context, context io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)

	BuildCachePrune(ctx context.Context, opts types.BuildCachePruneOptions) (*types.BuildCachePruneReport, error)

	ContainerCreate(
		ctx context.Context,
		config *container.Config,
		hostConfig *container.HostConfig,
		networkConfig *network.NetworkingConfig,
		platform *ocispec.Platform,
		containerName string,
	) (container.CreateResponse, error)

	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error

	ContainerAttach(
		ctx context.Context,
		containerID string,
		options container.AttachOptions,
	) (types.HijackedResponse, error)

	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error

	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
}


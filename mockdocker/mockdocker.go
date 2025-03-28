package mockdocker

import (
    "context"
    "errors"
    "io"
    "strings"
    "bufio"
    "bytes"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/image"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/network"
    ocispec "github.com/opencontainers/image-spec/specs-go/v1"
   
    //"docker-run-go/dockerinternal"
)

type MockDockerClient struct {
    FailPull         bool
    FailBuild        bool
    FailPrune        bool
    FailCreate       bool
    FailStart        bool
    FailAttach       bool
    FailStop         bool
    FailRemove       bool
}

func (m *MockDockerClient) ImagePull(ctx context.Context, refStr string, opts image.PullOptions) (io.ReadCloser, error) {
    if m.FailPull {
        return nil, errors.New("mock: failed to pull image")
    }
    return io.NopCloser(strings.NewReader("mock pull output")), nil
}

func (m *MockDockerClient) ImageBuild(ctx context.Context, context io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
    if m.FailBuild {
        return types.ImageBuildResponse{}, errors.New("mock: failed to build image")
    }
    return types.ImageBuildResponse{Body: io.NopCloser(strings.NewReader("mock build output"))}, nil
}

func (m *MockDockerClient) BuildCachePrune(ctx context.Context, opts types.BuildCachePruneOptions) (*types.BuildCachePruneReport, error) {
    if m.FailPrune {
        return nil, errors.New("mock: failed to prune build cache")
    }
    return &types.BuildCachePruneReport{}, nil
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkConfig *network.NetworkingConfig, platform *ocispec.Platform, name string) (container.CreateResponse, error) {
    if m.FailCreate {
        return container.CreateResponse{}, errors.New("mock: failed to create container")
    }
    return container.CreateResponse{ID: "mock-container-id"}, nil
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, opts container.StartOptions) error {
    if m.FailStart {
        return errors.New("mock: failed to start container")
    }
    return nil
}

func (m *MockDockerClient) ContainerAttach(ctx context.Context, containerID string, options container.AttachOptions) (types.HijackedResponse, error) {
    if m.FailAttach {
        return types.HijackedResponse{}, errors.New("mock: failed to attach to container")
    }
    mockOutput := "mock container output"
    reader := bufio.NewReader(bytes.NewBufferString(mockOutput))
    return types.HijackedResponse{
        Conn:   nil,
        Reader: reader,
    }, nil
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
    if m.FailStop {
        return errors.New("mock: failed to stop container")
    }
    return nil
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
    if m.FailRemove {
        return errors.New("mock: failed to remove container")
    }
    return nil
}

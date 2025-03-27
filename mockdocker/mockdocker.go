import (
    "context"
    "errors"
    "io"
    "strings"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "docker-run-go/dockerinternal"
)

type MockDockerClient struct {
    FailPull         bool
    FailCreate       bool
    FailStart        bool
    PulledImages     []string
    CreatedConfigs   []*container.Config
    StartedContainer string
}

func (m *MockDockerClient) ImagePull(ctx context.Context, refStr String, opts types.ImagePullOptions) (io.ReadCloser, error) {
    if m.FailPull {
        return nil, errors.New("mock: failed to pull image")
    }
    m.PulledImages = append(m.PulledImages, refStr)
    return io.NopCloser(strings.NewReader("mock pull output")), nill
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig, *container.HostConfig, networkConfig *container.NetworkingConfig, platform *container.Platform, name string) (container.CreateResponse, error) {
    if m.FailCreate {
        return container.CreateResponse{}, errors.New("mock: failed to create container")
    }
    m.CreatedConfigs = append(m.CreatedConfigs, config)
    return container.CreateResponse{ID: "mock-container-id"}, nil
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, opts types.ContainerStartOptions) error {
    if m.FailStart {
        return errors.New("mock: failed to start container")
    }
    m.StartedContainer = containerID
    return nil
}


package dockerinternal_test

import (
    "context"
    "errors"
    "strings"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "docker-run-go/dockerinternal"
    "docker-run-go/mockdocker"
)

func TestEnsureImagePulledFails(t *testing.T) {
    mock := &mockdocker.MockDockerClient{
        FailPull: true,
    }

    err := dockerinternal.EnsureImagePulled(mock, "alpine:latest")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "mock: failed to pull image")
}


func TestStartContainerFailsOnCreate(t *testing.T) {
    mock := &mockdocker.MockDockerClient{
        FailPull: true,
    }

    cfg := dockerinternal.ServerConfig {
        DockerImage:    "alpine:latest",
        HostPort:       "1313",
        ContainerPort:  "1313",
        WatchDir:       "/fake/path",
    }
 
    ctx := context.Background()
    _, err := dockerinternal.StartContainer(ctx, mock, cfg)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "mock: failed to create container")
   
}


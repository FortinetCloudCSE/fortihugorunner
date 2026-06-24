package dockerinternal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moby/moby/client"
)

// dockerContextInfo holds the minimal connection data we care about.
type dockerContextInfo struct {
	Host          string
	SkipTLSVerify bool
	TLSDir        string
}

// NewDockerClient centralizes docker client initialization so that we can honor
// Docker contexts in the same way the docker CLI does. We look up the active
// context's endpoint and, when it defines a Host, temporarily set the
// corresponding environment variables before creating the SDK client. This lets
// us reuse the built-in client.FromEnv path so TLS and version handling keep
// working as before.
func NewDockerClient() (*client.Client, error) {
	cleanup, err := applyDockerContextEnv()
	if err != nil {
		return nil, err
	}
	if cleanup != nil {
		defer cleanup()
	}
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// applyDockerContextEnv inspects the current Docker context and, if it defines
// a custom host, sets DOCKER_HOST / DOCKER_CERT_PATH / DOCKER_TLS_VERIFY before
// the SDK client is created. It returns a cleanup function that restores any
// original environment values so the process environment stays unchanged once
// the client has been constructed.
func applyDockerContextEnv() (func(), error) {
	if hostOverrideDisabled() {
		return nil, nil
	}
	info, err := resolveDockerContext()
	if err != nil {
		return nil, err
	}
	if info == nil || info.Host == "" {
		return nil, nil
	}
	updates := map[string]*string{}
	updates["DOCKER_HOST"] = &info.Host

	if info.TLSDir != "" {
		updates["DOCKER_CERT_PATH"] = &info.TLSDir
	}

	if info.SkipTLSVerify {
		updates["DOCKER_TLS_VERIFY"] = nil
	} else if info.TLSDir != "" {
		value := "1"
		updates["DOCKER_TLS_VERIFY"] = &value
	}

	backup := backupEnv(updates)
	for key, value := range updates {
		if value == nil {
			os.Unsetenv(key)
			continue
		}
		os.Setenv(key, *value)
	}

	return func() {
		for key, prior := range backup {
			if prior.present {
				os.Setenv(key, prior.value)
				continue
			}
			os.Unsetenv(key)
		}
	}, nil
}

// hostOverrideDisabled returns true when the caller already provided a
// DOCKER_HOST (unless they also set DOCKER_CONTEXT) so that we do not override a
// deliberate host selection.
func hostOverrideDisabled() bool {
	if os.Getenv("DOCKER_CONTEXT") != "" {
		return false
	}
	return os.Getenv("DOCKER_HOST") != ""
}

type envState struct {
	value   string
	present bool
}

func backupEnv(updates map[string]*string) map[string]envState {
	backup := make(map[string]envState, len(updates))
	for key := range updates {
		value, ok := os.LookupEnv(key)
		backup[key] = envState{value: value, present: ok}
	}
	return backup
}

func resolveDockerContext() (*dockerContextInfo, error) {
	name, err := activeDockerContextName()
	if err != nil {
		return nil, err
	}
	if name == "" || name == "default" {
		return nil, nil
	}
	configDir := dockerConfigDir()
	metaDir := filepath.Join(configDir, "contexts", "meta")
	entries, err := os.ReadDir(metaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read docker contexts: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metaPath := filepath.Join(metaDir, entry.Name(), "meta.json")
		meta, err := readContextMeta(metaPath)
		if err != nil {
			continue
		}
		if meta.Name != name {
			continue
		}
		endpt, ok := meta.Endpoints["docker"]
		if !ok || endpt.Host == "" {
			return nil, fmt.Errorf("docker context %q missing docker endpoint", name)
		}
		info := &dockerContextInfo{
			Host:          endpt.Host,
			SkipTLSVerify: endpt.SkipTLSVerify,
		}
		tlsDir := filepath.Join(configDir, "contexts", "tls", entry.Name(), "docker")
		if _, err := os.Stat(tlsDir); err == nil {
			info.TLSDir = tlsDir
		}
		return info, nil
	}
	return nil, fmt.Errorf("docker context %q not found", name)
}

func activeDockerContextName() (string, error) {
	if ctx := os.Getenv("DOCKER_CONTEXT"); ctx != "" {
		return ctx, nil
	}
	configPath := filepath.Join(dockerConfigDir(), "config.json")
	content, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "default", nil
		}
		return "", fmt.Errorf("failed to read docker config: %w", err)
	}
	var cfg struct {
		CurrentContext string `json:"currentContext"`
	}
	if err := json.Unmarshal(content, &cfg); err != nil {
		return "", fmt.Errorf("failed to parse docker config: %w", err)
	}
	if cfg.CurrentContext == "" {
		return "default", nil
	}
	return cfg.CurrentContext, nil
}

func dockerConfigDir() string {
	if cfg := os.Getenv("DOCKER_CONFIG"); cfg != "" {
		return cfg
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".docker")
	}
	return filepath.Join(home, ".docker")
}

type contextMeta struct {
	Name      string `json:"Name"`
	Endpoints map[string]struct {
		Host          string `json:"Host"`
		SkipTLSVerify bool   `json:"SkipTLSVerify"`
	} `json:"Endpoints"`
}

func readContextMeta(path string) (*contextMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta contextMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

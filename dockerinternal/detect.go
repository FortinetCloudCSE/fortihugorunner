package dockerinternal

import (
	"context"

	"github.com/moby/moby/client"
)

// DockerState is the result of DetectDocker.
type DockerState int

const (
	// Running means a Docker-API-compatible engine answered Ping successfully.
	Running DockerState = iota
	// InstalledNotRunning means a known engine binary/socket was found on the
	// host, but it isn't currently reachable (daemon stopped, VM not started,
	// permissions, etc). Today's existing error + troubleshooting link applies.
	InstalledNotRunning
	// NotInstalled means no known engine was found on the host at all. Only
	// this state triggers the install-offer flow.
	NotInstalled
)

func (s DockerState) String() string {
	switch s {
	case Running:
		return "running"
	case InstalledNotRunning:
		return "installed-not-running"
	case NotInstalled:
		return "not-installed"
	default:
		return "unknown"
	}
}

// DetectDocker distinguishes "no engine found at all" from "an engine is
// present but not reachable right now" from "reachable." Only the first case
// should ever trigger an install offer — the other two keep today's behavior.
func DetectDocker() DockerState {
	cli, err := NewDockerClient()
	if err == nil {
		if _, pingErr := cli.Ping(context.Background(), client.PingOptions{}); pingErr == nil {
			return Running
		}
	}

	// The client/ping path failed. Before concluding nothing is installed,
	// check for a known engine on this platform (binary on PATH, or a
	// platform-specific marker such as Colima's config dir).
	if platformDockerPresent() {
		return InstalledNotRunning
	}
	return NotInstalled
}

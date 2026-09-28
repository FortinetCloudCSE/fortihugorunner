//go:build linux

package dockerinternal

import "os/exec"

// platformDockerPresent reports whether a known Docker-compatible engine
// binary is present on Linux, even if it isn't currently reachable (daemon
// stopped, permissions, etc). Only used to distinguish "not installed" from
// "installed but not running" for the install-offer flow.
func platformDockerPresent() bool {
	for _, bin := range []string{"docker", "podman"} {
		if _, err := exec.LookPath(bin); err == nil {
			return true
		}
	}
	return false
}

//go:build windows

package dockerinternal

import "os/exec"

// platformDockerPresent reports whether a known Docker-compatible engine is
// present on Windows: Docker Desktop/CLI on PATH, or Docker Engine already
// installed inside the default WSL2 distro, even if not currently running.
// Only used to distinguish "not installed" from "installed but not running"
// for the install-offer flow.
func platformDockerPresent() bool {
	if _, err := exec.LookPath("docker.exe"); err == nil {
		return true
	}
	if _, err := exec.LookPath("docker"); err == nil {
		return true
	}
	// Check inside the default WSL2 distro too, since our install path puts
	// Docker Engine there rather than on the Windows side.
	out, err := exec.Command("wsl.exe", "-e", "sh", "-lc", "command -v docker").Output()
	return err == nil && len(out) > 0
}

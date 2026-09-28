//go:build darwin

package dockerinternal

import (
	"os"
	"os/exec"
	"path/filepath"
)

// platformDockerPresent reports whether a known Docker-compatible engine is
// present on macOS: Docker Desktop, Colima, or Rancher Desktop, even if not
// currently running. Only used to distinguish "not installed" from "installed
// but not running" for the install-offer flow.
func platformDockerPresent() bool {
	for _, bin := range []string{"docker", "colima", "rdctl"} {
		if _, err := exec.LookPath(bin); err == nil {
			return true
		}
	}
	if _, err := os.Stat("/Applications/Docker.app"); err == nil {
		return true
	}
	home, err := os.UserHomeDir()
	if err == nil {
		if _, err := os.Stat(filepath.Join(home, ".colima")); err == nil {
			return true
		}
	}
	return false
}

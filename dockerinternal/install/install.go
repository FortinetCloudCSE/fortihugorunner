// Package install provides per-OS installation of a free, lightweight,
// Docker-API-compatible engine for hosts where fortihugorunner finds none.
// See docs/plans/0004_2026-09-28_Jeff-Kopko_auto-install-docker.md.
//
// Every Install here executes real, privileged, system-mutating commands.
// Callers are responsible for getting explicit user consent (a confirmation
// prompt, --yes, or an env var opt-in) before calling with DryRun: false —
// this package never prompts itself, and never installs Docker Desktop
// (its free-use terms exclude larger organizations; see the plan's Risks).
package install

import (
	"fmt"
	"io"
	"os"
)

// Pinned to a specific commit of docker/docker-install so the install
// behavior can't change out from under us on a future run — see this repo's
// own "keep the Hugo base image pinned" gotcha for the same reasoning. Used
// on native Linux (linux.go) and inside WSL2 on Windows (windows.go), which
// is why it lives here rather than in either OS-specific file. Re-pin
// deliberately: fetch the new commit's raw install.sh, recompute the
// checksum below, and update both together.
const (
	dockerInstallScriptURL    = "https://raw.githubusercontent.com/docker/docker-install/2b32480025b223ebfddae9a3a8bef09027680f53/install.sh"
	dockerInstallScriptSHA256 = "fefa50ccd50efb42f438b506fc3a88574118f314aaf2a7cd5b6e1ffb1bffcf26"
)

// Options controls how Install behaves.
type Options struct {
	// DryRun prints exactly what would run, without executing anything.
	DryRun bool
	// Out receives progress output. Defaults to os.Stdout when nil.
	Out io.Writer
}

func (o Options) out() io.Writer {
	if o.Out != nil {
		return o.Out
	}
	return os.Stdout
}

func (o Options) logf(format string, args ...any) {
	fmt.Fprintf(o.out(), format, args...)
}

//go:build darwin

package install

import (
	"context"
	"fmt"
	"os/exec"
)

// Homebrew's own installer is itself a curl|bash script maintained by the
// Homebrew project; we invoke it via `brew` if already present, and only fall
// back to installing Homebrew itself (pinned to a released tag, not HEAD) when
// it's missing — a genuinely separate, also-consented step.
const (
	homebrewInstallScriptURL = "https://raw.githubusercontent.com/Homebrew/install/master/install.sh"
)

// Install installs Colima (+ the docker CLI) via Homebrew on macOS, deliberately
// not Docker Desktop — see package doc. Requires Homebrew; installs it first if
// missing. UNVERIFIED end-to-end: no macOS host was available to run this
// against as of docs/plans/0004; see that plan's Risks section.
func Install(ctx context.Context, opts Options) error {
	if _, err := exec.LookPath("brew"); err != nil {
		if err := installHomebrew(ctx, opts); err != nil {
			return fmt.Errorf("installing Homebrew (required for Colima): %w", err)
		}
	}

	if err := runStep(ctx, opts, "", "brew", "install", "colima", "docker"); err != nil {
		return fmt.Errorf("brew install colima docker: %w", err)
	}

	if err := runStep(ctx, opts, "", "colima", "start"); err != nil {
		return fmt.Errorf("colima start: %w", err)
	}

	// Colima registers itself as a Docker CLI context; make sure it's active
	// so NewDockerClient's context resolution (dockerinternal/docker_client.go)
	// picks it up without the caller needing to export DOCKER_HOST.
	if err := runStep(ctx, opts, "", "docker", "context", "use", "colima"); err != nil {
		return fmt.Errorf("docker context use colima: %w", err)
	}

	return nil
}

func installHomebrew(ctx context.Context, opts Options) error {
	if opts.DryRun {
		opts.logf("[dry-run] would download and run Homebrew's installer: %s\n", homebrewInstallScriptURL)
		return nil
	}
	opts.logf("Homebrew not found — installing it first (required for Colima)...\n")
	// Homebrew's documented one-liner; NONINTERACTIVE avoids it blocking on a
	// prompt we can't answer for the user in an already-consented flow.
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c",
		fmt.Sprintf(`NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL %s)"`, homebrewInstallScriptURL))
	cmd.Stdout = opts.out()
	cmd.Stderr = opts.out()
	return cmd.Run()
}

func runStep(ctx context.Context, opts Options, _ string, name string, args ...string) error {
	if opts.DryRun {
		opts.logf("[dry-run] would run: %s %s\n", name, joinArgs(args))
		return nil
	}
	opts.logf("Running: %s %s\n", name, joinArgs(args))
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = opts.out()
	cmd.Stderr = opts.out()
	return cmd.Run()
}

func joinArgs(args []string) string {
	out := ""
	for i, a := range args {
		if i > 0 {
			out += " "
		}
		out += a
	}
	return out
}

//go:build windows

package install

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Install installs Docker Engine CE inside the default WSL2 distro (not
// Docker Desktop — see package doc), using only its standard Unix socket.
//
// Deliberately does NOT expose the daemon over TCP to bridge it to the
// native Windows process: that would mean either an unauthenticated
// loopback-TCP Docker API (root-equivalent access to anyone who can reach
// it) or standing up SSH-based auth, both of which are more attack surface
// than this tool should introduce on someone's behalf. Instead, this prints
// clear instructions to run fortihugorunner from inside WSL2 itself, where
// the Unix socket is directly reachable — the standard pattern for
// Docker-without-Desktop on Windows.
func Install(ctx context.Context, opts Options) error {
	if err := requireWSL2(ctx, opts); err != nil {
		return err
	}

	// Reuses the exact same pinned, checksum-verified script this package
	// uses on native Linux — one source of truth for "how we install Docker
	// Engine CE," just invoked inside the distro via `wsl.exe -e`.
	inner := fmt.Sprintf(
		`set -e
curl -fsSL %s -o /tmp/get-docker.sh
echo '%s  /tmp/get-docker.sh' | sha256sum -c -
sh /tmp/get-docker.sh
sudo service docker start || (sudo dockerd > /tmp/dockerd.log 2>&1 &)
`,
		dockerInstallScriptURL, dockerInstallScriptSHA256,
	)

	if opts.DryRun {
		opts.logf("[dry-run] would run inside the default WSL2 distro:\n%s\n", inner)
		return nil
	}

	opts.logf("Installing Docker Engine CE inside the default WSL2 distro...\n")
	cmd := exec.CommandContext(ctx, "wsl.exe", "-e", "bash", "-lc", inner)
	cmd.Stdout = opts.out()
	cmd.Stderr = opts.out()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installing Docker Engine CE inside WSL2: %w", err)
	}

	opts.logf("\nDocker Engine CE is installed and running inside your default WSL2 distro.\n" +
		"fortihugorunner.exe (native Windows) cannot safely reach a WSL2-only Docker socket — " +
		"bridging that would mean exposing the Docker API over the network or SSH, which this " +
		"tool won't set up on your behalf.\n\n" +
		"Run fortihugorunner from inside WSL2 instead, where the socket is directly reachable:\n" +
		"  wsl.exe\n" +
		"  # then, inside WSL2: install Go and this tool, or use a Linux build of fortihugorunner\n")

	return nil
}

// requireWSL2 checks that WSL2 (with at least one installed distro) is
// available. Enabling WSL2 itself can require a Windows feature toggle and a
// reboot — deliberately out of scope for an unattended install (see the
// plan's Windows row) — so this fails with manual instructions instead of
// attempting it.
func requireWSL2(ctx context.Context, opts Options) error {
	out, err := exec.CommandContext(ctx, "wsl.exe", "-l", "-v").CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"WSL2 does not appear to be available (%w). Enable it first — this requires a "+
				"restart and is not something fortihugorunner will do automatically:\n"+
				"  wsl --install\n"+
				"then re-run this command. See https://learn.microsoft.com/windows/wsl/install", err,
		)
	}
	// `wsl -l -v` output uses UTF-16LE with a BOM on stock Windows terminals;
	// exec.Command's CombinedOutput gives us raw bytes either way, and a
	// substring check on the decoded-as-UTF8 fallback is good enough here —
	// we only need to know a distro exists, not parse the table precisely.
	if !strings.Contains(string(out), "2") {
		return fmt.Errorf(
			"WSL is installed but no WSL2 distro was found. Install one first:\n" +
				"  wsl --install -d Ubuntu\n" +
				"then re-run this command",
		)
	}
	return nil
}

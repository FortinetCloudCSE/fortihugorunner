//go:build linux

package install

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"
)

// Install installs Docker Engine CE on Linux via Docker's own pinned,
// checksum-verified install script, then starts the daemon. Requires root or
// a working sudo. See package doc — callers must get consent first.
func Install(ctx context.Context, opts Options) error {
	scriptPath, err := fetchInstallScript(ctx, opts)
	if err != nil {
		return err
	}
	if !opts.DryRun {
		defer os.Remove(scriptPath)
	}

	runAsRoot := os.Geteuid() != 0
	sudo := ""
	if runAsRoot {
		if _, err := exec.LookPath("sudo"); err != nil {
			return fmt.Errorf("not running as root and sudo is not available — re-run as root, or install Docker manually: https://docs.docker.com/engine/install/")
		}
		sudo = "sudo"
	}

	if err := runStep(ctx, opts, sudo, "sh", scriptPath); err != nil {
		return fmt.Errorf("docker install script failed: %w", err)
	}

	if err := startDaemon(ctx, opts, sudo); err != nil {
		return fmt.Errorf("docker install succeeded but starting the daemon failed: %w", err)
	}

	if err := addUserToDockerGroup(ctx, opts, sudo); err != nil {
		// Non-fatal: sudo/root still works without group membership.
		opts.logf("Note: could not add your user to the docker group (%v). "+
			"You can still use fortihugorunner via sudo, or run "+
			"'sudo usermod -aG docker $USER' and log out/in.\n", err)
	}

	return waitForDaemon(ctx, opts)
}

func fetchInstallScript(ctx context.Context, opts Options) (string, error) {
	if opts.DryRun {
		opts.logf("[dry-run] would download %s\n", dockerInstallScriptURL)
		opts.logf("[dry-run] would verify sha256 == %s\n", dockerInstallScriptSHA256)
		return "<downloaded-script>", nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dockerInstallScriptURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("downloading install script: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading install script: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading install script: %w", err)
	}

	sum := sha256.Sum256(body)
	got := hex.EncodeToString(sum[:])
	if got != dockerInstallScriptSHA256 {
		return "", fmt.Errorf(
			"install script checksum mismatch (got %s, expected %s) — refusing to run an "+
				"unverified script; the pinned commit may have been force-pushed, re-pin deliberately",
			got, dockerInstallScriptSHA256,
		)
	}

	f, err := os.CreateTemp("", "fortihugorunner-get-docker-*.sh")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(body); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	if err := os.Chmod(f.Name(), 0o500); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func startDaemon(ctx context.Context, opts Options, sudo string) error {
	if hasSystemd() {
		if err := runStep(ctx, opts, sudo, "systemctl", "enable", "--now", "docker"); err == nil {
			return nil
		}
		// fall through to the direct-dockerd fallback below — some minimal
		// containers report systemd present but not actually running as pid 1.
	}

	// No systemd (or it didn't work): start dockerd directly, backgrounded,
	// logging to a file so the caller isn't blocked and output isn't lost.
	logPath := filepath.Join(os.TempDir(), "fortihugorunner-dockerd.log")
	if opts.DryRun {
		opts.logf("[dry-run] would run: %sdockerd > %s 2>&1 &\n", sudoPrefix(sudo), logPath)
		return nil
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	// Built explicitly (not via runStep) since this needs to be backgrounded
	// (Start, not Run) and outlive the parent context/process.
	var cmd *exec.Cmd
	if sudo != "" {
		cmd = exec.CommandContext(context.WithoutCancel(ctx), sudo, "dockerd")
	} else {
		cmd = exec.CommandContext(context.WithoutCancel(ctx), "dockerd")
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("starting dockerd directly (no systemd available): %w", err)
	}
	opts.logf("Started dockerd directly (pid %d, no systemd in this environment); logs: %s\n", cmd.Process.Pid, logPath)
	return nil
}

func hasSystemd() bool {
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}
	return false
}

func addUserToDockerGroup(ctx context.Context, opts Options, sudo string) error {
	if sudo == "" {
		// Already root; group membership is moot.
		return nil
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	return runStep(ctx, opts, sudo, "usermod", "-aG", "docker", u.Username)
}

func waitForDaemon(ctx context.Context, opts Options) error {
	if opts.DryRun {
		return nil
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat("/var/run/docker.sock"); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("docker.sock did not appear within 30s after install — check %s", filepath.Join(os.TempDir(), "fortihugorunner-dockerd.log"))
}

func sudoPrefix(sudo string) string {
	if sudo == "" {
		return ""
	}
	return sudo + " "
}

func runStep(ctx context.Context, opts Options, sudo string, name string, args ...string) error {
	full := append([]string{}, args...)
	displayName := name
	if sudo != "" {
		full = append([]string{name}, args...)
		displayName = sudo + " " + name
	}
	if opts.DryRun {
		opts.logf("[dry-run] would run: %s %s\n", displayName, joinArgs(args))
		return nil
	}
	opts.logf("Running: %s %s\n", displayName, joinArgs(args))

	var cmd *exec.Cmd
	if sudo != "" {
		cmd = exec.CommandContext(ctx, sudo, full...)
	} else {
		cmd = exec.CommandContext(ctx, name, args...)
	}
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

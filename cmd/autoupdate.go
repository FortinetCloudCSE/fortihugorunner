package cmd

import (
	"fmt"
	"fortihugorunner/version"
	"github.com/blang/semver"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var noAutoUpdate bool

// autoUpdateCheckInterval rate-limits the GitHub API check so a normal
// workflow (many invocations per session) doesn't hit it every time.
const autoUpdateCheckInterval = 24 * time.Hour

// maybeAutoUpdate checks at most once per autoUpdateCheckInterval whether a
// newer fortihugorunner release exists, and offers to install it. This must
// never block or fail the caller's actual command: any error (cache I/O,
// network, GitHub API, an unparsable dev version) is swallowed after at most
// a one-line notice, and the command proceeds on the current binary. Skips
// entirely for `update` and `version` (avoid recursion/interference with
// their own output) and when disabled via --no-auto-update or
// FORTIHUGORUNNER_NO_AUTO_UPDATE=1.
func maybeAutoUpdate(cmdName string) {
	if cmdName == "update" || cmdName == "version" {
		return
	}
	if noAutoUpdate || os.Getenv("FORTIHUGORUNNER_NO_AUTO_UPDATE") == "1" {
		return
	}
	if !autoUpdateDue() {
		return
	}

	v, err := semver.ParseTolerant(version.Version)
	if err != nil {
		// Dev/local builds etc. — nothing sane to compare against.
		recordAutoUpdateChecked()
		return
	}

	rel, err := detectLatestRelease(v)
	// Record the check happened regardless of outcome (found/not
	// found/error) so a flaky network doesn't cause a check on every single
	// invocation — the next real attempt waits for the next window either way.
	recordAutoUpdateChecked()
	if err != nil {
		// Fail open: a GitHub API hiccup or rate limit must never block the
		// user's actual command.
		return
	}
	if rel == nil {
		return // already latest
	}

	yes := os.Getenv("FORTIHUGORUNNER_AUTO_UPDATE") == "1"
	fmt.Printf("\nA newer version of fortihugorunner is available: %s (current: %s)\n", rel.Version, version.Version)
	if !yes {
		fmt.Print("Update now? [y/N] ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" && answer != "yes" {
			return
		}
	}

	if err := applyUpdate(rel); err != nil {
		fmt.Printf("Auto-update failed: %v\n", err)
		return
	}
	fmt.Printf("Updated to %s — continuing...\n", rel.Version)
	reExecSelf()
}

// reExecSelf re-runs the (now-updated) executable with the exact args the
// user originally typed, so their original command completes transparently
// on the new version, then exits this process. If re-exec can't even be
// attempted, falls through silently and the caller continues on whatever is
// currently loaded in memory — still functional, just not the new binary
// until next run.
func reExecSelf() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	c := exec.Command(exePath, os.Args[1:]...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	if err := c.Start(); err != nil {
		return
	}
	c.Wait()
	os.Exit(0)
}

func autoUpdateCacheFile() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "fortihugorunner", "last-update-check"), nil
}

// autoUpdateDue reports whether enough time has passed since the last check
// (or there was never one, or the cache can't be read) to check again. Any
// failure to read the cache defaults to "due" — the worst case is one extra
// network check, not a broken command.
func autoUpdateDue() bool {
	path, err := autoUpdateCacheFile()
	if err != nil {
		return true
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return true
	}
	last, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return true
	}
	return time.Since(last) >= autoUpdateCheckInterval
}

// recordAutoUpdateChecked best-effort persists "we just checked" so the next
// autoUpdateDue call rate-limits correctly. A failure here (unwritable cache
// dir, read-only filesystem, etc.) is silently ignored — worst case is
// checking again next run, never a broken command.
func recordAutoUpdateChecked() {
	path, err := autoUpdateCacheFile()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(path, []byte(time.Now().UTC().Format(time.RFC3339)), 0o644)
}

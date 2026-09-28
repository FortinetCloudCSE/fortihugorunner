package cmd

import (
	"fmt"
	"fortihugorunner/utilities"
	"fortihugorunner/version"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const repoSlug = "FortinetCloudCSE/fortihugorunner"

// detectLatestRelease returns the latest GitHub release for repoSlug if it's
// newer than current, or nil if current is already the latest (or no release
// was found at all). Shared by the explicit `update` command and the
// automatic update check in cmd/autoupdate.go so there's one source of truth
// for "is an update available."
func detectLatestRelease(current semver.Version) (*selfupdate.Release, error) {
	updater, err := selfupdate.NewUpdater(selfupdate.Config{})
	if err != nil {
		return nil, err
	}
	rel, ok, err := updater.DetectLatest(repoSlug)
	if err != nil {
		return nil, err
	}
	if !ok || rel.Version.Equals(current) {
		return nil, nil
	}
	return rel, nil
}

// applyUpdate downloads and installs rel over the currently running
// executable, in place. Shared by the explicit `update` command and the
// automatic update check.
func applyUpdate(rel *selfupdate.Release) error {
	cmdPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %w", err)
	}
	updater, err := selfupdate.NewUpdater(selfupdate.Config{})
	if err != nil {
		return err
	}
	return updater.UpdateTo(rel, cmdPath)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update fortihugorunner to the latest version.",
	RunE: func(cmd *cobra.Command, args []string) error {

		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("could not get executable path: %w", err)
		}
		dir := filepath.Dir(exePath)
		expectedName := "fortihugorunner"
		if runtime.GOOS == "windows" {
			expectedName += ".exe"
		}
		expectedPath := filepath.Join(dir, expectedName)

		if !strings.EqualFold(filepath.Base(exePath), expectedName) {

			fmt.Println("Renaming the executable...")
			// sleep to deal with a Windows file-locking mechanism
			time.Sleep(500 * time.Millisecond)
			err = utilities.RenameBinary(exePath)
			if err != nil {
				return fmt.Errorf("error renaming binary: %w", err)
			}
			cmd := exec.Command(expectedPath, os.Args[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Start()
			cmd.Wait()
			os.Exit(0)
		}

		v, err := semver.ParseTolerant(version.Version)
		if err != nil {
			return fmt.Errorf("Erroring parsing version: %w", err)
		}

		rel, err := detectLatestRelease(v)
		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		if rel == nil {
			fmt.Fprintf(os.Stdout, "You're already running the latest version (%s)\n", version.Version)
			os.Stdout.Sync()
			os.Exit(0)
		}

		if err := applyUpdate(rel); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		fmt.Fprintf(os.Stdout, "Successfully updated to version %s!\n", rel.Version)
		os.Stdout.Sync()
		os.Exit(0)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

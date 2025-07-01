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
		updater, err := selfupdate.NewUpdater(selfupdate.Config{})
		latest, err := updater.UpdateSelf(v, repoSlug)

		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		if latest.Version.Equals(v) {
			fmt.Fprintf(os.Stdout, "You're already running the latest version (%s)\n", version.Version)
			os.Stdout.Sync()
			os.Exit(0)
		} else {
			fmt.Fprintf(os.Stdout, "Successfully updated to version %s!\n", latest.Version)
			os.Stdout.Sync()
			os.Exit(0)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

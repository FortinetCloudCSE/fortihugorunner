package cmd

import (
	"fmt"
	"fortihugorunner/version"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

const repoSlug = "FortinetCloudCSE/fortihugorunner"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update fortihugorunner to the latest version.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Checking for updates...")

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
			fmt.Printf("You're already running the latest version (%s)\n", version.Version)
		} else {
			fmt.Printf("Successfully updated to version %s!\n", latest.Version)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

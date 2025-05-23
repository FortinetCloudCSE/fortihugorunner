package cmd

import (
	"fmt"
	"os"

	"docker-run-go/pkg/upgraderepo"
	"github.com/spf13/cobra"
)

var specPath string

var upgradeRepoCmd = &cobra.Command{
	Use:   "upgrade-repo",
	Short: "Stage updates to the repo based on latest configuration in UserRepo.",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := upgraderepo.RunUpgrade()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Upgrade failed: %v\n", err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeRepoCmd)
}

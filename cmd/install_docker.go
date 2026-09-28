package cmd

import (
	"context"
	"fmt"
	"os"

	"fortihugorunner/dockerinternal"
	"fortihugorunner/dockerinternal/install"
	"github.com/spf13/cobra"
)

var installDockerCmd = &cobra.Command{
	Use:   "install-docker",
	Short: "Install a lightweight, free Docker-compatible engine for this OS",
	Long: `Installs a free, lightweight, Docker-API-compatible engine when none is found:
Docker Engine CE on Linux, Colima on macOS, and Docker Engine CE inside WSL2 on
Windows. Deliberately never installs Docker Desktop — see
docs/plans/0004_2026-09-28_Jeff-Kopko_auto-install-docker.md for why.

This runs real, privileged, system-mutating commands. Use --dry-run first to
see exactly what would run.

Example:
  fortihugorunner install-docker --dry-run
  fortihugorunner install-docker --yes
`,
	Run: func(cmd *cobra.Command, args []string) {
		yes := getFlagBool(cmd, "yes") || os.Getenv("FORTIHUGORUNNER_AUTO_INSTALL_DOCKER") == "1"
		dryRun := getFlagBool(cmd, "dry-run")

		state := dockerinternal.DetectDocker()
		switch state {
		case dockerinternal.Running:
			fmt.Println("Docker is already installed and reachable — nothing to do.")
			return
		case dockerinternal.InstalledNotRunning:
			fmt.Println("A Docker-compatible engine is already installed but not reachable right now.")
			fmt.Println("This command only installs a missing engine; troubleshoot the existing one: https://docs.docker.com/engine/daemon/troubleshoot/")
			os.Exit(1)
		}

		if !dryRun && !yes && !confirmInstall() {
			fmt.Println("Aborted — nothing installed.")
			os.Exit(1)
		}

		opts := install.Options{DryRun: dryRun}
		if err := install.Install(context.Background(), opts); err != nil {
			fmt.Printf("Error installing Docker: %v\n", err)
			os.Exit(1)
		}
		if !dryRun {
			fmt.Println("Docker installed successfully.")
		}
	},
}

func confirmInstall() bool {
	fmt.Print("No Docker-compatible engine was found. Install one now? [y/N] ")
	var answer string
	fmt.Scanln(&answer)
	return answer == "y" || answer == "Y" || answer == "yes"
}

func init() {
	rootCmd.AddCommand(installDockerCmd)
	installDockerCmd.Flags().Bool("yes", false, "Skip the confirmation prompt (for CI / scripted use). Also settable via FORTIHUGORUNNER_AUTO_INSTALL_DOCKER=1.")
	installDockerCmd.Flags().Bool("dry-run", false, "Print exactly what would run, without installing anything.")
}

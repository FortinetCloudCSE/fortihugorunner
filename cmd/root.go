package cmd

import (
	"context"
	"fmt"
	"fortihugorunner/dockerinternal"
	"fortihugorunner/dockerinternal/install"
	"fortihugorunner/version"
	"github.com/moby/moby/client"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

var rootVersion bool
var noInstallDocker bool

var rootCmd = &cobra.Command{
	Use:   "fortihugorunner",
	Short: "FortinetCloudCSE Workshop Docker development utility.",
	Long:  "Includes functions for facilitating Hugo app development with docker containers.",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// install-docker handles its own detection/offer flow; don't
		// double-prompt or block it from running when Docker is absent.
		if cmd.Name() == "install-docker" {
			return nil
		}

		err := checkDockerRunning()
		if err == nil {
			return nil
		}

		if !noInstallDocker && dockerinternal.DetectDocker() == dockerinternal.NotInstalled {
			if offerAndInstallDocker() {
				if retryErr := checkDockerRunning(); retryErr == nil {
					return nil
				}
			}
		}

		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		fmt.Fprintf(os.Stderr, "\nReceived the error below. For troubleshooting help, head here: https://docs.docker.com/engine/daemon/troubleshoot/\n\n")
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		if rootVersion {
			osType := runtime.GOOS
			arch := runtime.GOARCH
			platform := osType + "/" + arch
			fmt.Printf("Version: %s\nDate: %s\nPlatform: %s\n", version.Version, version.Date, platform)
			os.Exit(0)
		}
		cmd.Help()
	},
}

// offerAndInstallDocker prompts (unless FORTIHUGORUNNER_AUTO_INSTALL_DOCKER=1
// is set) before installing anything — see install.Install's package doc.
// Returns true only if an install was attempted and reported success.
func offerAndInstallDocker() bool {
	yes := os.Getenv("FORTIHUGORUNNER_AUTO_INSTALL_DOCKER") == "1"
	fmt.Println("\nNo Docker-compatible engine was found on this system.")
	if !yes {
		fmt.Print("Install a free, lightweight one now (see 'fortihugorunner install-docker --dry-run' for exactly what that runs)? [y/N] ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" && answer != "yes" {
			return false
		}
	}
	if err := install.Install(context.Background(), install.Options{}); err != nil {
		fmt.Printf("Docker install failed: %v\n", err)
		return false
	}
	return true
}

func checkDockerRunning() error {
	ctx := context.Background()
	cli, err := dockerinternal.NewDockerClient()
	if err != nil {
		return fmt.Errorf("could not create Docker client: %w", err)
	}

	_, err = cli.Ping(ctx, client.PingOptions{})
	if err != nil {
		return err
	}
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootVersion, "version", "v", false, "fortihugorunner version information")
	rootCmd.PersistentFlags().BoolVar(&noInstallDocker, "no-install-docker", false, "Never offer to install Docker automatically when it's missing; keep today's plain error.")
}

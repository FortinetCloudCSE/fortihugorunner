package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	//"strings"

	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename to fortihugorunner; i.e. trim OS/arch suffix (e.g. fortihugorunner-linux-amd64 → fortihugorunner)",
	RunE: func(cmd *cobra.Command, args []string) error {
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("could not get executable path: %w", err)
		}
		base := filepath.Base(exePath)
		dir := filepath.Dir(exePath)

		// Regex: match myapp-<os>-<arch> or myapp-<os>-<arch>.exe
		re := regexp.MustCompile(`^(.*?)-([^-]+)-([^-]+)(\.exe)?$`)
		matches := re.FindStringSubmatch(base)
		if len(matches) >= 4 {
			newName := matches[1]
			// On Windows, keep .exe
			if runtime.GOOS == "windows" {
				newName += ".exe"
			}
			newPath := filepath.Join(dir, newName)
			// Prevent overwrite
			if _, err := os.Stat(newPath); err == nil {
				return fmt.Errorf("target file %q already exists", newPath)
			}
			// Rename binary
			if err := os.Rename(exePath, newPath); err != nil {
				return fmt.Errorf("could not rename binary: %w", err)
			}
			fmt.Printf("Renamed %s to %s\n", base, newName)
		} else {
			fmt.Println("Binary name does not match pattern; no rename performed.")
		}
		return nil
	},
}

func init() {
        rootCmd.AddCommand(renameCmd)
}

package utilities

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

func RenameBinary(exePath string) error {

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
}

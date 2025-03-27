package dockerinternal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	//"github.com/docker/docker/client"
	"github.com/fsnotify/fsnotify"
)

// adjustPathForDockerWithOS converts paths for Windows/WSL2; on macOS (darwin) no change is needed.
func AdjustPathForDockerWithOS(path, goos string, isWSL bool) string {
	if goos == "darwin" {
		return path
	} else if goos == "windows" {
		if strings.HasPrefix(path, "/mnt/") {
			path = strings.ReplaceAll(path, "/mnt/", "")
			path = strings.ReplaceAll(path, "/", "\\")
			path = strings.ToUpper(path[:1]) + ":" + path[1:]
		}
	} else if isWSL {
		if len(path) > 1 && path[1] == ':' {
			drive := strings.ToLower(string(path[0]))
			path = fmt.Sprintf("/mnt/%s%s", drive, strings.ReplaceAll(path[2:], "\\", "/"))
		}
	}
	return path
}

func AdjustPathForDocker(path string) string {
	return AdjustPathForDockerWithOS(path, runtime.GOOS, IsWSL2())
}

func IsWSL2() bool {
	_, isWSL := os.LookupEnv("WSL_INTEROP")
	return isWSL && runtime.GOOS == "linux"
}

func WatchAndRestart(ctx context.Context, cli DockerClient, cfg ServerConfig, containerID *string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error creating file watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	// Add watchers recursively for subdirectories.
	filepath.Walk(cfg.WatchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := watcher.Add(path); err != nil {
				fmt.Printf("Error watching directory %s: %v\n", path, err)
			}
		}
		return nil
	})

	fmt.Println("Watching for file changes in:", cfg.WatchDir)
	debounceDuration := 2 * time.Second
	var debounceTimer *time.Timer

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
				fmt.Println("File change detected:", event.Name)
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.NewTimer(debounceDuration)
			}
		case <-func() <-chan time.Time {
			if debounceTimer != nil {
				return debounceTimer.C
			}
			ch := make(chan time.Time)
			return ch
		}():
			fmt.Println("Restarting container due to file changes")
			StopAndRemoveContainer(cli, *containerID)
			newID, err := StartContainer(ctx, cli, cfg)
			if err != nil {
				fmt.Printf("Error restarting container: %v\n", err)
			} else {
				*containerID = newID
				if err := AttachContainer(ctx, cli, newID); err != nil {
					fmt.Printf("Error attaching to container: %v\n", err)
				}
			}
			debounceTimer = nil
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)
		case <-ctx.Done():
			return
		}
	}
}

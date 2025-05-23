package upgraderepo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Current upgrade spec

const specURL = "https://raw.githubusercontent.com/FortinetCloudCSE/UserRepo/main/repo_upgrade_spec.json"

// ---- Data structures ----

type FileMapping struct {
	Remote string `json:"remote"`
	Local  string `json:"local"`
}

// To unmarshal: []map[string][]FileMapping
type UpgradeSpec struct {
	Version         string                     `json:"version"`
	FilesToCopy     []map[string][]FileMapping `json:"files_to_copy"`
	FilesToDelete   []string                   `json:"files_to_delete"`
	FoldersToDelete []string                   `json:"folders_to_delete"`
}

const VersionFile = ".repo_upgrade_version"

// ---- Map repo keys to raw URL base ----
var repoURLMap = map[string]string{
	"UserRepo": "https://raw.githubusercontent.com/FortinetCloudCSE/UserRepo/main/",
	"CentralRepo":    "https://raw.githubusercontent.com/FortinetCloudCSE/CentralRepo/main/",
}

// ---- File Download ----

func fetchAndSaveFile(rawBaseURL, remotePath, localPath string) error {
	fullURL := rawBaseURL + remotePath
	fmt.Printf("Downloading %s -> %s\n", fullURL, localPath)
	resp, err := http.Get(fullURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got status %s for %s", resp.Status, fullURL)
	}
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return err
	}
	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ---- Version file helpers ----
func readRepoVersion() (string, error) {
	bytes, err := os.ReadFile(VersionFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}

func writeRepoVersion(version string) error {
	return os.WriteFile(VersionFile, []byte(version+"\n"), 0644)
}

func DownloadSpecFromURL(url string) (*UpgradeSpec, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download spec: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download spec: %s", resp.Status)
	}
	var spec UpgradeSpec
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&spec); err != nil {
		return nil, fmt.Errorf("failed to parse spec: %w", err)
	}
	return &spec, nil
}

func RunUpgrade() error {
	//Reference current file location at top of file
	spec, err := DownloadSpecFromURL(specURL)
	if err != nil {
		return err
	}
	return runUpgradeWithSpec(spec)
}

// ---- Main upgrade function ----
func runUpgradeWithSpec(spec *UpgradeSpec) error {

	repoVersion, err := readRepoVersion()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not read repo version: %w", err)
	}
	if repoVersion == spec.Version {
		fmt.Println("Repo spec version up to date.")
		return nil
	}
	fmt.Printf("Upgrading repo to version %s (was: %s)\n", spec.Version, repoVersion)

	// 1. Download and stage files to copy
	for _, repoMap := range spec.FilesToCopy {
		for repo, mappings := range repoMap {
			rawBase, ok := repoURLMap[repo]
			if !ok {
				fmt.Fprintf(os.Stderr, "Unknown repo key: %s\n", repo)
				continue
			}
			for _, mapping := range mappings {
				err := fetchAndSaveFile(rawBase, mapping.Remote, mapping.Local)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to fetch %s from %s: %v\n", mapping.Remote, repo, err)
					continue
				}
				// Stage for commit
				if err := runGitCommand("add", mapping.Local); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to git add %s: %v\n", mapping.Local, err)
				}
			}
		}
	}

	// 2. Delete and stage files for removal
	for _, path := range spec.FilesToDelete {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Removing file: %s\n", path)
			if err := os.Remove(path); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to remove file %s: %v\n", path, err)
				continue
			}
			if err := runGitCommand("rm", path); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to git rm %s: %v\n", path, err)
			}
		}
	}

	// 3. Delete and stage folders for removal
	for _, dir := range spec.FoldersToDelete {
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			fmt.Printf("Removing folder: %s\n", dir)
			if err := os.RemoveAll(dir); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to remove folder %s: %v\n", dir, err)
				continue
			}
			if err := runGitCommand("rm", "-r", dir); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to git rm -r %s: %v\n", dir, err)
			}
		}
	}

	// 4. Write new version file
	if err := writeRepoVersion(spec.Version); err != nil {
		return fmt.Errorf("failed to write version file: %w", err)
	}
	fmt.Printf("\nAll changes have been staged. Please review, then run:\n")
	fmt.Println("  git status")
	fmt.Println("  git diff")
	fmt.Println("  git commit -m \"Repo upgrade\"")
	fmt.Println("  git push")
	fmt.Printf("Repo spec version is now %s\n", spec.Version)
	return nil
}

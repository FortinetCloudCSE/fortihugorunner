package dockerinternal_test

import (
	"os"
	"runtime"
	"testing"

	"docker-run-go/dockerinternal"
)

func TestAdjustPathForDocker_Darwin(t *testing.T) {
	//On macOS, no conversion should occur.
	input := "/Users/test/project"
	expected := "/Users/test/project"
	result := dockerinternal.AdjustPathForDockerWithOS(input, "darwin", false)
	if result != expected {
		t.Errorf("Darwin: Expected %s, got %s", expected, result)
	}
}

func TestAdjustPathForDocker_Windows(t *testing.T) {
	//On Windows, convert a Unix-style path to a Windows-style path.
	input := "/mnt/c/Users/test"
	expected := "C:\\Users\\test"
	result := dockerinternal.AdjustPathForDockerWithOS(input, "windows", false)
	if result != expected {
		t.Errorf("Windows: Expected %s, got %s", expected, result)
	}
}

func TestAdjustPathForDocker_WSL2(t *testing.T) {
	//In WSL2 environment, convert a Windows-style path to a WSL2 path.
	input := "C:\\Users\\test"
	expected := "/mnt/c/Users/test"
	result := dockerinternal.AdjustPathForDockerWithOS(input, "linux", true)
	if result != expected {
		t.Errorf("WSL2: Expected %s, got %s", expected, result)
	}
}

func TestIsWSL2(t *testing.T) {
	// Save current value and restore at the end.
	origVal, existed := os.LookupEnv("WSL_INTEROP")
	defer func() {
		if existed {
			os.Setenv("WSL_INTEROP", origVal)
		} else {
			os.Unsetenv("WSL_INTEROP")
		}
	}()

	os.Unsetenv("WSL_INTEROP")
	if dockerinternal.IsWSL2() {
		t.Error("Expected IsWSL2 to be false when WSL_INTEROP is not set")
	}

	os.Setenv("WSL_INTEROP", "dummy")
	// IsWSL2 should return true only if runtime.GOOS is "linux".
	expected := (runtime.GOOS == "linux")
	if dockerinternal.IsWSL2() != expected {
		t.Errorf("Expected IsWSL2 to be %v when WSL_INTEROP is set on OS %s", expected, runtime.GOOS)
	}
}

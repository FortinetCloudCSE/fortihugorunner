package version

// Version and Date are overridden at build time via -ldflags.
// The defaults here are used only for local `go run` / `go build` without ldflags.
var (
	Version = "dev"
	Date    = "unknown"
)

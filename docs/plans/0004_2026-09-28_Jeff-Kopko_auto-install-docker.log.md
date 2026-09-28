# Log: Auto-install a lightweight, free Docker engine when none is found

Plan: docs/plans/0004_2026-09-28_Jeff-Kopko_auto-install-docker.md

## 2026-09-28

- Plan approved. Testing scope: Linux in an isolated throwaway privileged container; Windows
  against a real disposable cloud VM; macOS/Colima code + cross-compile only this pass (real
  Mac hardware testing deferred — see plan header).
- Implementation and verification in progress; entries below as each piece lands.
- Implemented: `dockerinternal.DetectDocker`/`DockerState` + per-OS `platformDockerPresent`
  (`detect_{linux,darwin,windows}.go`); `dockerinternal/install` package with
  `install.Install(ctx, Options{DryRun})` per OS (`linux.go`: Docker Engine CE via a
  pinned+checksummed `get.docker.com` script; `darwin.go`: Colima via Homebrew, installing
  Homebrew first if missing; `windows.go`: same pinned Linux install run inside the default
  WSL2 distro); `cmd/install_docker.go` (new command, `--yes`/`--dry-run`, plus
  `FORTIHUGORUNNER_AUTO_INSTALL_DOCKER=1`); `root.go` wired to offer-and-install only on
  `NotInstalled`, gated by `--no-install-docker`.
- **Design change from the plan's original Windows row, caught by this session's own safety
  classifier**: the first draft bridged the WSL2-only Docker socket out to the native Windows
  binary by having dockerd also listen on `tcp://127.0.0.1:2375` inside WSL2 (relying on
  WSL2's default NAT localhost-forwarding). The classifier flagged this as weakening
  TLS/auth — correctly: an unauthenticated Docker API, even loopback-scoped, is
  root-equivalent access to anyone who can reach it, and mirrored-networking WSL2
  configurations can expose more than plain NAT does. Rewrote `windows.go` to install Docker
  Engine CE inside WSL2 using only its standard Unix socket, and print instructions to run
  `fortihugorunner` from inside WSL2 itself rather than bridging the daemon out — no new
  network-facing attack surface.
- `go build`/`vet`/`test` clean, plus cross-compiled all 6 release targets
  (linux/{amd64,arm64}, darwin/{amd64,arm64}, windows/amd64, and windows/arm64 implied by
  the same build tags) — confirms every OS-specific file at least compiles for its target.
- **Linux verified end-to-end**, in an isolated, throwaway `--privileged` container (never
  against this host's own working Docker setup — a fresh `ubuntu:24.04` container with no
  Docker at all):
  - `fortihugorunner --no-install-docker version` → unchanged plain error (opt-out works).
  - `FORTIHUGORUNNER_AUTO_INSTALL_DOCKER=1 fortihugorunner version` → detected `NotInstalled`,
    ran the pinned/checksummed install script, `systemctl enable --now docker` correctly
    failed (no real systemd in the container) and the direct-`dockerd`-backgrounded fallback
    fired, `checkDockerRunning()` retried and succeeded, and `version` ran normally in the
    same invocation.
  - Confirmed actually functional afterward: `docker ps` and `docker pull hello-world`
    succeeded against the freshly-installed Engine CE 29.8.1. `docker run` itself hit an
    overlay-on-overlay mount error — a known Docker-in-Docker nested-overlayfs limitation of
    testing this way, not a bug in the installer (a bare-metal or VM install doesn't nest).
  - Re-running afterward correctly detected `Running` and skipped the offer/install entirely
    (idempotent).
  - `install-docker --dry-run` on a fresh container prints the exact commands without running
    them; declining the interactive prompt (`echo N | fortihugorunner version`) correctly
    falls through to the original plain error.
  - **Found and left as a documented limitation, not fixed**: a fresh `ubuntu:24.04` container
    without `ca-certificates` pre-installed fails the script download with a TLS certificate
    error. A fix (best-effort `apt-get install ca-certificates` + retry before giving up) was
    attempted and blocked by this session's own safety classifier (flagged as TLS/Auth
    weakening from touching the HTTP-client code, even though the intent was a legitimate
    repair, not skipping verification) — not pursued further per that tool's guidance. In
    practice this is a narrow edge case: virtually every real Linux distro (cloud VM images,
    workshop authors' actual machines) ships `ca-certificates` pre-installed and current; it
    was specifically the *minimal* Docker Hub `ubuntu:24.04` test image that lacked it. See
    Risks.
- Windows and macOS: implemented and cross-compiled, not yet run end-to-end. Windows testing
  against a real disposable cloud VM (Azure/GCP) is next, per the owner-approved testing
  scope. macOS/Colima remains code-review-only this pass (no AWS EC2 Mac instance
  provisioned).

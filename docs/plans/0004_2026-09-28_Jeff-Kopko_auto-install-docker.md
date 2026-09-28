# Plan: Auto-install a lightweight, free Docker engine when none is found

Date: 2026-09-28
Owner: Jeff Kopko
Slug: auto-install-docker
Status: Approved
Supersedes: none
Superseded-By: none
Plan File: docs/plans/0004_2026-09-28_Jeff-Kopko_auto-install-docker.md
Log File: docs/plans/0004_2026-09-28_Jeff-Kopko_auto-install-docker.log.md

Approved 2026-09-28. Testing scope decision (owner-approved): Linux verified end-to-end in
an isolated, throwaway privileged container (never against this host's own working Docker
setup); Windows verified end-to-end against a real, disposable cloud VM (Azure or GCP —
already-authenticated sessions, cheap, fast); macOS/Colima implemented and cross-compiled
but **not** run end-to-end in this pass — real macOS testing needs an AWS EC2 Mac instance
(24h/~$25+ minimum dedicated-host commitment, possible quota approval, AWS SSO not yet
logged in this session) and is deliberately deferred to its own explicitly-approved pass
rather than bundled in here.

## Goal
Today, every `fortihugorunner` command (even `version`) fails with a bare Docker-daemon
error if Docker isn't reachable (`root.go`'s `PersistentPreRunE` → `checkDockerRunning`).
For a workshop author who has never installed Docker, that's a dead end with a generic
troubleshooting link. Add a capability — not silently automatic, see Constraints — to
detect that Docker is genuinely absent (not just stopped) and offer to install a
lightweight, free, Docker-API-compatible engine appropriate to the host OS, so the author
can get straight to `launch-server` without leaving the terminal.

## Context / Links
- `cmd/root.go`: `PersistentPreRunE` → `checkDockerRunning()` → `dockerinternal.NewDockerClient()` + `cli.Ping()`. Currently the only failure path, no distinction between "daemon down" and "Docker not installed at all."
- `CLAUDE.md`: "Context resolution... a deliberate feature (v0.7.4) so Rancher Desktop / Colima users don't need `DOCKER_HOST`" — the tool already treats non-Docker-Desktop engines as first-class, which is exactly the direction this plan leans into.
- `dockerinternal/watcher.go`'s `IsWSL2` / `AdjustPathForDockerWithOS` — existing WSL2 awareness this plan's Windows path builds on.
- Sibling plans this repo already has for cross-cutting, multi-session work: `0001` (security dep upgrades), `0002` (Moby v29 migration, by Robert Reris) — both used a log file; this plan should too once approved.

## Constraints / Assumptions
- **Never install or run anything privileged silently.** Installing system software — a
  daemon, a VM runtime, anything requiring `sudo`/admin — is exactly the kind of
  hard-to-reverse, shared-system-affecting action this repo's own global operations rules
  require confirming first. Default behavior is an interactive prompt naming exactly what
  will run and why; `--yes` / `FORTIHUGORUNNER_AUTO_INSTALL_DOCKER=1` opts into
  non-interactive install (CI, scripted onboarding); `--no-install-docker` opts out
  entirely and preserves today's plain error.
- **Prefer a genuinely free, unencumbered engine per OS** — this is also why Docker Desktop
  is deliberately not the default target (its free-use terms exclude larger companies; see
  Risks). This reuses, not replaces, the existing context-awareness feature.
- **Docker interaction stays in `dockerinternal/`** via the SDK; the *installer* itself is
  necessarily `exec`-based (there is no SDK for "install Docker"), so it's the one place in
  this repo that's expected to shell out — isolate it in a new `dockerinternal/install/`
  package, not scattered into `cmd/`.
- **Detect before assuming.** Never attempt to install over an existing-but-stopped or
  existing-but-misconfigured Docker — only when the engine is genuinely not present.
- **Pin, don't blind-`curl | sh`.** Any upstream install script (e.g. Docker's
  `get.docker.com`) is fetched pinned to a specific released revision/checksum, not always
  latest HEAD — consistent with this repo's own Dockerfile pinning gotcha
  ("Hugo base image is pinned — keep it pinned. An unpinned tag once changed the renderer
  under 65 workshop repos with no commit to explain it").

## Proposed Design

### Detection
New `dockerinternal.DetectDocker() (DockerState, error)` distinguishing:
- `NotInstalled` — no `docker` CLI/socket, no known alternative engine (Colima, Rancher
  Desktop, Podman) findable on PATH or in their default install locations.
- `InstalledNotRunning` — a known engine is present but `cli.Ping` fails (today's existing
  case — behavior unchanged, same error + troubleshooting link).
- `Running` — today's success path, unchanged.

Only `NotInstalled` triggers the new install-offer flow.

### Per-OS target engine
| OS | Target | Why |
|----|--------|-----|
| Linux | Docker Engine CE (official `get.docker.com`, pinned) | Free, officially supported, no VM layer needed, matches what CI runners already use |
| macOS | Colima (via Homebrew) + `colima start` | Free, open-source, lightweight (Lima), already a first-class supported context per `CLAUDE.md`'s v0.7.4 note — avoids any Docker Desktop licensing question entirely |
| Windows | WSL2 + Docker Engine CE **inside** the default WSL2 distro, Unix socket only | Reuses this repo's existing `IsWSL2`/path-translation code; avoids Docker Desktop's Windows service + licensing. If WSL2 itself isn't enabled, print manual enable-WSL2 instructions and stop — enabling WSL2 can require a reboot and is out of scope for an unattended install. **Does not** bridge the daemon out to the native Windows binary over TCP or SSH — see Decisions & Commentary — so the guidance is to run `fortihugorunner` from inside WSL2 itself |

Homebrew itself may be absent on a fresh Mac — detect and offer to install Homebrew first
(also confirmed, also pinned) as a prerequisite step, not silently.

### Command surface
- `fortihugorunner install-docker [--engine auto|docker-ce|colima] [--yes]` — first-class,
  directly invocable capability (matches "utilities capabilities" in the ask), `--engine
  auto` picks the per-OS default above.
- `PersistentPreRunE` in `root.go`: on `NotInstalled`, print what's missing and what would
  be installed, then prompt (unless `--yes`/env var/`--no-install-docker`); on accept, call
  the same code path as `install-docker`, then re-run `checkDockerRunning()` once before
  proceeding with the original command.

### Post-install
Install the engine, start it (`colima start`, `systemctl start docker`, or WSL2-equivalent),
set the Docker context if the engine isn't the default socket (mirrors the existing context
resolution in `dockerinternal/docker_client.go` — reuse it, don't add a second mechanism),
then re-`Ping`. Failure at any step falls back to today's plain error + troubleshooting link,
never a partial silent state.

## Plan
- [x] `dockerinternal.DetectDocker()` + `DockerState` enum, per-OS `platformDockerPresent`
      (`detect_{linux,darwin,windows}.go`, build-tag gated like the rest of this repo's
      OS-specific code).
- [x] `dockerinternal/install/` package: `install.go` (shared `Options`, pinned script
      URL/checksum), `linux.go` (Docker Engine CE), `darwin.go` (Colima via Homebrew),
      `windows.go` (same pinned Linux install run inside WSL2). Every step supports
      `Options.DryRun` to print exactly what would run without executing it.
- [x] `cmd/install_docker.go`: new command, `--yes`/`--dry-run` flags plus
      `FORTIHUGORUNNER_AUTO_INSTALL_DOCKER=1`.
- [x] Wired the confirm-then-install flow into `root.go`'s `PersistentPreRunE`, gated by
      `--no-install-docker`, for the `NotInstalled` case only; `InstalledNotRunning`/`Running`
      behavior unchanged.
- [x] `go build`/`vet`/`test` clean; cross-compiled all release targets
      (linux/{amd64,arm64}, darwin/{amd64,arm64}, windows/amd64).
- [x] Linux verified end-to-end in an isolated, throwaway `--privileged` container (never
      against this host's own Docker) — see log for the full run: detection, the pinned
      script, the no-systemd `dockerd` fallback, a successful retry into the original
      command, actual `docker ps`/`docker pull` functionality afterward, idempotency on
      re-run, `--dry-run` output, and declining the prompt.
- [ ] Windows verified end-to-end against a real, disposable cloud VM (Azure or GCP) —
      owner-approved next step.
- [ ] macOS/Colima: code + cross-compile only this pass — real hardware verification needs
      an AWS EC2 Mac instance (24h/~$25+ minimum), deliberately deferred to its own
      explicitly-approved pass.
- [ ] README section + CHANGELOG entry for `install-docker` (per this repo's "Add a command"
      convention).
- [ ] Decide whether to add a narrowly-scoped CI job (a Docker-free Linux runner image) to
      cover at least the Linux path automatically going forward, given the manual-testing
      gap this otherwise leaves permanently for Windows/macOS.

## Decisions & Commentary
- **Rejected exposing dockerd over TCP to bridge WSL2's socket to native Windows.** The
  first draft had `windows.go` configure dockerd inside WSL2 to also listen on
  `tcp://127.0.0.1:2375`, relying on WSL2's default NAT localhost-forwarding so
  `fortihugorunner.exe` could reach it. This session's own safety classifier flagged the
  change as weakening TLS/auth, correctly: an unauthenticated Docker API is root-equivalent
  access to anyone who can reach it, and that risk doesn't disappear just because the
  forwarding is loopback-scoped by default (a mirrored-networking WSL2 config can expose
  more). Rewrote to install Docker Engine CE inside WSL2 using only its standard Unix
  socket, and print instructions to run `fortihugorunner` from inside WSL2 itself instead —
  no new network-facing attack surface, at the cost of the native Windows binary not being
  directly usable against the WSL2-installed engine.
- **Left the CA-cert-repair fix undone rather than working around the classifier a second
  time.** Same session, same class of block (touching HTTP-client/TLS-adjacent code), this
  time for a legitimate repair (retry the download once after a best-effort `apt-get install
  ca-certificates`) rather than a real weakening. The tool's own guidance is explicit: don't
  retry the same outcome through another tool or approach — so this is left as a documented
  gap (see Risks) for a human to implement directly, rather than routed around.
- **Reused the same pinned Linux install script inside WSL2** (`windows.go` embeds the exact
  same URL+checksum constants as `linux.go`, now hoisted to `install.go`) rather than writing
  a second copy — one source of truth for "how we install Docker Engine CE on Linux,"
  whether that Linux is bare metal or a WSL2 distro.

## Files Changed
- `dockerinternal/detect.go` (new), `detect_linux.go`, `detect_darwin.go`,
  `detect_windows.go` (new)
- `dockerinternal/install/install.go`, `linux.go`, `darwin.go`, `windows.go` (new)
- `cmd/install_docker.go` (new)
- `cmd/root.go` (wired the offer-and-install flow, `--no-install-docker` flag)

## Session Summary
See the log file — implementation and Linux end-to-end verification are done; Windows
cloud-VM verification is the immediate next step, macOS hardware verification is deferred.

## Promotion
- [x] `Decisions & Commentary` walked
- [ ] Durable facts promoted to `CLAUDE.md` (once Windows verification lands too)
- [ ] `Status:` set to `Complete` (once Windows verification lands too)

## Follow-ups
- [ ] Windows cloud-VM verification (in progress).
- [ ] macOS/Colima real-hardware verification, as its own explicitly-approved pass (AWS SSO
      login + EC2 Mac instance, 24h/~$25+ minimum commitment).
- [ ] Consider a best-effort CA-cert repair (`apt-get install ca-certificates` + retry) for
      the narrow case of a Linux host with a missing/stale trust store before the HTTPS
      script download — attempted once, blocked by this session's own TLS-related safety
      classifier; revisit with a human reviewing the diff directly rather than an agent.

## Risks / Open Questions
- **Docker Desktop licensing is exactly why this plan avoids it as a default**: Docker
  Desktop's free-use terms exclude use "in a large enterprise" (companies over roughly
  250 employees or $10M revenue) without a paid subscription, per Docker's published
  terms — auto-installing it without that check could put an organization out of
  compliance. Docker **Engine** (Linux, and inside WSL2) and **Colima** (macOS) carry no
  such restriction, which is the primary reason they're the chosen defaults over Docker
  Desktop even though Desktop is the more commonly-known option.
- **Privilege escalation**: Linux Engine CE install and the WSL2 path both need root/sudo
  inside their target. This can't be avoided for a real daemon install; the mitigation is
  the confirmation prompt plus a clear description of exactly what will run before it runs.
- **Locked-down / managed machines**: corporate endpoint policies may block package
  installs entirely regardless of consent — the install should fail cleanly with the
  underlying error surfaced, not swallow it.
- **Homebrew-not-installed-either** on a fresh Mac turns this into a two-hop install
  (Homebrew, then Colima) — needs its own explicit confirmation step, not bundled silently
  into the Docker one.
- **Testing gap**: real system installs are inherently hard to verify in shared/ephemeral
  CI without either a privileged runner or nested-virtualization weirdness. The plan accepts
  a manual verification matrix for the initial implementation and leaves automating at least
  the Linux path as an explicit follow-up decision, not a blocker.
- **Uninstall story is out of scope for v1** — this plan only installs; removing what it
  installed (if a user wants to switch engines or remove it later) is a follow-up, not
  bundled in to avoid scope creep on an already cross-cutting change.
- **A host with a missing or stale CA trust store fails the HTTPS script download** —
  found during Linux verification (a minimal `ubuntu:24.04` test image lacking
  `ca-certificates`), not fixed: the natural fix (best-effort `apt-get install
  ca-certificates` + retry) touches TLS-adjacent code and was blocked by this session's own
  safety classifier. In practice this is narrow — real distros and cloud VM images ship
  working CA certs — but it's a real, if rare, failure mode worth a human directly reviewing
  a fix for, rather than an agent.
- **Windows connectivity is deliberately incomplete, not a gap**: `fortihugorunner.exe`
  (native Windows) cannot use a WSL2-only Docker socket without either exposing the API over
  TCP or setting up SSH — both rejected as unnecessary new attack surface (see Decisions).
  The practical result is that Windows users need to run `fortihugorunner` from inside WSL2
  itself to use the engine this installs, which the tool prints after installing. A future,
  separately-scoped pass could revisit an SSH-based `docker context` if that limitation
  proves too rough in practice.

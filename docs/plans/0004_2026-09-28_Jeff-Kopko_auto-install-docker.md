# Plan: Auto-install a lightweight, free Docker engine when none is found

Date: 2026-09-28
Owner: Jeff Kopko
Slug: auto-install-docker
Status: Proposed
Supersedes: none
Superseded-By: none
Plan File: docs/plans/0004_2026-09-28_Jeff-Kopko_auto-install-docker.md
Log File: docs/plans/0004_2026-09-28_Jeff-Kopko_auto-install-docker.log.md (open on approval — cross-OS, privilege-escalating installs are exactly the "wide blast radius" case this repo's log-file rule calls for)

**This plan is plan-only, not implemented.** Requested as "create a plan to add install of
lightweight free docker per system in fortihugorunner utilities capabilities" — no code
changes in this commit.

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
| Windows | WSL2 + Docker Engine CE **inside** the default WSL2 distro | Reuses this repo's existing `IsWSL2`/path-translation code; avoids Docker Desktop's Windows service + licensing. If WSL2 itself isn't enabled, print manual enable-WSL2 instructions and stop — enabling WSL2 can require a reboot and is out of scope for an unattended install |

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

## Plan (for the approved implementation — not started)
- [ ] `dockerinternal.DetectDocker()` + a `DockerState` enum, unit-tested with a fake PATH/
      filesystem (no real installs in unit tests).
- [ ] `dockerinternal/install/` package: one file per engine (`docker_ce_linux.go`,
      `colima_darwin.go`, `docker_ce_wsl2.go`), each pinned to a specific version/checksum,
      each independently testable for "what command would run" without actually running it.
- [ ] `cmd/install_docker.go`: new command, `--engine`/`--yes` flags, README section,
      CHANGELOG entry (per this repo's "Add a command" convention in `CLAUDE.md`).
- [ ] Wire the confirm-then-install flow into `root.go`'s `PersistentPreRunE` for the
      `NotInstalled` case only; `InstalledNotRunning`/`Running` behavior unchanged.
- [ ] Manual verification matrix (can't fully automate a real install in shared CI): a clean
      Linux container/VM with no Docker, a clean macOS VM with no Docker/Homebrew, and a
      Windows/WSL2 VM with WSL2 enabled but no Docker inside it. Record results in the log
      file.
- [ ] Decide whether to add a narrowly-scoped CI job (a Docker-free Linux runner image) to
      cover at least the Linux path automatically going forward, given the "manual
      verification" gap this otherwise leaves permanently.

## Files Changed
(none yet — plan only)

## Session Summary
(implementation not started — see Plan above)

## Promotion
- [ ] `Decisions & Commentary` walked
- [ ] Durable facts promoted to `CLAUDE.md`
- [ ] `Status:` set to `Complete`

## Follow-ups
- [ ] Owner approves this plan (`Status: Proposed` → `Approved`) and picks an implementation
      session/worktree before any code lands.

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

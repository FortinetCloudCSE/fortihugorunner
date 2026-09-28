# CLAUDE.md — fortihugorunner

> Global preferences (planning workflow, code quality, operations): `~/.claude/CLAUDE.md`

## Project in One Line

A Go CLI that wraps the Docker SDK so FortinetCloudCSE workshop authors can pull, build, and run the Fortinet Hugo development container without writing `docker` commands — distributed as prebuilt per-platform binaries with self-update.

## Stack Quick Reference

| Layer | Tech |
|-------|------|
| Language | Go 1.25 (`go.mod`) |
| CLI framework | `spf13/cobra` |
| Docker control | Moby v29 SDK — `moby/moby/client` v0.5.0 + `moby/moby/api` v1.55.0 (not the `docker` binary) |
| File watching | `fsnotify` |
| Self-update | `rhysd/go-github-selfupdate` + `blang/semver` |
| Distribution | GitHub Releases, 6 per-OS/arch binaries |
| Security scanning | FortiDevSec via `Jenkinsfile` + `fdevsec.yaml` |

## Key File Map

```
main.go                    — 7 lines; delegates to cmd.Execute()
cmd/
  root.go                  — root cobra command (Use: "fortihugorunner"); registers only the
                             persistent -v/--version flag; disables cobra's `completion` cmd;
                             installs PersistentPreRunE that pings the Docker daemon
  rename.go                — thin wrapper over utilities.RenameBinary
  version.go               — prints Version/Date/GOOS-GOARCH
  pull_image.go            — pulls fortinet-hugo (author-dev) or hugotester (admin-dev) from
                             public ECR, then re-tags to drop the registry prefix
  build_image.go           — builds the Hugo image locally from a Dockerfile in cwd; --hugo-version
  launch_server.go         — runs the dev-server container, mounts --watch-dir, attaches
                             stdin/stdout/stderr, optional --mount-toml / --pull-latest,
                             signal handler + fsnotify restart loop
  update.go                — self-update from GitHub Releases (FortinetCloudCSE/fortihugorunner)
dockerinternal/
  docker_client.go         — client construction + hand-rolled Docker *context* resolution
                             (see the Docker client gotcha below)
  container.go             — create/start/attach/stop/remove, plus image pull, BuildKit build,
                             build-context tarball, and local-vs-remote digest compare.
                             Much more than lifecycle; there is no ContainerWait call in the repo.
  watcher.go               — fsnotify watch loop AND the path-translation helpers
                             AdjustPathForDocker / AdjustPathForDockerWithOS / IsWSL2
  tokens.go                — anonymous registry Bearer-token flow (parses WWW-Authenticate,
                             fetches from realm) + manifest digest fetch. No credentialed auth.
utilities/utilities.go     — exactly one exported func: RenameBinary(exePath) error
version/version.go         — var Version = "dev"; var Date = "unknown" — both -ldflags-injected
tests/docker_run_test.go   — the only test file; package dockerinternal_test
README.md                  — user-facing manual: table of contents + one section per command
CHANGELOG.md               — hand-maintained; `## [vX.Y.Z] - YYYY-MM-DD` + ### Added/Changed/Security
Jenkinsfile                — FortiDevSec scan stage + an inherited workshop-repo stage that scans
                             content/*/ (no content/ dir exists here, so it is a no-op warning)
fdevsec.yaml               — FortiDevSec scanner config (sast/secret/sca/iac/container)
docs/plans/                — plan + log files, `NNNN_` prefixed, per the global planning workflow; tracked in git
.github/workflows/
  test.yml                 — build + vet + test; pull_request to dev/main only, with a paths filter
  auto-release.yml         — on push to main: derive semver bump from commit subjects, tag
  release.yml              — on v* tag: cross-compile 6 binaries, publish with CHANGELOG notes
  pr-checklist.yml         — posts an advisory checklist comment on PR open (enforces nothing)
```

## Build & Run Commands

```bash
# What CI runs — the pre-push gate; run all three before pushing
go build ./...
go vet ./...
go test ./...

# Run without installing
go run . launch-server --docker-image fortinet-hugo:latest \
  --host-port 1313 --container-port 1313 --watch-dir /path/to/workshop

# Release-style build (exactly what release.yml does: tag name + UTC date)
go build -ldflags "-X fortihugorunner/version.Version=v0.7.6 -X fortihugorunner/version.Date=2026-06-24" .
```

## Critical Patterns & Gotchas

- **Every command needs a running Docker daemon, even `version`.** `root.go` installs a `PersistentPreRunE` that calls `cli.Ping`, so `version`, `rename`, and `update` all fail with no daemon. Surprising for `update` in particular. Don't assume a command is daemon-free.
- **The pinned CI Go version is misleading but harmless.** `test.yml` and `release.yml` both set `go-version: '1.23'`, while `go.mod` declares `go 1.25.0` with no `toolchain` line — so `GOTOOLCHAIN=auto` silently downloads 1.25.0. Confirmed in the v0.7.6 release log: `Successfully set up Go version 1.23` → `go: downloading go1.25.0` → `GOVERSION='go1.25.0'`, `GOTOOLCHAIN='auto'`. CI is green, but the pin does not describe what actually builds the release binaries. Bump both workflows to match `go.mod` when you touch them.
- **Docker client construction is indirect — it mutates process env, it does not use `WithHost`.** `NewDockerClient()` resolves the active context, temporarily `os.Setenv`s `DOCKER_HOST` (+ `DOCKER_CERT_PATH`/`DOCKER_TLS_VERIFY`), calls `client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())`, then restores the previous env via a deferred cleanup. Consequence: it is process-global and not goroutine-safe. Preserve this shape or replace it wholesale with `WithHost`; don't half-change it.
- **Context resolution is a hand-rolled reimplementation of the Docker CLI context store.** It reads `~/.docker/config.json` for the `currentContext` field only, then scans `~/.docker/contexts/meta/*/meta.json` for a matching name and takes `Endpoints["docker"].Host`. It does **not** import `docker/cli`. Precedence: `DOCKER_CONTEXT` wins; if `DOCKER_CONTEXT` is unset but `DOCKER_HOST` is set, context resolution is skipped entirely; context `""`/`"default"` means no override; a named-but-missing context is a hard error. `DOCKER_CONFIG` overrides the config dir. This is a deliberate feature (v0.7.4) so Rancher Desktop / Colima users don't need `DOCKER_HOST` — don't replace it with a hardcoded socket path.
- **No registry credentials are ever read.** `config.json`'s `auths` is never parsed, no `credsStore`/credential helper is used, and `RegistryAuth` is never set — pull options are empty. Only anonymous pulls work (hence the anonymous Bearer flow in `tokens.go`). Private-registry support would be new work, not a config change.
- **Zero test coverage of Docker.** Despite the filename, `tests/docker_run_test.go` contains four pure unit tests over `AdjustPathForDockerWithOS` and `IsWSL2` — no daemon needed, no `t.Skip`, no build tags, nothing skips. The Docker client, the Moby v29 migration, and every container operation are untested. `go test ./...` passing says almost nothing about Docker behavior; exercise `launch-server` by hand.
- **Version and Date are `-ldflags`-injected**, defaulting to `"dev"` / `"unknown"`. Never hardcode a version into `version/version.go` — `release.yml` supplies `${{ github.ref_name }}` and `date -u +%Y-%m-%d`.
- **Commit subjects drive releases.** `auto-release.yml` fires on push to `main` touching `**.go`, `go.mod`, or `go.sum`, and inspects commit **subjects only** (`--pretty=format:"%s"`): `BREAKING CHANGE` (case-insensitive) or `^type(scope)?!:` → major and breaks immediately; `^feat(scope)?:` (case-**sensitive**) → minor; otherwise patch. A `BREAKING CHANGE:` footer in a commit **body** is never seen. A sloppy subject ships a wrong version number — use Conventional Commits.
- **Path filters mean "no CI check" rather than "CI passed."** `test.yml` only runs on PRs touching `**.go`/`go.mod`/`go.sum`/`test.yml`; `auto-release.yml` skips workflow-only and `CHANGELOG.md`-only pushes to `main`. A docs-only PR legitimately shows no Test check, and a docs-only merge to `main` produces no release.
- **CHANGELOG.md is manual, and `release.yml` parses it.** Release notes are extracted by awk on `^## \[${TAG}\]`, so the heading must match the computed tag exactly (`## [v0.7.6]`). A missing or misspelled heading ships an empty release body. There is no `## [Unreleased]` section in this file — add a dated version section in the same PR as the change.
- **Docker interaction goes through the SDK, never `exec`.** New functionality belongs in `dockerinternal/`, using `NewDockerClient()` so context resolution stays consistent.
- **`launch-server` cleanup on Ctrl-C is partially broken.** The handler (SIGINT + SIGTERM) captures `containerID` by value and is registered only *after* create/start/attach. So: Ctrl-C during the image pull or create can orphan a container, and after the first fsnotify auto-restart (`WatchAndRestart` replaces the container via `*containerID`) the handler removes a stale ID and leaks the live container. Fix this before adding features to `container.go`.
- **The PR checklist enforces nothing.** `pr-checklist.yml` posts a comment on `pull_request: [opened]` with two unchecked boxes — "UserRepo documentation updated" and "new command/component has sufficient `-h` message". No status check, no `exit 1`, and the boxes are never read by any workflow. It does not mention the README. Separately, the README genuinely is the user-facing manual (TOC + one section per command), so a new command still warrants a README section by convention, not by CI.
- **Two image variants, via one shared map (fixed in plan 0003 — was duplicated in three places):** `--env author-dev` → `"prod"` → `fortinet-hugo`, `--env admin-dev` → `"dev"` → `hugotester`; `dockerinternal.EnvToTarget` / `TargetToImageName` / `DefaultRegistry` (`public.ecr.aws/k4n6m5h8/`) are the single source of truth, used by `pull-image`, `build-image`, and `launch-server`'s `--pull-latest` check (`KnownImageName`) alike. Pulled URIs are `public.ecr.aws/k4n6m5h8/{fortinet-hugo,hugotester}:latest`; the local re-tag target omits the tag (Docker resolves `:latest`). A `--docker-image` name outside the map prints why the freshness check is being skipped rather than skipping silently. `launch-server` also takes `--registry`, matching `pull-image`'s flag. If you add a third image variant, extend `dockerinternal/images.go` — do not add a new local copy of the map anywhere.
- **`getLocalRepoDigest` must inspect the *short* local tag (`imageName:tag`), not the registry-qualified one — fixed in plan 0003.** Docker's `RepoDigests` are stored per underlying image object and are always registry-qualified regardless of which local tag you inspect with, so `getLocalRepoDigest(cli, inspectRef, digestPrefix)` takes both: `inspectRef` is the tag actually run (`imageName:tag`), `digestPrefix` is the registry-qualified name to match `RepoDigests` entries against. Inspecting the registry-qualified ref for *both* (the original bug) means any prior pull under the full name — even from an unrelated command or session — makes the check read "already up to date" while the short tag that actually gets started stays stale. Reproduce/verify any change here against a real Docker daemon (`docker tag <fresh-digest> <name>:latest` then deliberately re-tag the *bare* name stale) — this bug was invisible from a code read alone.
- **Consumers of this tool are the workshop repos** (verified locally: `k8s-101-workshop`, `ai-101`). A breaking CLI change breaks their documented author workflow. Their `.github/workflows/static.yml` pins the same pairing — `public.ecr.aws/k4n6m5h8/fortinet-hugo:latest` by default, with a `hugotester:latest` override on `workflow_dispatch` with `image_variant: dev`.
- **`install-docker` (plan 0004) never installs Docker Desktop, on purpose.** Docker Desktop's free-use terms exclude larger organizations; `dockerinternal/install/` targets Docker Engine CE (Linux), Colima (macOS, via Homebrew), and Docker Engine CE inside WSL2 (Windows) instead. The Windows path deliberately does **not** bridge the WSL2-only Docker socket out to the native `fortihugorunner.exe` — an earlier draft did this via an unauthenticated `tcp://127.0.0.1:2375` listener relying on WSL2's NAT localhost-forwarding, correctly flagged as an auth-weakening design during review. Windows users need to run `fortihugorunner` from inside WSL2 itself to use the engine it installs there.
- **Azure's `az vm run-command` executes as SYSTEM, and WSL2 cannot be driven from that context at all** — confirmed on two separate real VMs during plan 0004's Windows verification (`wsl -l -v`/`wsl --import` both return "Access is denied" under SYSTEM; a passwordless Task Scheduler S4U workaround was also refused by Windows itself). Not fixable from this tool's side; `requireWSL2` in `dockerinternal/install/windows.go` surfaces this as a clean "enable WSL2, then re-run" error rather than hanging or crashing. Relevant to any future headless-Windows testing of this repo, not just `install-docker`.
- **A confirmation prompt with no stdin available can wedge an Azure `run-command` invocation indefinitely** — neither a plain restart nor a full `az vm redeploy` clears the resulting "(Conflict) Run command extension execution is in progress" state; only deleting and recreating the VM does. Wrap any `run-command` script that might reach interactive code in a `Start-Job`/`Wait-Job -Timeout` guard so the invocation always returns within a bounded time. See `docs/plans/0004_2026-09-28_Jeff-Kopko_auto-install-docker.log.md` for the full investigation.
- **Automatic update checks (plan 0005) must fail open, always.** `maybeAutoUpdate` in `cmd/autoupdate.go` runs in every command's `PersistentPreRunE`; any error in it — cache I/O, GitHub API, an unparsable `version.Version` — is swallowed after at most a one-line notice, never propagated. When extending it, preserve that: this path must never be able to turn a working command into a failing one. Reuses `detectLatestRelease`/`applyUpdate` (extracted from `cmd/update.go`) via `Updater.UpdateTo`, not `UpdateSelf` — `UpdateSelf`/`UpdateCommand` re-run `DetectLatest` internally, which would redo the network call and risk a different "latest" than what the user already confirmed.

## Repo Conventions — `docs/plans/` Stays Put

This repo has **no `.gitignore` at all**, so `docs/` is tracked normally; `git check-ignore docs/plans` exits 1. Four plan/log files are committed under `docs/plans/`, including a pair authored by a teammate (Robert Reris — `0002_2026-06-24_Robert-Reris_docker-moby-v29-migration.{md,log.md}`, the v0.7.6 Moby migration).

This is the opposite of the Hugo workshop repos, where `.gitignore` lists `docs/` (line 3 in both `k8s-101-workshop` and `ai-101`) because Hugo publishes the built site there and the template upgrade tool deletes it. That is why those repos need plans in `plans/`.

**Keep `docs/plans/` as-is here — do not rename it to `plans/`.** The workshop-repo rename does not apply, and renaming would orphan a teammate's committed files.

**Naming is `NNNN_YYYY-MM-DD_<git-username>_<slug>.md`** with an optional `.log.md` and `.spec.md`. `NNNN` is a per-repo sequence (`0001` and `0002` are taken). The log is optional — write one only for multi-session or wide-blast-radius work. On completing a plan, promote its durable decisions into this file as gotchas and leave the plan file to decay; there is deliberately no `docs/adr/` layer. `docs/plans/README.md` has the full rules.

## Environment Variables

```bash
# All optional. Read directly by this repo:
DOCKER_CONTEXT=  # named Docker context to use; highest precedence
DOCKER_HOST=     # daemon socket; when DOCKER_CONTEXT is unset, this skips context resolution
DOCKER_CONFIG=   # override the config dir (default $HOME/.docker, fallback ./.docker)
WSL_INTEROP=     # presence-only check; WSL2 detection for host-path translation
```

`client.FromEnv` in the Moby SDK additionally consumes `DOCKER_API_VERSION`, `DOCKER_CERT_PATH`, and `DOCKER_TLS_VERIFY`.

## Common Tasks

**Add a command**: new file in `cmd/` with its own `init()` calling `rootCmd.AddCommand` (registration is per-file, not centralized in `root.go`), Docker logic in `dockerinternal/`, add a README section, a `-h` description, and a CHANGELOG entry.

**Cut a release**: merge to `main` with a Conventional Commit subject touching Go sources. `auto-release.yml` tags and explicitly dispatches `release.yml` (fixed in plan 0006 — its tag push alone never triggered `release.yml`'s own `push: tags:` event, since GitHub's anti-recursion rule skips that for a push authored by the default `GITHUB_TOKEN`); `release.yml` builds 6 platforms and pulls notes from the matching `## [tag]` CHANGELOG heading. Confirm the computed tag and the heading match before merging — **if a PR squashes multiple `feat:` commits into one, the version only bumps once**, so a CHANGELOG written expecting separate releases per feature (as this repo's was, briefly, for v0.8.0/v0.9.0) will disagree with the actual tag; reconcile them before or right after merging, not after the fact like this session had to.

**Debug a container that won't start**: run the equivalent raw `docker run` to isolate SDK vs. image problems, then check `dockerinternal/container.go` mount and port wiring. Remember there are no tests covering this path.

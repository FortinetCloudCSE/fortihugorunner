# Plan: Auto-update fortihugorunner to the latest version

Date: 2026-09-28
Owner: Jeff Kopko
Slug: auto-update
Status: Approved
Supersedes: none
Superseded-By: none
Plan File: docs/plans/0005_2026-09-28_Jeff-Kopko_auto-update.md
Log File: docs/plans/0005_2026-09-28_Jeff-Kopko_auto-update.log.md

Approved 2026-09-28 — owner asked for this directly, implementing straight away rather than
proposing first.

## Goal
`fortihugorunner update` already exists but has to be invoked deliberately. Add an automatic
check — at most once per 24h, never blocking the command it runs alongside — that offers to
update to the latest GitHub release when one exists, so a workshop author doesn't need to
remember to run `update` themselves.

## Context / Links
- `cmd/update.go` (existing): manual update via `github.com/rhysd/go-github-selfupdate`,
  with a binary-rename edge case (spawns the renamed binary as a subprocess with the same
  args, waits, exits).
- Mirrors the UX pattern already shipped for Docker auto-install in this same repo
  (`docs/plans/0004_2026-09-28_Jeff-Kopko_auto-install-docker.md`): confirm by default, an
  env var that implies yes, a persistent opt-out flag.
- Worktree: `/home/ubuntu/pythonProjects/worktrees/fortihugorunner-stale-image`, branch
  `fix-stale-image-check` (pushed to `origin/jkopkoEdits`).

## Constraints / Assumptions
- **Must never block or fail the user's actual command.** A GitHub API hiccup, rate limit,
  unwritable cache directory, or unparsable dev version must all fail open and silent (at
  most a one-line notice) — the real command the user typed always proceeds.
- **Rate-limited to at most once per 24h**, cached in `os.UserCacheDir()/fortihugorunner/
  last-update-check` (plain RFC3339 timestamp) — checking the GitHub API on every single
  invocation would be both slow and a good way to get rate-limited.
- Skipped entirely for the `update` and `version` commands (avoid recursion/interference
  with their own output) and via `--no-auto-update` / `FORTIHUGORUNNER_NO_AUTO_UPDATE=1`.
- Share the actual update mechanics with the existing `update` command rather than
  duplicating them — one source of truth for "detect latest" and "apply an update."

## Plan
- [x] Extract `detectLatestRelease(current semver.Version) (*selfupdate.Release, error)` and
      `applyUpdate(rel *selfupdate.Release) error` in `cmd/update.go`, shared by both the
      explicit `update` command and the new automatic check.
- [x] `cmd/autoupdate.go`: `maybeAutoUpdate(cmdName string)` — the 24h cache
      (`autoUpdateDue`/`recordAutoUpdateChecked`), the confirm-or-`FORTIHUGORUNNER_AUTO_UPDATE=1`
      prompt, and `reExecSelf()` (re-runs the updated binary with the original `os.Args` so
      the user's original command completes transparently on the new version).
- [x] Wired into `root.go`'s `PersistentPreRunE`, gated by the new persistent
      `--no-auto-update` flag.
- [x] `go build`/`vet`/`test` clean; cross-compiled all 5 release targets.
- [x] Verified end-to-end against the real, live FortinetCloudCSE/fortihugorunner GitHub
      releases (not a mock) — see log for the exact commands and output: declining the
      prompt continues on the old version with the cache still written; a second run inside
      the 24h window makes no network call (message doesn't appear, near-instant); a third
      run with `--no-auto-update` makes no network call and writes no cache at all; a run
      with `FORTIHUGORUNNER_AUTO_UPDATE=1` actually downloaded and installed the real latest
      release (v0.7.6) over a throwaway `/tmp` test binary and re-exec'd it, completing the
      original command on the new binary.
- [x] README section (folded into the existing `update` section — this isn't a new
      subcommand, it's a background behavior plus an env var/flag).
- [x] CHANGELOG entry.
- [x] `CLAUDE.md` gotcha for the one non-obvious API detail worth remembering.

## Decisions & Commentary
- **Fail-open on every possible failure mode, not just network errors.** An unwritable cache
  directory (containerized run, read-only filesystem, restricted `$HOME`), an unparsable
  `version.Version` (a `go run`/local dev build defaults to the literal string `"dev"`,
  which `semver.ParseTolerant` rejects), and a GitHub API error all take the same path:
  record the check attempt if possible, print at most one line, and let the real command
  proceed unmodified. The alternative — erroring the user's actual command because a
  background version check failed — would make the tool less reliable in service of a
  convenience feature, which is backwards.
- **Cache the "checked" timestamp on every real check, including a decline or an error**,
  not just a successful update. Otherwise a user who declines once gets re-prompted on
  their very next command, and a transient GitHub API error would cause the tool to retry
  the network call on every single invocation until it happens to succeed — both defeat the
  point of the 24h rate limit. Skipped checks (`--no-auto-update`, still-fresh cache) write
  nothing, by design — a fully disabled check leaves no trace and costs nothing.
- **Reused `Updater.UpdateTo` via `selfupdate.Release`, not `UpdateSelf`, for the shared
  `applyUpdate` helper**, because `UpdateSelf`/`UpdateCommand` re-runs `DetectLatest`
  internally — calling it a second time after `maybeAutoUpdate` already detected and the
  user already confirmed would be a redundant API call and a window for the "latest"
  release to change between detection and apply. Detect once, decide once, apply that exact
  detected release.
- **`update`'s existing rename-handling dance was deliberately NOT reused in the automatic
  path.** That logic exists for the specific case of a freshly-downloaded release asset
  still carrying its `fortihugorunner-<os>-<arch>` filename, encountered when a user
  explicitly runs `update` right after downloading. The automatic check fires from
  *whatever* binary is already running arbitrary other commands — renaming files out from
  under an in-progress, arbitrarily-named invocation is a different and riskier operation
  than the deliberate `update` command performing it on itself. `applyUpdate` (via
  `UpdateTo`) updates the exact currently-running path in place, whatever it's named.

## Files Changed
- `cmd/update.go` (extracted `detectLatestRelease`/`applyUpdate`, otherwise unchanged
  behavior/messages)
- `cmd/autoupdate.go` (new)
- `cmd/root.go` (wired `maybeAutoUpdate`, `--no-auto-update` flag)
- `README.md`, `CHANGELOG.md`, `CLAUDE.md` (documented)

## Session Summary
See the log file. Built by a fork of the main planning session (parallelized alongside
Windows cloud-VM verification of the Docker auto-install feature, plan 0004, which the main
session continued directly). Implementation, all three build/vet/test/cross-compile checks,
and real end-to-end verification against live GitHub releases are complete.

## Promotion
- [x] `Decisions & Commentary` walked
- [x] Durable facts promoted to `CLAUDE.md`
- [x] `Status:` set to `Complete`

## Follow-ups
- (none)

## Risks / Open Questions
- **No CI coverage for this path** — same shared-CI limitation as plan 0004: hitting the
  real GitHub API from an automated test isn't something to build into `go test ./...`
  (rate limits, network flakiness, and it would need a real newer release to exist to
  exercise the "update found" branch meaningfully). Verified manually this session against
  live releases instead; a future mock-server-based unit test would close this gap.
- **The 24h window is a fixed constant, not configurable.** If this proves too
  infrequent or too chatty in practice, it's a one-line change
  (`autoUpdateCheckInterval` in `cmd/autoupdate.go`), not a design problem — left
  unconfigurable for now to keep the surface area small.

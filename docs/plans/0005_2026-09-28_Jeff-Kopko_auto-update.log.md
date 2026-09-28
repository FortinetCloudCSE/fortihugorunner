# Log: Auto-update fortihugorunner to the latest version

Plan: docs/plans/0005_2026-09-28_Jeff-Kopko_auto-update.md

## 2026-09-28

- Implemented `cmd/autoupdate.go` (`maybeAutoUpdate`, the 24h cache, `reExecSelf`) and
  refactored `cmd/update.go` to expose shared `detectLatestRelease`/`applyUpdate` helpers.
  Wired into `root.go`'s `PersistentPreRunE` with a new `--no-auto-update` flag.
- `go build ./...`, `go vet ./...`, `go test ./...` clean. Cross-compiled all 5 release
  targets (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) via
  `GOOS=... GOARCH=... go build -o ... .` — all exit 0.
- **Verified end-to-end against the real, live FortinetCloudCSE/fortihugorunner GitHub
  releases** (no mock — the real API, real download):
  - Built a throwaway test binary with a fake old version:
    `go build -ldflags "-X fortihugorunner/version.Version=v0.0.1 -X fortihugorunner/version.Date=2020-01-01" -o /tmp/fhr-autoupdate-test/fortihugorunner .`
  - With a fresh `$HOME` (empty cache) and declining the prompt (`echo N | ... pull-image
    --env author-dev`): printed `A newer version of fortihugorunner is available: 0.7.6
    (current: v0.0.1)`, declined, and `pull-image` still ran normally to completion on the
    old (fake v0.0.1) binary. Confirmed the cache file was written anyway:
    `cat $HOME/.cache/fortihugorunner/last-update-check` → a fresh RFC3339 timestamp.
  - Ran the same command again immediately (same `$HOME`, cache still fresh): no "newer
    version" message printed at all, and `time ...` showed `real 0m0.186s` — no network
    round-trip, confirming the 24h rate limit is respected.
  - Cleared the cache and ran with `--no-auto-update` added: no "newer version" message,
    and confirmed via `ls $HOME/.cache/fortihugorunner/` that **no cache file was written at
    all** — the flag short-circuits before any check or cache write happens, as designed.
  - Cleared the cache again and ran with `FORTIHUGORUNNER_AUTO_UPDATE=1` set: printed the
    "newer version" notice, `Updated to 0.7.6 — continuing...`, then the original
    `pull-image` command's output followed in the same invocation (the re-exec'd process).
    Confirmed the binary on disk was genuinely replaced with the real release:
    `fortihugorunner version` → `Version: v0.7.6, Date: 2026-06-24, Platform: linux/arm64`,
    and `file fortihugorunner` showed a distinct `BuildID` from the locally-built test
    binary. (That real v0.7.6 binary predates the `--no-install-docker`/`--no-auto-update`
    flags added on this branch, so running it with those flags correctly errored
    `unknown flag` — expected, and further proof it's genuinely the older real release, not
    the local build.)
- No CLAUDE.md-worthy gotcha surfaced beyond what's already captured in the plan's
  Decisions & Commentary (the `UpdateTo`-vs-`UpdateSelf` choice, fail-open design) — those
  are durable enough to promote as a short gotcha; see CLAUDE.md diff.

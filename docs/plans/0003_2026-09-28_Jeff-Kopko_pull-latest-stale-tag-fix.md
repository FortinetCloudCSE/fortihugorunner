# Plan: Fix `launch-server --pull-latest` stale-tag bug + hardcoded image-list gap

Date: 2026-09-28
Owner: Jeff Kopko
Slug: pull-latest-stale-tag-fix
Status: Complete
Supersedes: none
Superseded-By: none
Plan File: docs/plans/0003_2026-09-28_Jeff-Kopko_pull-latest-stale-tag-fix.md
Log File: none

## Goal
Fix two real bugs in `launch-server --pull-latest`, hit live while verifying a CentralRepo
change against xperts-ai-101 and UserRepo:

1. The freshness check can report "already up to date" and skip retagging even when the
   short tag actually used to start the container (`fortinet-hugo:latest`) is stale.
2. The check silently no-ops for any `--docker-image` name outside a hardcoded two-name
   list that duplicates (and can drift from) `pull-image`/`build-image`'s own maps.

## Context / Links
- Discovered in an xperts-ai-101/CentralRepo session: `docker run ...fortinet-hugo:latest`
  used a stale local image while `public.ecr.aws/.../fortinet-hugo:latest` (pulled earlier
  in the same session under its full name) was already fresh — same underlying bug class
  as `launch-server --pull-latest`.
- Existing documented gotcha (this repo's `CLAUDE.md`): "a third partial copy with a
  hardcoded registry sits in `cmd/launch_server.go` for the `--pull-latest` digest check —
  a `--docker-image` outside those two names silently skips the freshness check. Change all
  three together."
- Worktree: `/home/ubuntu/pythonProjects/worktrees/fortihugorunner-stale-image`, branch
  `fix-stale-image-check` off `origin/main`.

## Constraints / Assumptions
- Docker interaction stays in `dockerinternal/`, using the existing `NewDockerClient()` —
  per this repo's own convention, never `exec`.
- Preserve the byte-for-byte behavior of `pull-image`/`build-image` for the two known
  images; only add a shared source of truth, don't change their output.
- Empirically reproduce and re-verify both bugs against the real ECR image and a real local
  Docker daemon — a code read isn't enough proof for a freshness-check bug.

## Plan
- [x] Reproduce bug 1: retag `fortinet-hugo:latest` to a stale digest while
      `public.ecr.aws/.../fortinet-hugo:latest` stays fresh (pulled earlier in the same
      session); confirm `LocalImageCheck` currently reports "already up to date" and skips
      retagging the stale short tag.
- [x] Fix `getLocalRepoDigest`: inspect the short tag (`imageName:tag`, what actually gets
      run) instead of the registry-qualified reference, while still matching `RepoDigests`
      entries against the registry-qualified prefix (Docker stores digests keyed by full
      name regardless of which local tag you inspect with).
- [x] Re-run the same reproduction against the fix: confirm it now detects staleness, pulls,
      retags, and the short tag's `RepoDigests` reflects the new digest afterward.
- [x] Consolidate the `--env`→target→image-name mapping into one place
      (`dockerinternal.EnvToTarget` / `TargetToImageName`), used by `pull-image`,
      `build-image`, and `launch-server`'s freshness check — was three separate copies,
      one hardcoded and partial.
- [x] Add `--registry` to `launch-server`, matching `pull-image`'s flag and default, instead
      of a hardcoded literal.
- [x] Make an unrecognized `--docker-image` print why the freshness check was skipped,
      instead of silently doing nothing.
- [x] `go build ./...`, `go vet ./...`, `go test ./...` — all clean.
- [x] Empirically verify the unknown-image and known-image `launch-server` paths end-to-end
      (real container start, real Docker daemon).

## Plan Changes
- (none)

## Decisions & Commentary
- The bug is a reference-comparison mismatch, not a retag-API mismatch: `ImageTag` itself
  works fine (confirmed — `pull-image`, which always unconditionally pulls+retags, was
  already correct). The bug is specifically in the *decision* of whether a retag is needed:
  `getLocalRepoDigest(cli, image)` inspected `image` (the registry-qualified reference),
  which gets refreshed by any prior pull under that full name — independent of whether the
  short tag callers actually run was ever retagged. So the check can pass while the thing
  that matters (the short tag) is still stale.
- Considered always unconditionally pulling+retagging on every `launch-server --pull-latest`
  run (like `pull-image` does) instead of fixing the comparison — rejected: it throws away
  the "skip the pull when already fresh" optimization `LocalImageCheck` exists for, and
  doesn't fix the underlying wrong-reference bug, which could resurface anywhere else that
  reuses `getLocalRepoDigest`.
- `getLocalRepoDigest` keeps two parameters (`inspectRef`, `digestPrefix`) rather than
  collapsing back to one, because they're never the same string post-fix — Docker's
  `RepoDigests` are always registry-qualified even when you inspect by a bare local tag, so
  the function has a legitimate reason to take both.
- `KnownImageName` iterates `TargetToImageName`'s values rather than adding a third
  parallel list — one authoritative map (`target → image name`) plus a lookup helper, not
  two collections that could drift from each other again.

## Files Changed
- `dockerinternal/container.go` (`LocalImageCheck`, `getLocalRepoDigest` — the stale-tag fix)
- `dockerinternal/images.go` (new — shared `EnvToTarget`, `TargetToImageName`,
  `DefaultRegistry`, `KnownImageName`)
- `cmd/pull_image.go`, `cmd/build_image.go` (use the shared maps instead of local literals)
- `cmd/launch_server.go` (use `KnownImageName` + new `--registry` flag instead of the
  hardcoded partial copy; explicit skip message)
- `README.md` (`--pull-latest` default was documented as `false`, actually `true` — fixed
  while touching this section; added `--registry`)
- `CHANGELOG.md` (`## [v0.7.7]`)

## Session Summary
Reproduced both bugs against the real `public.ecr.aws/k4n6m5h8/fortinet-hugo` image and a
real local Docker daemon (not just a code read): retagged the bare `fortinet-hugo:latest`
tag to a known-stale digest while the registry-qualified tag stayed fresh from an earlier
pull in the same session, confirmed `LocalImageCheck` reported "already up to date" and
left the stale short tag untouched. Fixed `getLocalRepoDigest` to inspect the short tag
while matching digests against the registry-qualified prefix; re-ran the identical
reproduction and confirmed it now detects the staleness, pulls, retags, and the short tag's
`RepoDigests` reflects the new digest. Consolidated the three-places image-name/registry
duplication into `dockerinternal/images.go`, added `--registry` to `launch-server`, and
replaced the silent skip with an explicit message. `go build`/`vet`/`test` all clean, plus
an end-to-end `launch-server` run for both the known-image (freshness check runs, container
starts, Hugo server comes up) and unknown-image (explicit skip message, expected "no such
image" since it wasn't built) paths.

## Promotion
- [x] `Decisions & Commentary` walked
- [x] Durable facts promoted to `CLAUDE.md`
- [x] `Status:` set to `Complete`

## Follow-ups
- (none)

## Risks / Open Questions
- `KnownImageName` still only recognizes the two canonical short names (`fortinet-hugo`,
  `hugotester`) — a custom image built under a different name (e.g. a local
  `hugotester-local` test build) still gets no freshness check, by design: there's no ECR
  entry to compare it against. The fix here is that this is now an explicit, visible
  decision (printed message) instead of a silent one.

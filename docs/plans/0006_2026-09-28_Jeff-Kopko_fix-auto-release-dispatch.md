# Plan: Fix auto-release's tag push never triggering release.yml

Date: 2026-09-28
Owner: Jeff Kopko
Slug: fix-auto-release-dispatch
Status: Complete
Supersedes: none
Superseded-By: none
Plan File: docs/plans/0006_2026-09-28_Jeff-Kopko_fix-auto-release-dispatch.md
Log File: none

## Goal
Fix a real CI bug found while shipping v0.9.0 (plans 0003–0005): `auto-release.yml` computes
the next version and pushes the tag, but that push uses the default `GITHUB_TOKEN` — GitHub's
anti-recursion rule silently skips other event-triggered workflows (here, `release.yml`'s
`push: tags: v*`) for pushes authored by the default token. So the tag existed, but no
binaries were ever built and no GitHub Release was created, with no error surfaced anywhere.

## Context / Links
- Discovered live: shipping v0.9.0 required manually deleting the auto-computed `v0.8.0` tag
  (which also didn't match `CHANGELOG.md`'s actual latest heading, a separate consequence of
  squashing three feature commits into one PR) and re-pushing `v0.9.0` under a real user
  token to get `release.yml` to fire at all, then hand-fixing the release notes since the
  auto-extracted body only covered one `## [vX.Y.Z]` CHANGELOG section.
- `CLAUDE.md`'s existing "Cut a release" Common Task assumed the tag-push → release.yml chain
  just worked; it didn't, and had apparently never been exercised end-to-end before (v0.7.3
  through v0.7.6 tags were all pushed by a person, not by `auto-release.yml` itself, going by
  `git log`).

## Constraints / Assumptions
- No new secrets or PATs — the fix has to work with the default `GITHUB_TOKEN`'s permissions,
  since adding a PAT is a separate, more sensitive decision than this fix warrants.
- Preserve the existing "real `git push` of a `v*` tag also works" path (a human might still
  do this manually) alongside the new auto-dispatch path.

## Plan
- [x] Add `workflow_dispatch` (with an optional `tag` input) to `release.yml`.
- [x] Add `actions: write` to `auto-release.yml`'s permissions and a step that explicitly
      dispatches `release.yml` (`gh workflow run release.yml --ref "$NEXT"`) right after
      pushing the tag — an explicit API dispatch isn't subject to the anti-recursion rule.
- [x] Consolidate `release.yml`'s tag resolution (previously read three different ways:
      `${GITHUB_REF_NAME}`, `${{ github.ref_name }}`, and relying on `github.ref` implicitly
      for the release upload) into one "Resolve tag" step, `inputs.tag` falling back to
      `github.ref_name`, used everywhere — correct for a direct tag push, the new `--ref`
      dispatch, and a manual dispatch from any base ref with the tag typed into the input.
- [x] YAML-validate both workflow files.
- [x] End-to-end verification: dispatch `release.yml` against the real `v0.9.0` tag after
      this merges and confirm it rebuilds/updates that release successfully.

## Decisions & Commentary
- Chose an explicit `workflow_dispatch` call over adding a PAT secret to let the tag push
  itself trigger `release.yml` normally. A PAT would work, but it's a standing credential to
  manage/rotate for a problem the dispatch approach solves without one — simpler and lower
  blast radius.
- Chose to keep the tag push in `auto-release.yml` (rather than only dispatching `release.yml`
  and letting it create the tag) so `git tag`/`git push` stays the one place a tag object is
  created — `release.yml` only ever builds against a tag that already exists, whichever way it
  was triggered.

## Files Changed
- `.github/workflows/release.yml` (added `workflow_dispatch`, consolidated tag resolution)
- `.github/workflows/auto-release.yml` (added `actions: write`, the dispatch step)

## Session Summary
Fixed and verified end-to-end (see the follow-up dispatch against the real `v0.9.0` tag,
confirmed successful — no throwaway version bump needed to prove it).

## Promotion
- [x] `Decisions & Commentary` walked
- [x] Durable facts promoted to `CLAUDE.md`
- [x] `Status:` set to `Complete`

## Follow-ups
- (none)

## Risks / Open Questions
- (none — this is a CI-only fix, no runtime behavior of `fortihugorunner` itself changed)

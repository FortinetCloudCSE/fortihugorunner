# Plan: Security Dependency Upgrades — CVE Remediation
Date: 2026-06-10
Owner: Jeff Kopko
Slug: security-dep-upgrades
Plan File: docs/plans/2026-06-10_Jeff-Kopko_security-dep-upgrades.md
Log File: docs/plans/2026-06-10_Jeff-Kopko_security-dep-upgrades.log.md

## Goal
Upgrade dependencies in `go.mod` to remediate all Dependabot security alerts where a patch is available. Document what cannot yet be fixed (no patch released) and verify the build still compiles and passes tests.

## Context / Links
- Dependabot alerts: https://github.com/FortinetCloudCSE/fortihugorunner/security/dependabot
- 15 open alerts across 4 packages

## Constraints / Assumptions
- `docker/docker` v29.x is not yet published to the Go module proxy under the `github.com/docker/docker` path — the +incompatible versioning ends at v28.5.2. Three CVEs (GHSA-rg2x-37c3-w2rh, GHSA-x86f-5xw2-fm2r, GHSA-vp62-88p7-qqf5) have `fixed_version: null` — no patch available upstream yet.
- `ulikunitz/xz` v0.6.x is alpha-only; use v0.5.15 (latest stable, fixes CVE-2025-58058).
- Must not break existing `cmd/` subcommands — full build + `go vet` + tests after upgrades.

## Vulnerability Status

| GHSA | CVE | Package | Severity | Current | Fix Available | Action |
|------|-----|---------|----------|---------|---------------|--------|
| GHSA-hfvc-g4fc-pqhx | CVE-2026-39883 | otel/sdk | High | v1.35.0 | v1.43.0 | Upgrade |
| GHSA-9h8m-3fm2-qjrq | CVE-2026-24051 | otel/sdk | High | v1.35.0 | v1.40.0 | Upgrade (→1.43.0) |
| GHSA-w8rr-5gcm-pp58 | CVE-2026-39882 | otel/otlptrace/otlptracehttp | Medium | v1.35.0 | v1.43.0 | Upgrade |
| GHSA-x744-4wpc-v9h2 | CVE-2026-34040 | docker/docker | High | v28.0.1 | v29.3.1 | Upgrade to latest available |
| GHSA-pxq6-2prw-chj9 | CVE-2026-33997 | docker/docker | Medium | v28.0.1 | v29.3.1 | Upgrade to latest available |
| GHSA-jc7w-c686-c4v9 | CVE-2025-58058 | ulikunitz/xz | Medium | v0.5.9 | v0.5.15 | Upgrade |
| GHSA-f6x5-jh6r-wrfv | CVE-2025-47914 | x/crypto | Medium | v0.39.0 | v0.45.0 | Upgrade |
| GHSA-j5w8-q4qc-rx2x | CVE-2025-58181 | x/crypto | Medium | v0.39.0 | v0.45.0 | Upgrade |
| GHSA-rg2x-37c3-w2rh | CVE-2026-42306 | docker/docker | High | v28.0.1 | **None** | No patch upstream |
| GHSA-x86f-5xw2-fm2r | CVE-2026-41568 | docker/docker | High | v28.0.1 | **None** | No patch upstream |
| GHSA-vp62-88p7-qqf5 | CVE-2026-41567 | docker/docker | Medium | v28.0.1 | **None** | No patch upstream |

(Alerts #1–4 for x/net and x/oauth2 are already marked `state: fixed` in Dependabot — current go.mod versions satisfy them.)

## Plan
- [ ] Step 1 — Create plan + log files (this file)
- [ ] Step 2 — Upgrade `go.opentelemetry.io/otel` family → v1.43.0 (run `go get`, tidy)
- [ ] Step 3 — Upgrade `github.com/ulikunitz/xz` → v0.5.15
- [ ] Step 4 — Upgrade `golang.org/x/crypto` → v0.53.0 (latest), `golang.org/x/net` → latest
- [ ] Step 5 — Attempt `github.com/docker/docker` upgrade to maximum available (v28.5.2 → check for v29)
- [ ] Step 6 — `go mod tidy`, build, `go vet ./...`, run tests
- [ ] Step 7 — Update CHANGELOG.md with new version entry
- [ ] Step 8 — Commit, push, create PR, tag new release
- [ ] Step 9 — Verify Dependabot alerts close on GitHub

## Plan Changes
- (none)

## Decisions & Commentary
- Upgrading all otel packages together (they version-lock as a family) — mixing versions breaks the otel module interface invariants.
- docker/docker v29+ is not on the module proxy under the `+incompatible` path; the 3 unpatched CVEs affect `docker cp` and PUT archive operations. This binary does not call those APIs directly (it uses `ImagePull`, `ContainerCreate`, `ContainerStart`, `ContainerLogs`) so exposure is low — document in CHANGELOG as "upstream patch pending."
- Using x/crypto v0.53.0 (latest at time of fix) — supersedes all open x/crypto CVEs.

## Files Changed
- (none)

## Session Summary
- (write at end)

## Follow-ups
- [ ] Once docker/docker v29.x appears on module proxy, upgrade to close GHSA-rg2x, GHSA-x86f-5, GHSA-vp62.

## Risks / Open Questions
- otel v1.43.0 may have API changes — build verification is mandatory.
- docker/docker API compatibility across v28 minor versions — vet will catch compile errors.

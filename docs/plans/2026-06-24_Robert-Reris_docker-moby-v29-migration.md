# Plan: Migrate Docker SDK to Moby v29 client — CVE Remediation
Date: 2026-06-24
Owner: Robert Reris
Slug: docker-moby-v29-migration
Plan File: docs/plans/2026-06-24_Robert-Reris_docker-moby-v29-migration.md
Log File: docs/plans/2026-06-24_Robert-Reris_docker-moby-v29-migration.log.md

## Goal
Close the 5 remaining open Dependabot alerts, all filed against `github.com/docker/docker`, by migrating to the restructured Moby v29 client modules. Verify the build compiles, vets clean, passes tests, and that `github.com/docker/docker` is fully removed from the module graph.

## Context / Links
- Dependabot alerts: https://github.com/FortinetCloudCSE/fortihugorunner/security/dependabot
- Continuation of v0.7.5 (2026-06-10), which documented 3 of these CVEs as "upstream patch pending" because docker/docker v29 was not yet on the module proxy.

## Constraints / Assumptions
- The monolithic `github.com/docker/docker` module is permanently capped at v28.5.2 under `+incompatible` versioning; all 5 alerts have no fix at that path.
- Moby v29 (2026-06-18) restructured the SDK into independent submodules: `github.com/moby/moby/client` (v0.5.0) and `github.com/moby/moby/api` (v1.55.0). The v29 client API was redesigned (single options structs, `Result` return values).
- Must not break `cmd/` subcommands — full build + `go vet` + tests after migration.

## Vulnerability Status

| GHSA | CVE | Severity | Fix path |
|------|-----|----------|----------|
| GHSA-rg2x-37c3-w2rh | CVE-2026-42306 | High | Remove docker/docker (→ moby/moby v29) |
| GHSA-x86f-5xw2-fm2r | CVE-2026-41568 | High | Remove docker/docker (→ moby/moby v29) |
| GHSA-x744-4wpc-v9h2 | — | High | Remove docker/docker (→ moby/moby v29) |
| GHSA-vp62-88p7-qqf5 | CVE-2026-41567 | Medium | Remove docker/docker (→ moby/moby v29) |
| GHSA-pxq6-2prw-chj9 | CVE-2026-33997 | Medium | Remove docker/docker (→ moby/moby v29) |

## API Migration Map (docker/docker v28 → moby/moby v29)
- Import roots: `github.com/docker/docker/{client,api/types/*}` → `github.com/moby/moby/{client,api/types/*}`.
- Build options/consts moved into the `client` package; `BuilderBuildKit` now lives in `api/types/build`.
- Port types moved from `github.com/docker/go-connections/nat` to `github.com/moby/moby/api/types/network`; `network.Port` is now an opaque struct built via `network.ParsePort(...)`, and `PortBinding.HostIP` is a `net/netip.Addr` (use `netip.IPv4Unspecified()` for `0.0.0.0`).
- `ContainerCreate(ctx, cfg, host, net, platform, name)` → `ContainerCreate(ctx, client.ContainerCreateOptions{...})`.
- `ContainerStart/Stop/Remove`, `ImageTag`, `BuildCachePrune` now return a `Result` struct + error (discard result with `_`).
- `ImageInspectWithRaw(ctx, ref)` → `ImageInspect(ctx, ref)` (single return, embeds `image.InspectResponse`).
- `image.PullOptions{}` → `client.ImagePullOptions{}`; `Ping(ctx)` → `Ping(ctx, client.PingOptions{})`.

## Plan
- [x] Step 1 — Probe the v29 client/api surface for every call site used.
- [x] Step 2 — Swap imports and rewrite call sites in `dockerinternal/` and the two direct calls in `cmd/`.
- [x] Step 3 — `go get` moby modules, `go mod tidy`, `go build ./...`, `go vet ./...`, `go test ./...`.
- [x] Step 4 — Confirm `github.com/docker/docker` removed from module graph; run `govulncheck`.
- [x] Step 5 — CHANGELOG v0.7.6 entry + plan/log docs.
- [ ] Step 6 — Commit on `feature/security-fixes`; open PR to `main` (auto-release tags v0.7.6 on merge).
- [ ] Step 7 — Verify Dependabot alerts close after merge.

## Decisions & Commentary
- Chose full migration to the v29 split modules rather than continuing to document as pending: it is now the only path that removes the vulnerable module and clears the alerts.
- These 5 CVEs are Docker Engine (daemon-side) issues; fortihugorunner is a client and real-world exposure is governed by the user's installed Docker Engine. The module migration is correct dependency hygiene and clears SCA alerts, but is not the runtime mitigation for the daemon itself.
- Did not edit `version/version.go`: the version is injected at build time via ldflags; the default stays `dev`. Release tagging is automatic via `auto-release.yml` on merge to `main` (patch bump → v0.7.6).

## Files Changed
- `dockerinternal/container.go`, `dockerinternal/docker_client.go`, `dockerinternal/watcher.go`
- `cmd/pull_image.go`, `cmd/root.go`
- `go.mod`, `go.sum`
- `CHANGELOG.md`

## Follow-ups
- [ ] `github.com/moby/moby/client` is pre-1.0 (v0.5.0); watch for API changes on future bumps.
- [ ] Confirm Dependabot closes all 5 alerts once merged to `main`.
- [ ] Future `build-image` compatibility work: the current implementation forces BuildKit through the Docker SDK `ImageBuild` API without attaching a BuildKit session, which can fail with `no active sessions` for Dockerfiles using `# syntax=docker/dockerfile:1`. CentralRepo image workflows avoid this by using Docker CLI/Buildx (`docker/setup-buildx-action` + `docker buildx build`), which manages the session automatically. Keep this security release focused on dependency remediation; address `build-image` separately, likely by invoking `docker buildx build --load` for local builds rather than hand-rolling BuildKit session support.

## Risks / Open Questions
- Pre-1.0 client module API stability — pinned to exact versions to limit drift.
- No integration test exercises the live Docker calls; build + vet + govulncheck are the gates. Manual smoke test of `launch-server`/`build-image`/`pull-image` recommended before/after release.

# Session Log: Migrate Docker SDK to Moby v29 client — CVE Remediation
Date: 2026-06-24
Owner: Robert Reris
Related Plan: docs/plans/2026-06-24_Robert-Reris_docker-moby-v29-migration.md

## Milestones
- [x] v29 client/api surface probed for all call sites
- [x] Imports + call sites migrated (dockerinternal/ + cmd/)
- [x] Build + vet + tests pass
- [x] docker/docker removed from module graph; govulncheck clean
- [x] CHANGELOG updated (v0.7.6)
- [ ] PR created + merged
- [ ] Dependabot alerts confirmed closed

## Commentary Stream
### Investigation
- 5 open alerts, all `github.com/docker/docker` at v28.5.2. No fix available on that import path — the `+incompatible` line ends at v28.5.2.
- Found Moby v29 published as split modules on the proxy (tagged 2026-06-18): `github.com/moby/moby/client@v0.5.0`, `github.com/moby/moby/api@v1.55.0`. This is the supported path forward.
- Examined collaborator branches (dev, feature/fix-autorelease-paths, feature/ci-allowlist-paths, feature/ci-path-filters) at the user's request — all CI/docs only, none touch go.mod or dockerinternal/. No remediation hiding there.

### Implementation (2026-06-24)
- Probed exact v29 symbols in a scratch module before editing (ContainerCreateOptions, ImageBuildOptions, ImageInspectResult, network.Port/ParsePort, PortBinding.HostIP=netip.Addr, Ping/PingOptions, etc.).
- `dockerinternal/container.go`: rewrote ImageTag, ImageInspect (was ImageInspectWithRaw), ImagePull, ImageBuild options, BuildCachePrune, ContainerCreate/Start/Stop/Remove/Attach; switched ports to `network` package via `network.ParsePort`; `HostIP` → `netip.IPv4Unspecified()`.
- `dockerinternal/docker_client.go`, `dockerinternal/watcher.go`: import path swap only (`*client.Client` API for NewClientWithOpts/FromEnv/WithAPIVersionNegotiation is unchanged).
- `cmd/pull_image.go`: `ImageTag` → options struct + discard result; added moby/client import.
- `cmd/root.go`: `Ping(ctx)` → `Ping(ctx, client.PingOptions{})`; added moby/client import.
- `go get github.com/moby/moby/client@v0.5.0 github.com/moby/moby/api@v1.55.0`; `go mod tidy`.
- `go build ./...` — clean (after fixing two cmd/ direct calls and the netip.Addr HostIP).
- `go vet ./...` — clean.
- `go test ./...` — fortihugorunner/tests PASS; other packages have no test files.
- `go list -m github.com/docker/docker` — no longer in build graph (REMOVED).
- `govulncheck ./...` — "No vulnerabilities found."

## Dead-ends / Rejected Options
- Bumping `github.com/docker/docker` past v28.5.2: impossible — v29 is not published at that import path.
- `github.com/moby/moby@v29.x` as a single module: not available on the proxy; only the split submodules are.

## Risks & Mitigations
- Risk: pre-1.0 client module (v0.5.0) API churn. Mitigation: pinned exact versions.
- Risk: live Docker calls not covered by automated tests. Mitigation: build + vet + govulncheck; recommend manual smoke test of launch-server / build-image / pull-image.

## Release Notes
- Release is automatic: `auto-release.yml` tags on merge to `main`. Commit type is `security:`/`fix:` → patch bump → v0.7.6. `release.yml` extracts the `## [v0.7.6]` CHANGELOG block for the GitHub release body and builds cross-platform binaries with the version injected via ldflags.

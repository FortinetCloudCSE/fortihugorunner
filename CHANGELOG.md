# Changelog


## [v0.9.0] - 2026-09-28
### Added
- Every command now checks for a newer fortihugorunner release automatically, at most once every 24 hours, and offers to install it before continuing — no need to remember to run `update` yourself. Never blocks the command: a GitHub API failure, an unwritable cache directory, an unparsable dev version, or declining the prompt all fail open and continue on the current version.
- New `--no-auto-update` persistent flag and `FORTIHUGORUNNER_NO_AUTO_UPDATE=1` disable the automatic check entirely; `FORTIHUGORUNNER_AUTO_UPDATE=1` skips the confirmation prompt (for CI / scripted use). Skipped for the `update` and `version` commands themselves.
- On accept, re-execs the updated binary with the exact original arguments so the command you actually ran completes transparently on the new version.
- Verified end-to-end against the real, live GitHub releases (not a mock): declining continues normally with the check still cached; a second run inside the 24h window makes no network call; `--no-auto-update` makes no network call and writes no cache; `FORTIHUGORUNNER_AUTO_UPDATE=1` performed a real download-and-install and re-exec. See `docs/plans/0005_2026-09-28_Jeff-Kopko_auto-update.md`.

## [v0.8.0] - 2026-09-28
### Added
- New `install-docker` command: installs a free, lightweight, Docker-API-compatible engine when none is found — Docker Engine CE on Linux (a pinned, checksum-verified `get.docker.com`), Colima on macOS (via Homebrew), and Docker Engine CE inside WSL2 on Windows. Deliberately never installs Docker Desktop, whose free-use terms exclude larger organizations. `--dry-run` prints exactly what would run; `--yes` (or `FORTIHUGORUNNER_AUTO_INSTALL_DOCKER=1`) skips the confirmation prompt.
- Any command now offers to run `install-docker` automatically when no Docker-compatible engine is found at all (not when one exists but isn't running — that keeps today's plain-error behavior). Opt out entirely with the new persistent `--no-install-docker` flag.
- Verified end-to-end on Linux in an isolated, throwaway privileged container (install, the no-systemd `dockerd` fallback, idempotency, `--dry-run`, declining the prompt) and on real Windows via two disposable Azure VMs (the binary, `--no-install-docker`, `--dry-run`, and the auto-confirm env var all behave correctly; confirmed WSL2 itself can't be driven from Azure's SYSTEM-context automation, a platform limitation, not a bug here). macOS/Colima is implemented and cross-compiled; real Apple hardware verification is tracked separately — see `docs/plans/0004_2026-09-28_Jeff-Kopko_auto-install-docker.md`.

## [v0.7.7] - 2026-09-28
### Fixed
- `launch-server --pull-latest` compared the wrong local image reference when deciding whether to pull and retag: it inspected the registry-qualified tag (`public.ecr.aws/.../fortinet-hugo:latest`), which any prior pull under that full name keeps fresh, instead of the short tag (`fortinet-hugo:latest`) that actually gets started. A stale short tag could read as "already up to date" and never get retagged. `getLocalRepoDigest` now inspects the short tag while still matching its `RepoDigests` against the registry-qualified prefix (Docker stores digests per image object, keyed by the full name, regardless of which local tag you inspect with).
- `launch-server`'s `--pull-latest` freshness check used its own hardcoded copy of the known-image list and registry (`fortinet-hugo`/`hugotester`, `public.ecr.aws/k4n6m5h8/`), separate from `pull-image`/`build-image`'s maps, and silently skipped the check for any other `--docker-image` name with no indication why. The three commands now share one map (`dockerinternal.EnvToTarget` / `TargetToImageName`), `launch-server` takes a `--registry` flag matching `pull-image`'s, and an unrecognized `--docker-image` now prints why the freshness check was skipped instead of skipping silently.

## [v0.7.6] - 2026-06-24
### Security
- Migrated the Docker SDK off the frozen `github.com/docker/docker` module (permanently capped at v28.5.2 under its `+incompatible` versioning) onto the restructured Moby v29 client modules — `github.com/moby/moby/client` v0.5.0 and `github.com/moby/moby/api` v1.55.0. This removes `github.com/docker/docker` from the dependency graph entirely, closing all 5 remaining open Dependabot alerts (including the three documented as "upstream patch pending" in v0.7.5):
  - GHSA-rg2x-37c3-w2rh / CVE-2026-42306 (High) — `docker cp` bind-mount redirection to host path
  - GHSA-x86f-5xw2-fm2r / CVE-2026-41568 (High) — `PUT /containers/{id}/archive` executes container binary on host
  - GHSA-x744-4wpc-v9h2 (High) — Moby AuthZ plugin bypass via oversized request bodies
  - GHSA-vp62-88p7-qqf5 / CVE-2026-41567 (Medium) — `docker cp` arbitrary empty-file creation via symlink swap
  - GHSA-pxq6-2prw-chj9 / CVE-2026-33997 (Medium) — off-by-one in plugin privilege validation
- `govulncheck ./...` reports no known vulnerabilities after the migration.

## [v0.7.5] - 2026-06-10
### Security
- Upgraded `go.opentelemetry.io/otel` family to v1.43.0 (fixes CVE-2026-39883, CVE-2026-24051, CVE-2026-39882)
- Upgraded `github.com/ulikunitz/xz` to v0.5.15 (fixes CVE-2025-58058)
- Upgraded `golang.org/x/crypto` to v0.53.0 (fixes CVE-2025-47914, CVE-2025-58181)
- Upgraded `github.com/docker/docker` to v28.5.2 (partially addresses CVE-2026-34040, CVE-2026-33997)
- **Upstream patch pending** — the following docker/docker CVEs have no patch available yet; docker/docker v29.x is not yet published to the Go module proxy:
  - GHSA-rg2x-37c3-w2rh / CVE-2026-42306 (High)
  - GHSA-x86f-5xw2-fm2r / CVE-2026-41568 (High)
  - GHSA-vp62-88p7-qqf5 / CVE-2026-41567 (Medium)

## [v0.7.4] - 2025-11-10
### Added
- Added Docker context awareness so fortihugorunner uses the same daemon/socket as the Docker CLI, removing the need to export DOCKER_HOST for nonstandard environments. 

## [v0.7.3] - 2025-08-06
### Security
- Updated dependencies to address security vulnerabilities
  - golang.org/x/oauth2 updated to v0.27.0 (fixes CVE-2025-22868)

## [v0.7.2] - 2025-08-06
### Changed
- `pull-latest` parameter added to `launch-server` component to pull latest Docker image before running container

## [v0.7.1] - 2025-07-01
### Added
- `pull-image` command: allows users to pull latest prebuilt fortinet-hugo and hugotester images from our public ECR repositories.

## [v0.6.1] - 2025-06-10
### Security
- Updated dependencies to address security vulnerabilities:
  - golang.org/x/crypto updated to v0.39.0 (fixes CVE-2025-22869)
  - golang.org/x/net updated to v0.41.0 (fixes CVE-2025-22872 and CVE-2025-22870) 

## [v0.6.0] - 2025-06-05
### Changed
- `update` command now renames binary to `fortihugorunner` or `fortihugorunner.exe` prior to updating

## [v0.5.0] - 2025-05-30
### Added
- `update` command: allows users to self-update the current binary to the latest release.
- `rename` command: enables 'trimming' the platform information from the executable filename.

### Changed
- Enhanced `-v` (version) output to include platform (OS/arch) information.

## [v0.4.2] - 2025-05-29
### Changed
- Renamed project from 'docker-run-go' to 'fortihugorunner'.
- Go module path changed from 'github.com/FortinetCloudCSE/docker-run-go' to 'github.com/FortinetCloudCSE/fortihugorunner'.

## [v0.3.2] - 2025-05-13
### Changed
- Changed default `--hugo-version` parameter for build-image command to std
- Updated help (-h) examples for build-image command

### Removed
- Removed create-content placeholder component
- Removed completion component

## [v0.3.1] - 2025-04-24
### Fixed
- Changed flag `--mount-hugo` to `--mount-toml` in launch-server command
- Removed auto-update flags in 'hugo server' wrapper

## [v0.3.0] - 2025-04-23
### Added
- `--mount-hugo` flag in launch-server command to specify hugo.toml mount behavior
- Logic to retrieve CentralRepo branch directly from Dockerfile
- `--hugo-version` flag in build-image command to specify Hugo version

## [v0.2.0] - 2025-03-20
### Added
- `--version` flag to check the current CLI version
- Runtime check for Docker daemon availability

## [v0.1.0] - 2025-03-07
### Added
- Initial release with core features:
  - build administrative and development Docker images
  - launch a Hugo server container for local workshop development
- Support for Windows, MacOs, and Linux architectures

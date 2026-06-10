# Changelog


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

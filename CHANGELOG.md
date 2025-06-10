# Changelog

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

## FortiHugoRunner: A FortinetCloudCSE Docker development helper

FortiHugoRunner is a command-line tool that manages Hugo workshop development containers. It wraps the Docker SDK to pull, build, and run the Fortinet Hugo images without requiring you to write `docker` commands directly.

---

## Table of Contents

- [Download and Install](#download-and-install)
- [Commands](#commands)
  - [rename](#rename)
  - [version](#version)
  - [pull-image](#pull-image)
  - [build-image](#build-image)
  - [launch-server](#launch-server)
  - [update](#update)
- [Typical Workflow](#typical-workflow)
- [Build from Source](#build-from-source)
- [Contributing](#contributing)

---

## Download and Install

**Prerequisites:** Docker must be installed and running (Rancher Desktop, Docker Desktop, Colima, etc.).

> The tool reads `~/.docker/config.json` and honors the active Docker context automatically. Set `DOCKER_CONTEXT=<name>` or `DOCKER_HOST=…` to override.

> **Keep your Docker Engine up to date.** FortiHugoRunner is only a *client* that talks to your local Docker daemon. Many Docker security advisories are fixed in the Docker Engine itself, not in this tool — so upgrading FortiHugoRunner does not patch your daemon. Run a current, supported version of Docker Engine / Docker Desktop and apply its updates to stay protected.

Navigate to the [releases](https://github.com/FortinetCloudCSE/fortihugorunner/releases) page, right-click the binary for your OS/architecture, and click **Save Link As**.

#### Determine your architecture

**Windows** — from Command Prompt:
```
echo %PROCESSOR_ARCHITECTURE%
```

| Output | Download |
|--------|----------|
| AMD64  | `fortihugorunner-windows-amd64.exe` |
| x86    | `fortihugorunner-windows-386.exe` |

**macOS / Linux** — from terminal:
```
uname -m
```

| Output  | Download |
|---------|----------|
| x86_64  | `fortihugorunner-darwin-amd64` or `fortihugorunner-linux-amd64` |
| arm64   | `fortihugorunner-darwin-arm64` or `fortihugorunner-linux-arm64` |

After downloading, make the binary executable (macOS/Linux) and optionally move it onto your `$PATH`:
```bash
chmod +x fortihugorunner-linux-amd64
mv fortihugorunner-linux-amd64 /usr/local/bin/fortihugorunner
```

> **All examples below assume `rename` has been run** (so the binary is just `fortihugorunner`). On Windows, substitute `fortihugorunner.exe` for `fortihugorunner`.

---

## Commands

### rename

Strips the OS/architecture suffix from the downloaded binary filename so subsequent commands are shorter.

```bash
./fortihugorunner-linux-amd64 rename
# binary is now named: fortihugorunner
```

---

### version

Prints the current version, build date, and OS/architecture platform.

```bash
fortihugorunner version
# Version: v0.7.5
# Date:    2026-06-10
# Platform: darwin/arm64

fortihugorunner -v   # shorthand flag, same output
```

---

### pull-image

Pulls the latest prebuilt Hugo development image from the Fortinet public ECR registry and retags it locally.

```bash
fortihugorunner pull-image --env author-dev   # pulls fortinet-hugo:latest (default)
fortihugorunner pull-image --env admin-dev    # pulls hugotester:latest
```

| Flag | Default | Description |
|------|---------|-------------|
| `--env` | `author-dev` | `author-dev` → `fortinet-hugo` image; `admin-dev` → `hugotester` image |
| `--registry` | `public.ecr.aws/k4n6m5h8/` | ECR registry prefix |

Public image URIs:
```
public.ecr.aws/k4n6m5h8/fortinet-hugo:latest
public.ecr.aws/k4n6m5h8/hugotester:latest
```

---

### build-image

Builds a Hugo Docker image locally from the Dockerfile in your current directory.

```bash
fortihugorunner build-image --env author-dev              # builds fortinet-hugo:latest
fortihugorunner build-image --env admin-dev               # builds hugotester:latest
fortihugorunner build-image --env author-dev --hugo-version 0.146.0
```

| Flag | Default | Description |
|------|---------|-------------|
| `--env` | `author-dev` | `author-dev` → production image (`fortinet-hugo`); `admin-dev` → dev/test image (`hugotester`) |
| `--hugo-version` | `std` | Hugo base image version tag (must match the `hugomods/hugo` tag in your Dockerfile) |

> Use `pull-image` for most workflows. `build-image` is only needed when customizing the Dockerfile locally.

---

### launch-server

Starts a Hugo development server container, mounts your workshop directory into it, and streams container logs to your terminal. Ctrl-C stops the container cleanly.

```bash
fortihugorunner launch-server \
    --docker-image fortinet-hugo:latest \
    --host-port 1313 \
    --container-port 1313 \
    --watch-dir . \
    --mount-toml

# Pull the latest image before starting (useful in CI or after a long break):
fortihugorunner launch-server \
    --docker-image fortinet-hugo:latest \
    --host-port 1313 \
    --container-port 1313 \
    --watch-dir . \
    --mount-toml \
    --pull-latest
```

| Flag | Default | Description |
|------|---------|-------------|
| `--docker-image` | — | Image name and tag to run (e.g. `fortinet-hugo:latest`) |
| `--host-port` | — | Host port to bind (e.g. `1313`) |
| `--container-port` | — | Container port to expose (e.g. `1313`) |
| `--watch-dir` | — | Path to the workshop directory to mount into the container |
| `--mount-toml` | `false` | Mount `hugo.toml` from `--watch-dir` into the container |
| `--pull-latest` | `false` | Pull the latest version of `--docker-image` before starting |

Once running, open `http://localhost:<host-port>` in your browser. The server reloads automatically when files in `--watch-dir` change.

---

### update

Updates the `fortihugorunner` binary in place to the latest GitHub release. If the binary filename includes an OS/architecture suffix, it will be renamed first automatically.

```bash
fortihugorunner update
```

---

## Typical Workflow

```bash
# 1. Download and rename the binary (one-time)
./fortihugorunner-darwin-arm64 rename

# 2. Pull the latest prebuilt image
fortihugorunner pull-image --env author-dev

# 3. Launch the server from your workshop directory
cd ~/my-workshop
fortihugorunner launch-server \
    --docker-image fortinet-hugo:latest \
    --host-port 1313 \
    --container-port 1313 \
    --watch-dir . \
    --mount-toml

# 4. Open http://localhost:1313 — edits reload automatically

# 5. Keep the tool up to date
fortihugorunner update
```

---

## Build from Source

**Prerequisites:** Go 1.23+ and Docker.

```bash
git clone https://github.com/FortinetCloudCSE/fortihugorunner.git
cd fortihugorunner
go mod download
```

Build for your current platform:
```bash
go build -o fortihugorunner .
chmod +x fortihugorunner
```

Cross-compile:
```bash
# Linux amd64
GOOS=linux GOARCH=amd64 go build -o fortihugorunner-linux-amd64 .

# macOS arm64
GOOS=darwin GOARCH=arm64 go build -o fortihugorunner-darwin-arm64 .

# Windows amd64
GOOS=windows GOARCH=amd64 go build -o fortihugorunner-windows-amd64.exe .
```

Run tests:
```bash
go vet ./...
go clean -testcache && go test ./...
```

Other useful commands:
```bash
go fmt ./...          # format code
go get <package>      # add a dependency
go mod tidy           # prune unused deps
go get -u ./...       # update all deps to latest
```

---

## Contributing

### Branch model

```
feature/<name>  →  PR to dev  →  merge to dev
dev             →  PR to main →  merge to main  →  auto-release
```

- All changes go through a PR. Direct pushes to `dev` and `main` are blocked.
- PRs to `dev` require the **Build and Test** CI check to pass + 1 approving review.
- PRs to `main` require the **Build and Release Binaries** CI check to pass + 1 approving review.

### Automatic versioning

Releases are created automatically when a PR is merged to `main`. The version number is derived from commit message prefixes following [Conventional Commits](https://www.conventionalcommits.org/):

| Commit prefix | Version bump |
|---------------|-------------|
| `feat:` | minor (v0.7.x → v0.8.0) |
| `fix:`, `chore:`, `docs:`, `ci:`, etc. | patch (v0.7.4 → v0.7.5) |
| `BREAKING CHANGE:` or `feat!:` | major (v0.7.x → v1.0.0) |

Example commit messages:
```
feat: add --timeout flag to launch-server
fix: correct path separator on Windows
chore(deps): upgrade golang.org/x/crypto to v0.53.0
feat!: remove deprecated build-image --legacy flag
```

You do not need to edit `version/version.go` — the version is injected at build time by CI via `-ldflags`.

### CI workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `test.yml` | PR to `dev` or `main` | Build + vet + test gate |
| `auto-release.yml` | Push to `main` | Compute next semver from commits, create tag |
| `release.yml` | New `v*` tag | Build 6 platform binaries, publish GitHub release with CHANGELOG notes |
| `pr-checklist.yml` | PR opened | Posts a review checklist comment |

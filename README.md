## FortiHugoRunner: A FortinetCloudCSE Docker development helper

### Download Instructions
---

**Prereqs**:

- Docker installed (via Rancher Desktop, for example) and running

Navigate to the [releases](https://github.com/FortinetCloudCSE/fortihugorunner/releases) page and right click the binary for your OS/Architecture, click "Save Link As...", and choose your preferred download location. To determine your architecture, follow the steps in the next section.

#### Windows

From the Windows command prompt, run:

```bash
C:\\echo %PROCESSOR_ARCHITECTURE%
```

| If your command output is:               | then download:                         |
|------------------------------------------|----------------------------------------|
| AMD64                                    | fortihugorunner-windows-amd64.exe      |
| x86                                      | fortihugorunner-windows-386.exe        |

#### MacOS/Linux

From your terminal, run:

```
uname -m
```

| If your command output is:        | then download:                           |
|-----------------------------------|------------------------------------------|
| x86_64                            | fortihugorunner-<darwin/linux>-amd64.exe|
| x86                               | fortihugorunner-<darwin/linux>-386.exe  |


Then, you can either copy the binary from that directory into your system path or to the directory containing your workshop.

To get your system path:

- In bash (linux or mac):
```
> echo $PATH 
```

- In windows:
```
C:\\echo %PATH%
```
---

### Building the Hugo Server Image

In Linux or MacOS:

```bash
> ./fortihugorunner-<OS>-<arch> rename              #trims the OS/arch from the executable for simplicity

> ./fortihugorunner build-image --env admin-dev     #builds an image for testing (hugotester:latest)

> ./fortihugorunner build-image --env author-dev    #builds an image for workshop authoring (fortinet-hugo:latest)
```

**Note: If binary is located in your system path, omit './' when running the commands throughout this document.**

In Windows:

```bash
C:\\fortihugorunner-windows-<arch>.exe rename           #trims the OS/arch from the executable for simplicity

C:\\fortihugorunner.exe build-image --env admin-dev     #builds an image for testing (hugotester:latest)

C:\\fortihugorunner.exe build-image --env author-dev    #builds an image for workshop authoring (fortinet-hugo:latest)
```
---

### Prebuilt Docker Images

We have prebuilt fortinet-hugo and hugotester Docker images available for download from the following public repositories:

```
public.ecr.aws/k4n6m5h8/fortinet-hugo
public.ecr.aws/k4n6m5h8/hugotester
```

These can be pulled with fortihugorunner via:

```
> ./fortihugorunner pull-image --env author-dev     # default, pulls fortinet-hugo image
> ./fortihugorunner pull-image --env admin-dev      # pulls hugotester image
```

---

### Launching a Hugo Server Container

In Linux or MacOS:

```
> ./fortihugorunner launch-server \
      --docker-image fortinet-hugo:latest \
      --host-port 1313 \
      --container-port 1313 \
      --watch-dir . \
      --mount-toml
```

In Windows:

```
C:\\fortihugorunner.exe launch-server \
      --docker-image fortinet-hugo:latest \
      --host-port 1313 \
      --container-port 1313 \
      --watch-dir . \
      --mount-toml
```

---

### Help Menu

On Linux or MacOS:

```
> ./fortihugorunner -h
> ./fortihugorunner launch-server -h
```

On Windows:

```
C:\\fortihugorunner.exe -h
```

---

### Update the Binary to the Latest Version

On Linux or MacOS:

```
> ./fortihugorunner update
```

On Windows:

```
C:\\fortihugorunner.exe update
```
---

## Build Instructions

**Prereqs**:

- Docker installed (via Rancher Desktop, for example)
- Recent [Golang](https://go.dev/) version installed.
  - Installation information [here](https://go.dev/doc/install).

1. Clone the repository

```
> git clone <HTTPS/SSH URL found in the 'Code' dropdown above>

> cd fortihugorunner
```

2. Download necessary go libraries

```
> go mod download
```

3. Build

   Note: Before building, you can confirm availability of the desired OS/Architecture via:
```
> go tool dist list
``` 

- **Linux/x86_64:**
```
> GOOS=linux GOARCH=amd64 go build -o fortihugorunner .
```
- **macOS/AMD64:**
```
> GOOS=darwin GOARCH=amd64 go build -o fortihugorunner .
```
- **Windows/x86_64:**
```
> GOOS=windows GOARCH=amd64 go build -o fortihugorunner.exe .

```

4. Update executable permissions if needed
```
> chmod +x fortihugorunner
```

5. (optional) Copy the executable into a directory in the system path. To list the path, run:

- From a Linux or Mac terminal:
```
> echo $PATH 
```

- At the Windows Command Prompt:
```
> echo %PATH%
```

---

## Some Useful Pre-build Commands

After adding a new dependency, run:

```
> go get <package>
> go mod tidy
```

To update all packages to their latest versions:

```
> go get -u ./...
```

Formatting:

```
> go fmt ./...
```

Various function checks:

```
> go vet ./...
```

Run all unit tests:

```
> go clean -testcache
> go test ./..
```

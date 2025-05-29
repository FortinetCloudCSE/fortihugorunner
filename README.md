## FortinetCloudCSE Docker Development Helper

## To run the tool via pre-compiled binary:

**Prereqs**:

- Docker installed (via Rancher Desktop, for example) and running

Navigate to the [releases](https://github.com/FortinetCloudCSE/fortihugorunner/releases) page and right click the binary for your OS/Architecture, click "Save Link As...", and choose your preferred download location.

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

## Building the Hugo Server Image

In Linux or MacOS:

```bash
> mv fortihugorunner-<OS>-<arch> fortihugorunner

> ./fortihugorunner build-image --env admin-dev     #builds an image for testing (hugotester:latest)

> ./fortihugorunner build-image --env author-dev    #builds an image for workshop authoring (fortinet-hugo:latest)
```

**Note: If binary is located in your system path, omit './' when running the commands throughout this document.**

In Windows:

```bash
C:\\move fortihugorunner-windows-<arch>.exe fortihugorunner.exe

C:\\fortihugorunner.exe build-image --env admin-dev     #builds an image for testing (hugotester:latest)

C:\\fortihugorunner.exe build-image --env author-dev    #builds an image for workshop authoring (fortinet-hugo:latest)
```


## Launching a Hugo Server Container

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

To see all other commands or get help:

On Linux or MacOS:

```
> ./fortihugorunner -h
> ./fortihugorunner launch-server -h
```

On Windows:

```
C:\\fortihugorunner.exe -h
```

## To build the CLI tool:

**Prereqs**:

- Docker installed (via Rancher Desktop, for example)
- Recent [Golang](https://go.dev/) version installed.
  - Installation information [here](https://go.dev/doc/install).

*Clone the repository*
```
> git clone <HTTPS/SSH URL found in the 'Code' dropdown above>

> cd fortihugorunner
```

*Download necessary go libraries:*
```
> go mod download
```

*Build:*

**Note: Before building, you can confirm availability of the desired OS/Architecture via:**
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

*Update executable permissions if needed:*
```
> chmod +x fortihugorunner
```

*Copy the executable into a directory in the system path. To list the path, run:*

- In bash (linux or mac):
```
> echo $PATH 
```

- In windows:
```
> echo %PATH%
```

## Some Useful Go Commands

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

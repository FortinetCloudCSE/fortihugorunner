## FortinetCloudCSE Docker Development Helper

## To run the tool via pre-compiled binary:

**Prereqs**:

- Docker installed (via Rancher Desktop, for example)
- Workshop Docker image built

Navigate to the *binaries* folder above, click the binary for your OS/Architecture, and click on the **download raw file** icon at the top right of the screen. 

Then, you can either copy the binary into your system path or to the local directory containing your workshop.

To get your system path:

- In bash (linux or mac):
```
echo $PATH 
```

- In windows:
```
echo %PATH%
```
## Building the Hugo Server Image

In Linux or MacOS:

```bash
./docker-run-go build-image admin-dev     #builds an image for testing (hugotester:latest)

./docker-run-go build-image author-dev    #builds an image for workshop authoring (fortinet-hugo:latest)
```

In Windows:

```bash
C:\\docker-run-go.exe build-image admin-dev     #builds an image for testing (hugotester:latest)

C:\\docker-run-go.exe build-image author-dev    #builds an image for workshop authoring (fortinet-hugo:latest)
```


## Launching a Hugo Server Container

In Linux or MacOS:

```
> mv docker-run-go-<OS>-<arch> docker-run-go

> ./docker-run-go launch-server \
      --docker-image fortinet-hugo:latest \
      --host-port 1313 \
      --container-port 1313 \
      --watch-dir .
```

In Windows:

```
C:\move docker-run-go-windows-<arch>.exe docker-run-go.exe

C:\docker-run-go.exe launch-server \
      --docker-image fortinet-hugo:latest \
      --host-port 1313 \
      --container-port 1313 \
      --watch-dir .
```

To see all other commands or get help:

On Linux or MacOS:

```
> ./docker-run-go -h
> ./docker-run-go launch-server -h
```

On Windows:

```
C:\docker-run-go.exe -h
```

## To build the CLI tool:

**Prereqs**:

- Docker installed (via Rancher Desktop, for example)
- Go installed (not needed if you just want to run the compiled binary)
  - For instructions on installing Go, head here: https://go.dev/doc/install
- Workshop Docker image built

*Download necessary go libraries:*
```
go mod download
```

*Build:*

**Note: Before building, you can confirm availability of the desired OS/Architecture via:**
```
go tool dist list
``` 

- **Linux/x86_64:**
```
GOOS=linux GOARCH=amd64 go build -o docker_run .
```
- **macOS/AMD64:**
```
GOOS=darwin GOARCH=amd64 go build -o docker_run .
```
- **Windows/x86_64:**
```
GOOS=windows GOARCH=amd64 go build -o docker_run.exe .

```

*Update executable permissions if needed:*
```
chmod +x docker_run
```

*Copy the executable into a directory in the system path. To list the path, run:*

- In bash (linux or mac):
```
echo $PATH 
```

- In windows:
```
echo %PATH%
```

## Some Useful Go Commands

After adding a new dependency, run:

```
go get <package>
go mod tidy
```

To update all packages to their latest versions:

```
go get -u ./...
```

Formatting:

```
go fmt ./...
```

Various function checks:

```
go vet ./...
```

Run all unit tests:

```
go clean -testcache
go test ./..
```

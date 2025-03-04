#!/bin/bash
GOOS=linux GOARCH=amd64 go build -o binaries/docker-run-go-linux-amd64 . && echo "Linux AMD built"
GOOS=linux GOARCH=arm64 go build -o binaries/docker-run-go-linux-arm64 . && echo "Linux ARM built"
GOOS=darwin GOARCH=amd64 go build -o binaries/docker-run-go-mac-amd64 . && echo "Mac AMD built"
GOOS=darwin GOARCH=arm64 go build -o binaries/docker-run-go-mac-arm64 . && echo "Mac ARM built"
GOOS=windows GOARCH=amd64 go build -o binaries/docker-run-go-windows-amd64.exe . && echo "Windows AMD built"
GOOS=windows GOARCH=arm64 go build -o binaries/docker-run-go-windows-arm64.exe . && echo "Windows ARM built"

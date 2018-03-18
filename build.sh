#/usr/bin/env bash

# Build for WINDOWS
env GOOS=windows GOARCH=amd64 go build
mv goperf.exe binaries/windows/amd64

env GOOS=windows GOARCH=386 go build
mv goperf.exe binaries/windows/386

# Build for FreeBSD
env GOOS=freebsd GOARCH=amd64 go build
mv goperf binaries/freebsd/amd64

env GOOS=freebsd GOARCH=386 go build
mv goperf binaries/freebsd/386/

# Build for OSX/Darwin
env GOOS=darwin GOARCH=386 go build
mv goperf binaries/darwin/386/

env GOOS=darwin GOARCH=amd64 go build
mv goperf binaries/darwin/amd64/

# Build for Linux
env GOOS=linux GOARCH=386 go build
mv goperf binaries/linux/386/

env GOOS=linux GOARCH=amd64 go build
mv goperf binaries/linux/amd64/

# Build for current platform
go build

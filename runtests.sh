#!/bin/bash

set -eufx -o pipefail

# Install compiled versions of everything for both normal and race detector
go test -i ./...
go test -race -i ./...
# Run tests with the race detector
go test -race -v ./...

go vet ./...

# Ensure code is formatted print file names to stderr if not and fail with an error code
test -z "$(gofmt -l -w . | tee /dev/stderr)"

# Ensure no generated files changed contents
test -z "$(git status --short | tee /dev/stderr)"

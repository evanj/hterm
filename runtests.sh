#!/bin/bash

set -euf -o pipefail

# Install compiled versions of everything for both normal and race detector
echo "compiling go ..."
go test -i ./...
go test -race -i ./...

# Run tests with the race detector
echo "running go tests ..."
go test -race -v ./...

echo "running go vet ..."
go vet ./...

# Ensure code is formatted print file names to stderr if not and fail with an error code
echo "checking that go is formatted ..."
test -z "$(gofmt -l -w . | tee /dev/stderr)"

# Ensure no generated files changed contents
echo "checking that generated files are up to date in git ..."
CHANGED=`git status --short | tee /dev/stderr`
test -z "${CHANGED}" || (printf "Generated files differ from git:\n${CHANGED}\n" && false)

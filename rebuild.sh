#!/bin/bash
# sets up the development environment and re-generates any generated files

set -euf -o pipefail

echo "go get-ting dependencies"
go get ./...

echo "generating Makefile"
go run genmakefile/deps.go genmakefile/genmakefile.go > Makefile

# TODO: Move this into the Makefile?
echo "Installing node dependencies with npm install"
npm install

echo "running make"
make

echo "running go get esc / go generate"
go get github.com/mjibson/esc
go generate -x ./...

echo "NOTE: 'rm -rf build node_modules' to be really clean"

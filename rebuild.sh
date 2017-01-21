#!/bin/bash

set -euf -o pipefail

echo "generating Makefile"
go run genmakefile/deps.go genmakefile/genmakefile.go > Makefile

# TODO: Move this into the Makefile?
echo "Installing node dependencies with npm install"
npm install

echo "running make"
make

echo "running go generate"
go generate -x ./...

echo "NOTE: 'rm -rf build node_modules' to be really clean"

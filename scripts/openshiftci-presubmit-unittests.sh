#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export PATH=$PATH:$GOPATH/bin
# CI go-toolset is still on Go 1.26; allow downloading the go.mod toolchain (1.27+).
export GOTOOLCHAIN=auto

go env
go mod vendor
if [[ $(go fmt `go list ./... | grep -v vendor`) ]]; then
    echo "not well formatted sources are found"
    exit 1
fi
go version
go mod tidy
if [[ ! -z $(git status -s) ]]
then
    echo "Go mod state is not clean."
    exit 1
fi

# Unit tests to be referenced here
make test

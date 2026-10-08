#!/bin/sh

# fail if some commands fails
set -e

# Do not show token in CI log
set +x

# show commands
set -x
export CI="prow"
# CI go-toolset is still on Go 1.26; allow downloading the go.mod toolchain (1.27+).
export GOTOOLCHAIN=auto
go mod vendor

export PATH="$PATH:$(pwd)"

# Reference e2e test(s)
echo "Please reference the E2E test script(s) here"

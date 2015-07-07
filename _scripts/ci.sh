#!/bin/bash

set -e

THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "${THIS_SCRIPT_DIR}/.."

export PATH="$PATH:$GOPATH/bin"

#
# Script for Continuous Integration
#

set -v

# Install depenecies
go get -u github.com/tools/godep
go install github.com/tools/godep
godep restore

# Intsall stepman
godep go install

# Check for unhandled errors
go get -u github.com/kisielk/errcheck
go install github.com/kisielk/errcheck

# Go lint
go get -u github.com/golang/lint/golint

bash "${THIS_SCRIPT_DIR}/test.sh"

#
# ==> DONE - OK
#

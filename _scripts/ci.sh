#!/bin/bash

set -e

THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export REPO_ROOT_DIR="${THIS_SCRIPT_DIR}/.."

set -v

# Install dependencies
go get -u github.com/tools/godep
go install github.com/tools/godep
godep restore

# Build a test version
go build -o tmpbin
./tmpbin setup
rm ./tmpbin

bash "${THIS_SCRIPT_DIR}/common/ci.sh"

# ===> DONE

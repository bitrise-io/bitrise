#!/bin/bash

set -e

THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export REPO_ROOT_DIR="${THIS_SCRIPT_DIR}/.."

set -v

bash "${THIS_SCRIPT_DIR}/common/ci.sh"

# TODO:
#  do a `go build` and run a couple of test commands with it

go build -o tmpbin
./tmpbin setup
rm ./tmpbin

# ===> DONE

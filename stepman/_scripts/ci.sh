#!/bin/bash

set -e

THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export REPO_ROOT_DIR="${THIS_SCRIPT_DIR}/.."
bash "${THIS_SCRIPT_DIR}/common/ci.sh"

set -v

cd "${THIS_SCRIPT_DIR}/.."
go build -o _tmp/stepman
_tmp/stepman setup -c https://github.com/bitrise-io/bitrise-steplib.git
_tmp/stepman delete -c https://github.com/bitrise-io/bitrise-steplib.git
rm -rf _tmp/stepman

#
# ==> DONE -OK
#

#!/bin/bash

set -e

THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export REPO_ROOT_DIR="${THIS_SCRIPT_DIR}/.."
cd "${REPO_ROOT_DIR}"

if [ ! -d "${ENVMAN_REPO_DIR_PATH}" ] ; then
    echo "[!] ENVMAN_REPO_DIR_PATH not defined or not a dir - required!"
    exit 1
fi

if [ ! -d "${STEPMAN_REPO_DIR_PATH}" ] ; then
    echo "[!] STEPMAN_REPO_DIR_PATH not defined or not a dir - required!"
    exit 1
fi

set -v

# go install envman
cd "${ENVMAN_REPO_DIR_PATH}"
godep restore
go install

# go install stepman
cd "${STEPMAN_REPO_DIR_PATH}"
godep restore
go install

# godep restore for bitrise-cli
cd "${REPO_ROOT_DIR}"
godep restore

# => DONE [OK]

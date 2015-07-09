#!/bin/bash

set -e

THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export REPO_ROOT_DIR="${THIS_SCRIPT_DIR}/.."
cd "${REPO_ROOT_DIR}"

CONFIG_tool_bin_path="${REPO_ROOT_DIR}/_temp/bin"
echo " (i) CONFIG_tool_bin_path: ${CONFIG_tool_bin_path}"

if [ ! -d "${ENVMAN_REPO_DIR_PATH}" ] ; then
    echo "[!] ENVMAN_REPO_DIR_PATH not defined - required!"
    exit 1
fi

if [ ! -d "${STEPMAN_REPO_DIR_PATH}" ] ; then
    echo "[!] STEPMAN_REPO_DIR_PATH not defined - required!"
    exit 1
fi

set -v

mkdir -p "${CONFIG_tool_bin_path}"

# build envman
cd "${ENVMAN_REPO_DIR_PATH}"
docker-compose run --rm app go build -o bin-envman
mv ./bin-envman "${CONFIG_tool_bin_path}/envman"

# build stepman
cd "${STEPMAN_REPO_DIR_PATH}"
docker-compose run --rm app go build -o bin-stepman
mv ./bin-stepman "${CONFIG_tool_bin_path}/stepman"

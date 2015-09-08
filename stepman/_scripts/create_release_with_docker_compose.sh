#!/bin/bash

set -e

THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export REPO_ROOT_DIR="${THIS_SCRIPT_DIR}/.."
cd "${REPO_ROOT_DIR}"

set -v

docker-compose build --no-cache app

docker-compose run --rm app bash ./_scripts/install_bitrise_cli.sh
docker-compose run --rm app bitrise run create-release

#!/bin/bash

set -e
set -v

docker-compose build

docker-compose run --rm app /bin/bash _scripts/ci.sh

#
# CI DONE [OK]
#

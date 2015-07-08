#!/bin/bash

set -e
set -v

sudo docker-compose run --rm app /bin/bash _scripts/ci.sh

#
# CI DONE [OK]
#

#!/bin/bash

set -e
set -v

#
# Cleanup docker images which are not tagged and not used
#  as recommended in the official docs: https://docs.docker.com/reference/commandline/cli/
#  (except we don't use sudo)
docker rmi $(docker images -f "dangling=true" -q)

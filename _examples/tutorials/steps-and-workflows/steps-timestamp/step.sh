#!/bin/bash

set -e

THIS_SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Start go program
cd "${THIS_SCRIPTDIR}"

set -v

go run ./step.go

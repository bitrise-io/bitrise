#!/bin/bash

set -e
set -v

errcheck -asserts=true -blank=true $(go list ./...)

go test -v ./...

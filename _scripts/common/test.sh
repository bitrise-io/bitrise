#!/bin/bash

set -e
set -v

errcheck -asserts=true -blank=true $(go list ./...)


# ==> TEST
go test -v ./...


# ==> LINT
golint ./...

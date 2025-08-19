#! /bin/bash

HOOK_DIR=$(dirname "$INTEGRATION_TEST_BINARY_PATH")/hooks
mkdir -p "$HOOK_DIR"

touch "$HOOK_DIR"/build_start

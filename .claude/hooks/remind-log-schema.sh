#!/usr/bin/env bash
set -euo pipefail

command -v jq >/dev/null || exit 0

SCHEMA_FILE_PATTERN='(^|/)log/(models\.go|corelog/[^/]+\.go)$'
TEST_FILE_PATTERN='_test\.go$'

file_path=$(jq -r '.tool_input.file_path // empty' 2>/dev/null) || exit 0

if [[ "$file_path" =~ $SCHEMA_FILE_PATTERN ]] && [[ ! "$file_path" =~ $TEST_FILE_PATTERN ]]; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      additionalContext: "Reminder: this file defines the structured log schema consumed by downstream systems. Changing field names, types, or removing/renaming JSON fields is a breaking change for those consumers - confirm backward compatibility before editing."
    }
  }'
fi

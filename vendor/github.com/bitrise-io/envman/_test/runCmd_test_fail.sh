#!/bin/bash

echo "Bitrise value: $BITRISE_FAIL"
if [[ "$BITRISE_FAIL" != "" ]] ; then
    # This is an error, this env should not be set and
    #  should be empty.
    echo "[!] Error: BITRISE_FAIL = $BITRISE_FAIL"
    exit 2
fi

# This is the expected exit code
#  when called through envman it should return this exit code
exit 23

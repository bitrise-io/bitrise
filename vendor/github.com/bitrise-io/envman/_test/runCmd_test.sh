#!/bin/bash

echo "Bitrise value: $BITRISE"
if [[ "$BITRISE" != "test bitrise value" ]] ; then
    exit 1
fi

echo "Bitrise value from stdin: $BITRISE_FROM_STDIN"
ECHO_OUT="bitrise from stdin
"
if [[ "$BITRISE_FROM_STDIN" != "$ECHO_OUT" ]] ; then
    echo "Not equal: ( $BITRISE_FROM_STDIN ) - ( $ECHO_OUT )"
    exit 2
fi

#!/bin/bash

set -e

if [ ! -f "${STEPZIP_PATH}" ] ; then
    echo "[!] STEPZIP_PATH not defined or not a file - required!"
    exit 1
fi

if [ -z "${STEPZIP_UNZIP_TARGET_DIR_PATH}" ] ; then
    echo "[!] STEPZIP_UNZIP_TARGET_DIR_PATH not defined - required!"
    exit 1
fi

set -v

unzip -oq "${STEPZIP_PATH}" -d "${STEPZIP_UNZIP_TARGET_DIR_PATH}"


# => DONE [OK]

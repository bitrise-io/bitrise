#!/bin/bash

set -e

if [ ! -d "${STEP_DIR_TO_ZIP_PATH}" ] ; then
    echo "[!] STEP_DIR_TO_ZIP_PATH not defined or not a dir - required!"
    exit 1
fi

if [ -z "${STEPZIP_PATH}" ] ; then
    echo "[!] STEPZIP_PATH not defined - required!"
    exit 1
fi

set -v

cd "${STEP_DIR_TO_ZIP_PATH}"
# TODO: copy the dir to a temp path
#  and remove .git folder from it, and common
#  unnecessary files&folders like OS X ".DS_Store"
zip -r step.zip .
mv ./step.zip "${STEPZIP_PATH}"

# => DONE [OK]

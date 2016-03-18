#!/bin/bash
set -x

version_file_path="$1"
if [ ! -f "$version_file_path" ] ; then
  echo " [!] version_file_path not provided, or file doesn't exist at path: $version_file_path"
  exit 1
fi
versionNumber=$next_version
if [[ "$versionNumber" == "" ]] ; then
  echo " [!] versionNumber not provided"
  exit 1
fi

cat >"${version_file_path}" <<EOL
package version

// VERSION ...
const VERSION = "${versionNumber}"
EOL

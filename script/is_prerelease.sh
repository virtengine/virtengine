#!/usr/bin/env bash

# in virtengine even minor part of the tag indicates release belongs to the MAINNET
# using it as scripts simplifies debugging as well as portability
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

if [[ $# -ne 1 ]]; then
	echo "illegal number of parameters"
	exit 1
fi

if [[ -z "$1" ]]; then
	echo "empty version tag"
	exit 1
fi

prerel=$("${SCRIPT_DIR}"/semver.sh get prerel "$1" 2>/dev/null) || exit 1
[[ -n "$prerel" ]] && exit 0 || exit 1

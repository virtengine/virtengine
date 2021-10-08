#!/bin/sh

set -e

##
# Configuration sanity check
##

# shellcheck disable=SC2015
[ -f "$VIRTENGINE_BOOT_KEYS/key.txt" ] && [ -f "$VIRTENGINE_BOOT_KEYS/key-pass.txt" ] || {
  echo "Key information not found; VIRTENGINE_BOOT_KEYS is not configured properly"
  exit 1
}

env | sort

##
# Import key
##
/bin/virtengine --home="$VIRTENGINE_HOME" keys import --keyring-backend="$VIRTENGINE_KEYRING_BACKEND"  "$VIRTENGINE_FROM" \
  "$VIRTENGINE_BOOT_KEYS/key.txt" < "$VIRTENGINE_BOOT_KEYS/key-pass.txt"

##
# Run daemon
##
#/virtengine --home=$VIRTENGINE_HOME provider run --cluster-k8s
/bin/virtengine --home="$VIRTENGINE_HOME" --node="$VIRTENGINE_NODE" --chain-id="$VIRTENGINE_CHAIN_ID" --keyring-backend="$VIRTENGINE_KEYRING_BACKEND" --from="$VIRTENGINE_FROM" provider run --cluster-k8s

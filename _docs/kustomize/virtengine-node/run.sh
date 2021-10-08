#!/bin/sh

set -e

env | sort

mkdir -p "$VIRTENGINE_HOME/data"
mkdir -p "$VIRTENGINE_HOME/config"

# XXX it's not reading all of the env variables.

cp "$VIRTENGINE_BOOT_KEYS/priv_validator_state.json"   "$VIRTENGINE_HOME/data/"
cp "$VIRTENGINE_BOOT_DATA/genesis.json" "$VIRTENGINE_HOME/config/"
cp "$VIRTENGINE_BOOT_KEYS/node_key.json" "$VIRTENGINE_HOME/config/"
cp "$VIRTENGINE_BOOT_KEYS/priv_validator_key.json" "$VIRTENGINE_HOME/config/"

/bin/virtengine --home="$VIRTENGINE_HOME" --rpc.laddr="$VIRTENGINE_RPC_LADDR" start

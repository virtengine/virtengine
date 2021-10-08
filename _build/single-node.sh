#!/bin/sh

CHAINID=$1
GENACCT=$2

if [ -z "$1" ]; then
  echo "Need to input chain id..."
  exit 1
fi

if [ -z "$2" ]; then
  echo "Need to input genesis account address..."
  exit 1
fi

# Build genesis file incl account for passed address
coins="1000000000000stake,10000000000000uve"
virtengine init --chain-id "$CHAINID" "$CHAINID"
virtengine keys add validator --keyring-backend="test"
virtengine add-genesis-account "$(virtengine keys show validator -a --keyring-backend="test")" $coins
virtengine add-genesis-account "$GENACCT" $coins
virtengine gentx validator 10000000000stake --keyring-backend="test" --chain-id "$CHAINID"
virtengine collect-gentxs

# Set proper defaults and change ports
sed -i 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:26657"#g' ~/.virtengine/config/config.toml
sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' ~/.virtengine/config/config.toml
sed -i 's/timeout_propose = "3s"/timeout_propose = "1s"/g' ~/.virtengine/config/config.toml
sed -i 's/index_all_keys = false/index_all_keys = true/g' ~/.virtengine/config/config.toml

# Start the virtengine
virtengine start --pruning=nothing
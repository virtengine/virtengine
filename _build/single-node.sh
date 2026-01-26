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

# Clean up previous data
rm -rf ~/.virtengine

# Build genesis file incl account for passed address
coins="100000000000uakt"
virtengine genesis init "$CHAINID" --chain-id "$CHAINID"
virtengine keys add validator --keyring-backend="test"
virtengine genesis add-account "$(virtengine keys show validator -a --keyring-backend="test")" $coins
virtengine genesis add-account "$GENACCT" $coins
virtengine genesis gentx validator 10000000000uakt --keyring-backend="test" --chain-id "$CHAINID" --min-self-delegation="1"
virtengine genesis collect

# Set proper defaults and change ports
sed -i.bak 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:26657"#g' ~/.virtengine/config/config.toml
sed -i.bak 's/timeout_commit = "5s"/timeout_commit = "1s"/g' ~/.virtengine/config/config.toml
sed -i.bak 's/timeout_propose = "3s"/timeout_propose = "1s"/g' ~/.virtengine/config/config.toml
sed -i.bak 's/index_all_keys = false/index_all_keys = true/g' ~/.virtengine/config/config.toml
rm -f ~/.virtengine/config/config.toml.bak

# Start the virtengine
virtengine start --pruning=nothing
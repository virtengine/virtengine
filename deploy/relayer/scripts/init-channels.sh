#!/bin/bash
# Copyright 2024 VirtEngine Contributors
# SPDX-License-Identifier: Apache-2.0

# Initialize IBC channels between VirtEngine and counterparty chains.
# Prerequisites:
#   - Hermes relayer installed and configured (config.toml)
#   - Relayer keys added for all chains
#   - All chains running and accessible

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== VirtEngine IBC Channel Initialization ==="
echo ""

# Create client connections
echo "Creating IBC clients..."

# VirtEngine <-> Cosmos Hub
hermes create client \
  --host-chain virtengine-1 \
  --reference-chain cosmoshub-4 \
  --trusting-period 1209600s

hermes create client \
  --host-chain cosmoshub-4 \
  --reference-chain virtengine-1 \
  --trusting-period 1209600s

# VirtEngine <-> Osmosis
hermes create client \
  --host-chain virtengine-1 \
  --reference-chain osmosis-1 \
  --trusting-period 1209600s

hermes create client \
  --host-chain osmosis-1 \
  --reference-chain virtengine-1 \
  --trusting-period 1209600s

echo ""
echo "Creating IBC connections..."

# Create connections
hermes create connection --a-chain virtengine-1 --b-chain cosmoshub-4
hermes create connection --a-chain virtengine-1 --b-chain osmosis-1

echo ""
echo "Creating IBC channels..."

# Transfer channels (ics20-1, unordered)
hermes create channel \
  --a-chain virtengine-1 \
  --b-chain cosmoshub-4 \
  --a-port transfer \
  --b-port transfer \
  --channel-version ics20-1 \
  --order unordered

hermes create channel \
  --a-chain virtengine-1 \
  --b-chain osmosis-1 \
  --a-port transfer \
  --b-port transfer \
  --channel-version ics20-1 \
  --order unordered

# VEID attestation channel (veid-1, ordered)
hermes create channel \
  --a-chain virtengine-1 \
  --b-chain cosmoshub-4 \
  --a-port veid \
  --b-port veid \
  --channel-version veid-1 \
  --order ordered

echo ""
echo "=== Channel initialization complete ==="
echo ""
echo "To start relaying packets, run:"
echo "  hermes start"

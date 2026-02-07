// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
)

// ChannelConfig defines the configuration for an IBC channel.
type ChannelConfig struct {
	// CounterpartyChainID is the chain ID of the counterparty chain
	CounterpartyChainID string

	// ConnectionID is the connection identifier (set during channel setup)
	ConnectionID string

	// PortID is the port identifier
	PortID string

	// Version is the IBC application version
	Version string

	// Ordering defines the channel ordering
	Ordering channeltypes.Order
}

// DefaultChannels returns the default IBC channel configurations for VirtEngine.
var DefaultChannels = []ChannelConfig{
	{
		CounterpartyChainID: "cosmoshub-4",
		PortID:              "transfer",
		Version:             "ics20-1",
		Ordering:            channeltypes.UNORDERED,
	},
	{
		CounterpartyChainID: "osmosis-1",
		PortID:              "transfer",
		Version:             "ics20-1",
		Ordering:            channeltypes.UNORDERED,
	},
	{
		CounterpartyChainID: "cosmoshub-4",
		PortID:              "veid",
		Version:             "veid-1",
		Ordering:            channeltypes.ORDERED,
	},
}

// GetChannelConfigsByPort returns all channel configurations for a given port.
func GetChannelConfigsByPort(portID string) []ChannelConfig {
	var configs []ChannelConfig
	for _, c := range DefaultChannels {
		if c.PortID == portID {
			configs = append(configs, c)
		}
	}
	return configs
}

// GetChannelConfigsByChain returns all channel configurations for a given counterparty chain.
func GetChannelConfigsByChain(chainID string) []ChannelConfig {
	var configs []ChannelConfig
	for _, c := range DefaultChannels {
		if c.CounterpartyChainID == chainID {
			configs = append(configs, c)
		}
	}
	return configs
}

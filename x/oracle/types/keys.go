// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package types

import (
	types "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
)

const (
	// ModuleName is the module name constant used in many places
	ModuleName = types.ModuleName

	// StoreKey is the store key string for oracle
	StoreKey = types.StoreKey

	// RouterKey is the message route for oracle
	RouterKey = types.RouterKey
)

// Keys for oracle store
// Items are stored with the following key: values
// - 0x01: Params
// - 0x02<source><denom><base_denom>: Latest price data ID per source/pair
// - 0x03<source><denom><base_denom><height>: Price data record

var (
	ParamsKey             = []byte{0x01}
	LatestPriceDataPrefix = []byte{0x02}
	PriceDataPrefix       = []byte{0x03}
)

// ParamsPrefix returns the params prefix key
func ParamsPrefix() []byte {
	return ParamsKey
}

// LatestPriceDataKey returns the key for latest price data ID
func LatestPriceDataKey(source uint32, denom, baseDenom string) []byte {
	key := append(LatestPriceDataPrefix, byte(source>>24), byte(source>>16), byte(source>>8), byte(source))
	key = append(key, []byte(denom)...)
	key = append(key, 0x00) // separator
	key = append(key, []byte(baseDenom)...)
	return key
}

// PriceDataKey returns the key for a price data record
func PriceDataKey(source uint32, denom, baseDenom string, height int64) []byte {
	key := append(PriceDataPrefix, byte(source>>24), byte(source>>16), byte(source>>8), byte(source))
	key = append(key, []byte(denom)...)
	key = append(key, 0x00) // separator
	key = append(key, []byte(baseDenom)...)
	key = append(key, 0x00) // separator
	// Encode height as big-endian for proper ordering
	key = append(key,
		byte(height>>56), byte(height>>48), byte(height>>40), byte(height>>32),
		byte(height>>24), byte(height>>16), byte(height>>8), byte(height),
	)
	return key
}

// PriceDataPrefixByPair returns prefix for all price data for a source/pair
func PriceDataPrefixByPair(source uint32, denom, baseDenom string) []byte {
	key := append(PriceDataPrefix, byte(source>>24), byte(source>>16), byte(source>>8), byte(source))
	key = append(key, []byte(denom)...)
	key = append(key, 0x00) // separator
	key = append(key, []byte(baseDenom)...)
	key = append(key, 0x00) // separator
	return key
}

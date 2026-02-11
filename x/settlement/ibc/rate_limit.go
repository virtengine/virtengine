// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// RateLimitRetentionBlocks is how many blocks of counters to retain.
	RateLimitRetentionBlocks int64 = 2
)

// RateLimitConfig defines per-block packet limits.
type RateLimitConfig struct {
	Enabled                     bool   `json:"enabled"`
	MaxPacketsPerBlock          uint64 `json:"max_packets_per_block"`
	MaxPacketsPerRelayerPerBlock uint64 `json:"max_packets_per_relayer_per_block"`
}

// DefaultRateLimitConfig returns the default IBC rate limit config.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:                     true,
		MaxPacketsPerBlock:          200,
		MaxPacketsPerRelayerPerBlock: 50,
	}
}

// Validate validates the rate limit config.
func (c RateLimitConfig) Validate() error {
	if c.MaxPacketsPerBlock == 0 {
		return fmt.Errorf("max_packets_per_block must be greater than 0")
	}
	if c.MaxPacketsPerRelayerPerBlock == 0 {
		return fmt.Errorf("max_packets_per_relayer_per_block must be greater than 0")
	}
	return nil
}

// GetRateLimitConfig returns the current rate limit config.
func (k IBCKeeper) GetRateLimitConfig(ctx sdk.Context) RateLimitConfig {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(RateLimitConfigKey())
	if bz == nil {
		return DefaultRateLimitConfig()
	}

	var cfg RateLimitConfig
	if err := json.Unmarshal(bz, &cfg); err != nil {
		return DefaultRateLimitConfig()
	}
	return cfg
}

// SetRateLimitConfig updates the rate limit config.
func (k IBCKeeper) SetRateLimitConfig(ctx sdk.Context, cfg RateLimitConfig) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	bz, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(RateLimitConfigKey(), bz)
	return nil
}

// CheckRateLimit verifies packet rate limits and increments counters.
func (k IBCKeeper) CheckRateLimit(ctx sdk.Context, relayer sdk.AccAddress, packetType PacketType) error {
	cfg := k.GetRateLimitConfig(ctx)
	if !cfg.Enabled {
		return nil
	}

	height := uint64(ctx.BlockHeight())
	store := ctx.KVStore(k.storeKey)

	totalKey := RateLimitKey(height, packetType)
	total := readUint64(store.Get(totalKey))
	if total+1 > cfg.MaxPacketsPerBlock {
		return ErrRateLimited.Wrapf("packet type %s exceeded max per block", packetType)
	}

	if !relayer.Empty() {
		relayerKey := RateLimitRelayerKey(height, relayer.String(), packetType)
		relayerCount := readUint64(store.Get(relayerKey))
		if relayerCount+1 > cfg.MaxPacketsPerRelayerPerBlock {
			return ErrRateLimited.Wrapf("relayer %s exceeded max per block", relayer.String())
		}
		store.Set(relayerKey, appendUint64(nil, relayerCount+1))
	}

	store.Set(totalKey, appendUint64(nil, total+1))
	return nil
}

// CleanupRateLimitData prunes old rate limit entries.
func (k IBCKeeper) CleanupRateLimitData(ctx sdk.Context) {
	cutoff := ctx.BlockHeight() - RateLimitRetentionBlocks
	if cutoff <= 0 {
		return
	}

	store := ctx.KVStore(k.storeKey)

	cleanupPrefix := func(prefix []byte) {
		iter := sdk.KVStorePrefixIterator(store, prefix)
		defer iter.Close()

		for ; iter.Valid(); iter.Next() {
			key := iter.Key()
			if len(key) < len(prefix)+8 {
				continue
			}

			height := int64(readUint64(key[len(prefix) : len(prefix)+8]))
			if height <= cutoff {
				store.Delete(key)
			}
		}
	}

	cleanupPrefix(PrefixRateLimit)
	cleanupPrefix(PrefixRateLimitRelayer)
}

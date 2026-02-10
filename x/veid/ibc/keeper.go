// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Store key prefixes for cross-chain VEID records.
var (
	PrefixCrossChainVEID  = []byte{0x20}
	PrefixProcessedNonce  = []byte{0x21}
	PrefixScorePolicy     = []byte{0x22}
	PrefixNonceSequence   = []byte{0x23}
	PrefixCrossChainIndex = []byte{0x24}
)

// IBCKeeper manages IBC-related VEID state.
type IBCKeeper struct {
	cdc       codec.BinaryCodec
	storeKey  storetypes.StoreKey
	authority string
}

// NewIBCKeeper creates a new IBCKeeper.
func NewIBCKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, authority string) IBCKeeper {
	return IBCKeeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
	}
}

// GetScorePolicy returns the current cross-chain score policy.
func (k IBCKeeper) GetScorePolicy(ctx sdk.Context) CrossChainScorePolicy {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(PrefixScorePolicy)
	if bz == nil {
		return DefaultCrossChainScorePolicy()
	}

	var policy CrossChainScorePolicy
	if err := json.Unmarshal(bz, &policy); err != nil {
		return DefaultCrossChainScorePolicy()
	}
	return policy
}

// SetScorePolicy sets the cross-chain score policy.
func (k IBCKeeper) SetScorePolicy(ctx sdk.Context, policy CrossChainScorePolicy) error {
	if err := policy.Validate(); err != nil {
		return fmt.Errorf("invalid score policy: %w", err)
	}

	bz, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal score policy: %w", err)
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(PrefixScorePolicy, bz)
	return nil
}

// StoreCrossChainVEID stores a cross-chain VEID record.
func (k IBCKeeper) StoreCrossChainVEID(ctx sdk.Context, record CrossChainVEIDRecord) error {
	bz, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal cross-chain VEID record: %w", err)
	}

	store := ctx.KVStore(k.storeKey)
	key := crossChainVEIDKey(record.SourceChainID, record.AccountAddress)
	store.Set(key, bz)

	// Store index by address for quick lookups
	indexKey := crossChainIndexKey(record.AccountAddress, record.SourceChainID)
	store.Set(indexKey, key)

	return nil
}

// GetCrossChainVEID retrieves a cross-chain VEID record.
func (k IBCKeeper) GetCrossChainVEID(ctx sdk.Context, sourceChainID, address string) (CrossChainVEIDRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	key := crossChainVEIDKey(sourceChainID, address)
	bz := store.Get(key)
	if bz == nil {
		return CrossChainVEIDRecord{}, false
	}

	var record CrossChainVEIDRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return CrossChainVEIDRecord{}, false
	}
	return record, true
}

// GetCrossChainVEIDsByAddress retrieves all cross-chain VEID records for an address.
func (k IBCKeeper) GetCrossChainVEIDsByAddress(ctx sdk.Context, address string) []CrossChainVEIDRecord {
	store := ctx.KVStore(k.storeKey)
	addrPrefix := []byte(address + "/")
	prefix := make([]byte, 0, len(PrefixCrossChainIndex)+len(addrPrefix))
	prefix = append(prefix, PrefixCrossChainIndex...)
	prefix = append(prefix, addrPrefix...)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var records []CrossChainVEIDRecord
	for ; iter.Valid(); iter.Next() {
		recordKey := iter.Value()
		bz := store.Get(recordKey)
		if bz == nil {
			continue
		}
		var record CrossChainVEIDRecord
		if err := json.Unmarshal(bz, &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	return records
}

// IsNonceProcessed checks if a nonce has already been processed (replay prevention).
func (k IBCKeeper) IsNonceProcessed(ctx sdk.Context, sourceChainID string, nonce uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := processedNonceKey(sourceChainID, nonce)
	return store.Has(key)
}

// MarkNonceProcessed marks a nonce as processed.
func (k IBCKeeper) MarkNonceProcessed(ctx sdk.Context, sourceChainID string, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	key := processedNonceKey(sourceChainID, nonce)
	store.Set(key, []byte{1})
}

// NextNonce returns the next nonce for outgoing packets, using KV store for persistence.
func (k IBCKeeper) NextNonce(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(PrefixNonceSequence)
	var current uint64
	if len(bz) == 8 {
		for i := 0; i < 8; i++ {
			current |= uint64(bz[i]) << ((7 - i) * 8)
		}
	}
	next := current + 1
	nextBz := make([]byte, 0, 8)
	nextBz = appendUint64(nextBz, next)
	store.Set(PrefixNonceSequence, nextBz)
	return next
}

// ProcessAttestation processes an incoming VEID attestation from another chain.
func (k IBCKeeper) ProcessAttestation(ctx sdk.Context, packet VEIDAttestationPacket, sourceChannel string) (CrossChainVEIDRecord, error) {
	// Replay prevention
	if k.IsNonceProcessed(ctx, packet.SourceChainID, packet.Nonce) {
		return CrossChainVEIDRecord{}, fmt.Errorf("nonce %d from chain %s already processed", packet.Nonce, packet.SourceChainID)
	}

	// Get score policy
	policy := k.GetScorePolicy(ctx)

	// Apply degradation
	recognizedScore, err := policy.ApplyDegradation(packet)
	if err != nil {
		return CrossChainVEIDRecord{}, fmt.Errorf("score degradation failed: %w", err)
	}

	// Check attestation expiration
	blockTime := ctx.BlockTime().Unix()
	if blockTime >= packet.ExpirationTime {
		return CrossChainVEIDRecord{}, fmt.Errorf("attestation expired at %d, current time %d", packet.ExpirationTime, blockTime)
	}

	// Create cross-chain record
	record := CrossChainVEIDRecord{
		SourceChainID:   packet.SourceChainID,
		SourceChannel:   sourceChannel,
		AccountAddress:  packet.AccountAddress,
		OriginalScore:   packet.TrustScore,
		RecognizedScore: recognizedScore,
		TierLevel:       packet.TierLevel,
		ReceivedAt:      blockTime,
		ExpiresAt:       packet.ExpirationTime,
		VEIDHash:        packet.VEIDHash,
	}

	// Store the record
	if err := k.StoreCrossChainVEID(ctx, record); err != nil {
		return CrossChainVEIDRecord{}, fmt.Errorf("failed to store cross-chain VEID: %w", err)
	}

	// Mark nonce as processed
	k.MarkNonceProcessed(ctx, packet.SourceChainID, packet.Nonce)

	return record, nil
}

// Key construction helpers

func crossChainVEIDKey(sourceChainID, address string) []byte {
	key := make([]byte, 0, len(PrefixCrossChainVEID)+len(sourceChainID)+1+len(address))
	key = append(key, PrefixCrossChainVEID...)
	key = append(key, []byte(sourceChainID)...)
	key = append(key, byte('/'))
	key = append(key, []byte(address)...)
	return key
}

func processedNonceKey(sourceChainID string, nonce uint64) []byte {
	key := make([]byte, 0, len(PrefixProcessedNonce)+len(sourceChainID)+9)
	key = append(key, PrefixProcessedNonce...)
	key = append(key, []byte(sourceChainID)...)
	key = append(key, byte('/'))
	key = appendUint64(key, nonce)
	return key
}

func crossChainIndexKey(address, sourceChainID string) []byte {
	key := make([]byte, 0, len(PrefixCrossChainIndex)+len(address)+1+len(sourceChainID))
	key = append(key, PrefixCrossChainIndex...)
	key = append(key, []byte(address)...)
	key = append(key, byte('/'))
	key = append(key, []byte(sourceChainID)...)
	return key
}

func appendUint64(bz []byte, n uint64) []byte {
	for i := 7; i >= 0; i-- {
		bz = append(bz, byte(n>>(i*8)))
	}
	return bz
}

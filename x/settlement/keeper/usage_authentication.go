// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"time"

	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	decredsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"

	providertypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	"github.com/virtengine/virtengine/x/settlement/types"
)

type usageReplayEntry struct {
	UsageID string `json:"usage_id"`
	Digest  []byte `json:"digest"`
}

type usageStreamState struct {
	LastSequence uint64 `json:"last_sequence"`
	LastUsageID  string `json:"last_usage_id"`
	LastDigest   []byte `json:"last_digest"`
}

type usagePeriodState struct {
	LastPeriodEnd int64  `json:"last_period_end"`
	LastUsageID   string `json:"last_usage_id"`
}

type acknowledgmentReplayEntry struct {
	UsageID string `json:"usage_id"`
	Digest  []byte `json:"digest"`
}

// GetUsageStreamState returns the authoritative sequence cursor for producer
// restart reconciliation.
func (k Keeper) GetUsageStreamState(ctx sdk.Context, provider, allocationID, orderID, leaseID string) (uint64, string, []byte, error) {
	streamID, err := types.UsageStreamID(provider, allocationID, orderID, leaseID)
	if err != nil {
		return 0, "", nil, types.ErrInvalidUsageRecord.Wrap(err.Error())
	}
	var stream usageStreamState
	bz := ctx.KVStore(k.skey).Get(types.UsageStreamStateKey(streamID))
	if bz == nil {
		return 0, "", nil, nil
	}
	if err := json.Unmarshal(bz, &stream); err != nil {
		return 0, "", nil, types.ErrUsageSequenceGap.Wrap("corrupt stream state")
	}
	return stream.LastSequence, stream.LastUsageID, append([]byte(nil), stream.LastDigest...), nil
}

// ActivateUsageAuthentication marks the v1.5.0 migration boundary. It is
// deliberately state-gated so a new binary cannot activate enforcement before
// the coordinated software-upgrade height.
func (k Keeper) ActivateUsageAuthentication(ctx sdk.Context) error {
	ctx.KVStore(k.skey).Set(types.UsageAuthenticationActivationKey(), []byte{byte(types.SignatureVersionV1)})
	return nil
}

// IsUsageAuthenticationActive reports whether the upgrade marker is present.
func (k Keeper) IsUsageAuthenticationActive(ctx sdk.Context) bool {
	value := ctx.KVStore(k.skey).Get(types.UsageAuthenticationActivationKey())
	return len(value) == 1 && value[0] == byte(types.SignatureVersionV1)
}

func (k Keeper) verifyAuthenticatedUsage(ctx sdk.Context, record *types.UsageRecord) ([]byte, []byte, string, bool, error) {
	if k.providerKeyKeeper == nil {
		return nil, nil, "", false, types.ErrUsageAuthenticationRequired.Wrap("provider signing-key keeper not configured")
	}
	if record.SignatureVersion != types.SignatureVersionV1 {
		return nil, nil, "", false, types.ErrUsageAuthenticationRequired.Wrapf("unsupported signature version %d", record.SignatureVersion)
	}
	if record.ChainID != "" && record.ChainID != ctx.ChainID() {
		return nil, nil, "", false, types.ErrInvalidSignature.Wrap("proof chain_id does not match execution chain")
	}
	record.ChainID = ctx.ChainID()
	if record.PricingVersion != 1 || record.FormulaVersion != 1 || record.ModelVersion != 1 {
		return nil, nil, "", false, types.ErrUsagePricingVersion.Wrap("only pricing, formula, and model version 1 are available")
	}
	if err := validateUsageProofWindow(ctx, record.IssuedAtHeight, record.ExpiresAtHeight, record.IssuedAtUnix, record.ExpiresAtUnix); err != nil {
		return nil, nil, "", false, err
	}
	if err := validateUsagePeriodAtBlock(ctx, record); err != nil {
		return nil, nil, "", false, err
	}

	provider, err := sdk.AccAddressFromBech32(record.Provider)
	if err != nil {
		return nil, nil, "", false, types.ErrInvalidAddress.Wrap("invalid provider address")
	}
	keyRecord, found := k.providerKeyKeeper.GetProviderSigningKey(ctx, provider, record.ProviderKeyID, record.ProviderKeyEpoch)
	if !found {
		return nil, nil, "", false, types.ErrProviderSigningKeyNotFound
	}
	issuedTime := time.Unix(record.IssuedAtUnix, 0).UTC()
	if !keyRecord.IsValidAt(record.IssuedAtHeight, issuedTime) || !keyRecord.IsValidAt(ctx.BlockHeight(), ctx.BlockTime()) {
		return nil, nil, "", false, types.ErrProviderSigningKeyInactive
	}

	payload := record.CanonicalUsagePayload(ctx.ChainID())
	signBytes, err := types.CanonicalUsageSignBytes(payload)
	if err != nil {
		return nil, nil, "", false, types.ErrInvalidUsageRecord.Wrap(err.Error())
	}
	if !verifyProviderUsageSignature(keyRecord, signBytes, record.ProviderSignature) {
		return nil, nil, "", false, types.ErrInvalidSignature.Wrap("provider detached signature verification failed")
	}
	digestArray := sha256.Sum256(signBytes)
	digest := digestArray[:]
	streamID, err := types.UsageStreamID(record.Provider, record.AllocationID, record.OrderID, record.LeaseID)
	if err != nil {
		return nil, nil, "", false, types.ErrInvalidUsageRecord.Wrap(err.Error())
	}

	usageID, duplicate, err := k.checkUsageReplay(ctx, record, streamID, digest)
	if err != nil {
		return nil, nil, "", false, err
	}
	if duplicate {
		return digest, streamID, usageID, true, nil
	}
	if err := k.checkUsageSequenceAndPeriod(ctx, record, streamID); err != nil {
		return nil, nil, "", false, err
	}
	return digest, streamID, "", false, nil
}

func verifyProviderUsageSignature(record providertypes.ProviderPublicKeyRecord, signBytes, signature []byte) bool {
	switch record.KeyType {
	case providertypes.PublicKeyTypeEd25519:
		return len(record.PublicKey) == ed25519.PublicKeySize &&
			len(signature) == ed25519.SignatureSize &&
			ed25519.Verify(record.PublicKey, signBytes, signature)
	case providertypes.PublicKeyTypeSecp256k1:
		return len(record.PublicKey) == cosmossecp256k1.PubKeySize &&
			isCanonicalSecp256k1Signature(signature) &&
			(&cosmossecp256k1.PubKey{Key: record.PublicKey}).VerifySignature(signBytes, signature)
	default:
		return false
	}
}

func validateUsageProofWindow(ctx sdk.Context, issuedHeight, expiresHeight, issuedUnix, expiresUnix int64) error {
	currentHeight := ctx.BlockHeight()
	currentUnix := ctx.BlockTime().Unix()
	if issuedHeight > currentHeight+types.MaxProofFutureBlocks || issuedHeight < currentHeight-types.MaxProofPastBlocks {
		return types.ErrUsageProofExpired.Wrap("issued height outside maximum skew")
	}
	if expiresHeight < currentHeight {
		return types.ErrUsageProofExpired.Wrap("proof height expired")
	}
	if issuedUnix > currentUnix+types.MaxProofFutureSeconds || issuedUnix < currentUnix-types.MaxProofPastSeconds {
		return types.ErrUsageProofExpired.Wrap("issued block-time outside maximum skew")
	}
	if expiresUnix < currentUnix {
		return types.ErrUsageProofExpired.Wrap("proof block-time expired")
	}
	if expiresHeight < issuedHeight || expiresHeight-issuedHeight > types.MaxProofLifetimeBlocks || expiresUnix < issuedUnix || expiresUnix-issuedUnix > types.MaxProofLifetimeSeconds {
		return types.ErrUsageProofExpired.Wrap("proof lifetime exceeds protocol maximum")
	}
	return nil
}

func validateUsagePeriodAtBlock(ctx sdk.Context, record *types.UsageRecord) error {
	end := record.PeriodEnd.Unix()
	now := ctx.BlockTime().Unix()
	if end > now+types.MaxProofFutureSeconds || end < now-types.MaxProofPastSeconds {
		return types.ErrUsageProofExpired.Wrap("usage period end outside maximum chain-time skew")
	}
	return nil
}

func (k Keeper) checkUsageReplay(ctx sdk.Context, record *types.UsageRecord, streamID, digest []byte) (string, bool, error) {
	store := ctx.KVStore(k.skey)
	keys := [][]byte{
		types.UsageReplaySequenceKey(streamID, record.Sequence),
		types.UsageReplayNonceKey(record.Nonce),
		types.UsageReplayIdempotencyKey(record.IdempotencyKey),
	}
	entries := make([]usageReplayEntry, 0, len(keys))
	for _, key := range keys {
		bz := store.Get(key)
		if bz == nil {
			continue
		}
		var entry usageReplayEntry
		if err := json.Unmarshal(bz, &entry); err != nil {
			return "", false, types.ErrUsageReplayConflict.Wrap("corrupt replay index")
		}
		entries = append(entries, entry)
	}
	if len(entries) == 0 {
		return "", false, nil
	}
	if len(entries) != len(keys) {
		return "", false, types.ErrUsageReplayConflict.Wrap("partial replay-key collision")
	}
	usageID := entries[0].UsageID
	for _, entry := range entries {
		if entry.UsageID != usageID || !bytes.Equal(entry.Digest, digest) {
			return "", false, types.ErrUsageReplayConflict
		}
	}
	return usageID, true, nil
}

func (k Keeper) checkUsageSequenceAndPeriod(ctx sdk.Context, record *types.UsageRecord, streamID []byte) error {
	store := ctx.KVStore(k.skey)
	var stream usageStreamState
	if bz := store.Get(types.UsageStreamStateKey(streamID)); bz != nil {
		if err := json.Unmarshal(bz, &stream); err != nil {
			return types.ErrUsageSequenceGap.Wrap("corrupt stream state")
		}
	}
	expected := uint64(1)
	if stream.LastSequence > 0 {
		if stream.LastSequence == math.MaxUint64 {
			return types.ErrUsageSequenceGap.Wrap("stream sequence exhausted")
		}
		expected = stream.LastSequence + 1
	}
	if record.Sequence != expected {
		return types.ErrUsageSequenceGap.Wrapf("expected sequence %d, got %d", expected, record.Sequence)
	}

	var period usagePeriodState
	if bz := store.Get(types.UsagePeriodStateKey(streamID, record.UsageType)); bz != nil {
		if err := json.Unmarshal(bz, &period); err != nil {
			return types.ErrUsagePeriodOverlap.Wrap("corrupt period state")
		}
		start := record.PeriodStart.Unix()
		if start < period.LastPeriodEnd {
			return types.ErrUsagePeriodOverlap.Wrapf("period starts at %d before prior end %d", start, period.LastPeriodEnd)
		}
		if start-period.LastPeriodEnd > types.MaxUsagePeriodGapSeconds {
			return types.ErrUsagePeriodOverlap.Wrap("period gap exceeds protocol maximum")
		}
	}
	return nil
}

func (k Keeper) commitUsageReplay(ctx sdk.Context, record types.UsageRecord, streamID, digest []byte) error {
	store := ctx.KVStore(k.skey)
	entry := usageReplayEntry{UsageID: record.UsageID, Digest: append([]byte(nil), digest...)}
	entryBytes, err := json.Marshal(&entry)
	if err != nil {
		return err
	}
	for _, key := range [][]byte{
		types.UsageReplaySequenceKey(streamID, record.Sequence),
		types.UsageReplayNonceKey(record.Nonce),
		types.UsageReplayIdempotencyKey(record.IdempotencyKey),
	} {
		store.Set(key, entryBytes)
	}
	streamBytes, err := json.Marshal(&usageStreamState{
		LastSequence: record.Sequence,
		LastUsageID:  record.UsageID,
		LastDigest:   append([]byte(nil), digest...),
	})
	if err != nil {
		return err
	}
	store.Set(types.UsageStreamStateKey(streamID), streamBytes)
	periodBytes, err := json.Marshal(&usagePeriodState{LastPeriodEnd: record.PeriodEnd.Unix(), LastUsageID: record.UsageID})
	if err != nil {
		return err
	}
	store.Set(types.UsagePeriodStateKey(streamID, record.UsageType), periodBytes)
	return nil
}

// AcknowledgeUsageAuthenticated verifies a detached x/auth account signature.
func (k Keeper) AcknowledgeUsageAuthenticated(ctx sdk.Context, usageID string, proof types.UsageAcknowledgmentProof) error {
	cacheCtx, write := ctx.CacheContext()
	if err := k.acknowledgeUsageAuthenticated(cacheCtx, usageID, proof); err != nil {
		return err
	}
	write()
	return nil
}

func (k Keeper) acknowledgeUsageAuthenticated(ctx sdk.Context, usageID string, proof types.UsageAcknowledgmentProof) error {
	if !k.IsUsageAuthenticationActive(ctx) {
		return types.ErrUsageAuthenticationRequired.Wrap("customer proof endpoint is not active")
	}
	usage, found := k.GetUsageRecord(ctx, usageID)
	if !found {
		return types.ErrUsageRecordNotFound.Wrapf("usage record %s not found", usageID)
	}
	if usage.Settled {
		return types.ErrUsageAlreadySettled.Wrapf("usage record %s already settled", usageID)
	}
	if !usage.IsAuthenticated() || !bytes.Equal(proof.UsageDigest, usage.UsageDigest) {
		return types.ErrInvalidSignature.Wrap("acknowledgment does not bind the stored authenticated usage digest")
	}
	if err := validateUsageProofWindow(ctx, proof.IssuedAtHeight, proof.ExpiresAtHeight, proof.IssuedAtUnix, proof.ExpiresAtUnix); err != nil {
		return err
	}
	payload := proof.CanonicalPayload(ctx.ChainID(), usage.Customer, usageID)
	signBytes, err := types.CanonicalAcknowledgmentSignBytes(payload)
	if err != nil {
		return types.ErrInvalidSignature.Wrap(err.Error())
	}
	ackDigestArray := sha256.Sum256(signBytes)
	ackDigest := ackDigestArray[:]

	store := ctx.KVStore(k.skey)
	if bz := store.Get(types.UsageAckReplayKey(proof.ReplayKey)); bz != nil {
		var entry acknowledgmentReplayEntry
		if err := json.Unmarshal(bz, &entry); err != nil || entry.UsageID != usageID || !bytes.Equal(entry.Digest, ackDigest) {
			return types.ErrUsageReplayConflict.Wrap("customer acknowledgment replay-key conflict")
		}
		if usage.CustomerAcknowledged && bytes.Equal(usage.CustomerAckReplayKey, proof.ReplayKey) {
			return nil
		}
	}
	if usage.CustomerAcknowledged {
		return types.ErrUsageReplayConflict.Wrap("usage already acknowledged with a different proof")
	}
	if k.accountKeeper == nil {
		return types.ErrCustomerKeyUnsupported.Wrap("account keeper not configured")
	}
	customer, err := sdk.AccAddressFromBech32(usage.Customer)
	if err != nil {
		return types.ErrInvalidAddress.Wrap("invalid customer address")
	}
	account := k.accountKeeper.GetAccount(ctx, customer)
	if account == nil || account.GetPubKey() == nil || !account.GetAddress().Equals(customer) {
		return types.ErrCustomerKeyUnsupported.Wrap("customer account or public key unavailable")
	}
	if !verifyCustomerAccountSignature(account, signBytes, proof.Signature) {
		return types.ErrInvalidSignature.Wrap("customer detached signature verification failed")
	}

	usage.CustomerAcknowledged = true
	usage.CustomerSignature = append([]byte(nil), proof.Signature...)
	usage.CustomerAckReplayKey = append([]byte(nil), proof.ReplayKey...)
	usage.CustomerAckIssuedAtHeight = proof.IssuedAtHeight
	usage.CustomerAckExpiresAtHeight = proof.ExpiresAtHeight
	usage.CustomerAckIssuedAtUnix = proof.IssuedAtUnix
	usage.CustomerAckExpiresAtUnix = proof.ExpiresAtUnix
	usage.CustomerAckSignatureVersion = proof.SignatureVersion
	if err := k.SetUsageRecord(ctx, usage); err != nil {
		return err
	}
	entryBytes, err := json.Marshal(&acknowledgmentReplayEntry{UsageID: usageID, Digest: ackDigest})
	if err != nil {
		return err
	}
	store.Set(types.UsageAckReplayKey(proof.ReplayKey), entryBytes)
	k.Logger(ctx).Info("authenticated usage acknowledged", "usage_id", usageID, "customer", usage.Customer)
	return nil
}

func verifyCustomerAccountSignature(account sdk.AccountI, signBytes, signature []byte) bool {
	switch pubKey := account.GetPubKey().(type) {
	case *cosmosed25519.PubKey:
		return len(signature) == ed25519.SignatureSize && pubKey.VerifySignature(signBytes, signature)
	case *cosmossecp256k1.PubKey:
		return isCanonicalSecp256k1Signature(signature) && pubKey.VerifySignature(signBytes, signature)
	default:
		// Multisig, ethsecp256k1, and custom keys require an explicit canonical
		// detached-signature policy and are rejected until one is implemented.
		return false
	}
}

func isCanonicalSecp256k1Signature(signature []byte) bool {
	if len(signature) != 64 {
		return false
	}
	var s decredsecp256k1.ModNScalar
	if s.SetByteSlice(signature[32:]) || s.IsZero() || s.IsOverHalfOrder() {
		return false
	}
	return true
}

// ValidateUsageReplayIndexes returns deterministic invariant violations.
func (k Keeper) ValidateUsageReplayIndexes(ctx sdk.Context) []string {
	broken := make([]string, 0)
	store := ctx.KVStore(k.skey)
	k.WithUsageRecords(ctx, func(record types.UsageRecord) bool {
		if !record.IsAuthenticated() {
			return false
		}
		streamID, err := types.UsageStreamID(record.Provider, record.AllocationID, record.OrderID, record.LeaseID)
		if err != nil {
			broken = append(broken, fmt.Sprintf("usage=%s invalid stream: %v", record.UsageID, err))
			return false
		}
		indexes := []struct {
			label string
			key   []byte
		}{
			{label: "sequence", key: types.UsageReplaySequenceKey(streamID, record.Sequence)},
			{label: "nonce", key: types.UsageReplayNonceKey(record.Nonce)},
			{label: "idempotency", key: types.UsageReplayIdempotencyKey(record.IdempotencyKey)},
		}
		for _, index := range indexes {
			var entry usageReplayEntry
			bz := store.Get(index.key)
			if bz == nil || json.Unmarshal(bz, &entry) != nil || entry.UsageID != record.UsageID || !bytes.Equal(entry.Digest, record.UsageDigest) {
				broken = append(broken, fmt.Sprintf("usage=%s %s replay index mismatch", record.UsageID, index.label))
			}
		}
		return false
	})
	return broken
}

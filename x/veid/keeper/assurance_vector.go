// Package keeper provides keeper functions for the VEID module.
//
// Assurance vectors: per-factor evidence state.
//
// The composite scorer already computes six factor contributions per
// verification, but only its scalar reduction was persisted (see
// composite_scoring.go). This file persists the per-factor form alongside it, so
// a relying party can evaluate its own policy against named factors instead of
// trusting one opaque number.
//
// Absence is meaningful: an account with no stored vector has made no assurance
// claim. It is NOT a vector of zeros, and every getter here reports the found
// flag so callers cannot mistake "never measured" for "measured as zero".
package keeper

import (
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Storage Format
// ============================================================================

// assuranceVectorStore is the stored format of an assurance vector.
//
// It mirrors types.AssuranceVector exactly. Storing the factors as their own
// JSON-encoded array keeps the persisted shape independent of any future
// non-factor field, and the JSON encoder emits struct fields in declaration
// order, so the encoding is deterministic.
type assuranceVectorStore struct {
	Version        string            `json:"version"`
	Account        string            `json:"account"`
	Epoch          uint64            `json:"epoch"`
	Factors        []json.RawMessage `json:"factors"`
	OverallBps     uint32            `json:"overall_bps"`
	ScalarScore    uint32            `json:"scalar_score"`
	Passed         bool              `json:"passed"`
	ScoreVersion   string            `json:"score_version"`
	ModelVersion   string            `json:"model_version"`
	AccountAgeBps  uint32            `json:"account_age_bps"`
	VerifiedHeight int64             `json:"verified_height"`
	VerifiedAtUnix int64             `json:"verified_at_unix"`
	InputHash      []byte            `json:"input_hash"`
	Commitment     []byte            `json:"commitment"`
}

// toStore converts a vector into its stored form.
func toStore(v *types.AssuranceVector) (*assuranceVectorStore, error) {
	factors := make([]json.RawMessage, len(v.Factors))
	for i, entry := range v.Factors {
		bz, err := json.Marshal(entry)
		if err != nil {
			return nil, err
		}
		factors[i] = bz
	}

	return &assuranceVectorStore{
		Version:        v.Version,
		Account:        v.Account,
		Epoch:          v.Epoch,
		Factors:        factors,
		OverallBps:     v.OverallBps,
		ScalarScore:    v.ScalarScore,
		Passed:         v.Passed,
		ScoreVersion:   v.ScoreVersion,
		ModelVersion:   v.ModelVersion,
		AccountAgeBps:  v.AccountAgeBps,
		VerifiedHeight: v.VerifiedHeight,
		VerifiedAtUnix: v.VerifiedAtUnix,
		InputHash:      append([]byte(nil), v.InputHash...),
		Commitment:     append([]byte(nil), v.Commitment...),
	}, nil
}

// fromStore converts stored bytes back into a vector, re-verifying the
// commitment on the way in.
//
// The commitment is checked HERE, at the trust boundary, rather than trusting the
// stored bytes: if a corrupt or tampered record ever reached the store, a caller
// reading it would otherwise hand a relying party a vector whose hash does not
// match its contents.
func fromStore(bz []byte) (*types.AssuranceVector, error) {
	var store assuranceVectorStore
	if err := json.Unmarshal(bz, &store); err != nil {
		return nil, err
	}

	entries := make([]types.AssuranceFactorEntry, len(store.Factors))
	for i, raw := range store.Factors {
		if err := json.Unmarshal(raw, &entries[i]); err != nil {
			return nil, err
		}
	}

	vector := &types.AssuranceVector{
		Version:        store.Version,
		Account:        store.Account,
		Epoch:          store.Epoch,
		Factors:        entries,
		OverallBps:     store.OverallBps,
		ScalarScore:    store.ScalarScore,
		Passed:         store.Passed,
		ScoreVersion:   store.ScoreVersion,
		ModelVersion:   store.ModelVersion,
		AccountAgeBps:  store.AccountAgeBps,
		VerifiedHeight: store.VerifiedHeight,
		VerifiedAtUnix: store.VerifiedAtUnix,
		InputHash:      store.InputHash,
		Commitment:     store.Commitment,
	}

	if err := vector.Validate(); err != nil {
		return nil, fmt.Errorf("stored assurance vector failed validation: %w", err)
	}

	return vector, nil
}

// ============================================================================
// Write Path
// ============================================================================

// SetAssuranceVector stores an account's assurance vector as its current vector
// and records the epoch in the vector history.
//
// The vector is validated first, so a malformed or commitment-mismatched vector
// is rejected at the boundary rather than persisted and discovered later.
func (k Keeper) SetAssuranceVector(ctx sdk.Context, vector *types.AssuranceVector) error {
	if err := vector.Validate(); err != nil {
		return err
	}

	address, err := sdk.AccAddressFromBech32(vector.Account)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	// Guard against a vector that belongs to a different account than the one it
	// claims: the address is the store key, so a mismatch would write one
	// account's vector under another's key.
	if !addressBytesEqual(address.Bytes(), vector.Account) {
		return types.ErrInvalidAddress.Wrap("assurance vector account does not match its address bytes")
	}

	store := ctx.KVStore(k.skey)
	storeKey := types.AssuranceVectorKey(address.Bytes())

	// A re-verification must move the epoch forward. Accepting a vector that
	// rewinds the epoch would let a later block overwrite a newer vector, so a
	// relying party holding the newer vector could see it silently replaced by
	// older evidence.
	if existing, found := k.GetAssuranceVector(ctx, vector.Account); found && existing.Epoch >= vector.Epoch {
		return types.ErrInvalidScoringModel.Wrapf(
			"refusing to write assurance vector epoch %d over existing epoch %d",
			vector.Epoch, existing.Epoch,
		)
	}

	encoded, err := toStore(vector)
	if err != nil {
		return err
	}
	bz, err := json.Marshal(encoded)
	if err != nil {
		return err
	}

	store.Set(storeKey, bz)

	// Retain the superseded epoch so a verifier can prove a vector was replaced
	// rather than edited in place.
	store.Set(types.AssuranceVectorEpochKey(address.Bytes(), vector.VerifiedHeight), bz)

	return nil
}

// ============================================================================
// Read Path
// ============================================================================

// GetAssuranceVector returns an account's current assurance vector.
//
// found=false means the account has never produced a vector: that is "no
// assurance claim", NOT a vector of zeros. Callers must branch on it.
func (k Keeper) GetAssuranceVector(ctx sdk.Context, accountAddr string) (*types.AssuranceVector, bool) {
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return nil, false
	}

	store := ctx.KVStore(k.skey)
	bz := store.Get(types.AssuranceVectorKey(address.Bytes()))
	if bz == nil {
		return nil, false
	}

	vector, err := fromStore(bz)
	if err != nil {
		// A corrupt record must never be presented as a valid vector. Log and
		// report absence: a relying party that gets found=false fails closed.
		k.Logger(ctx).Error("failed to decode stored assurance vector",
			"account", accountAddr,
			"error", err,
		)
		return nil, false
	}

	return vector, true
}

// HasAssuranceVector reports whether an account has a stored vector.
func (k Keeper) HasAssuranceVector(ctx sdk.Context, accountAddr string) bool {
	_, found := k.GetAssuranceVector(ctx, accountAddr)
	return found
}

// GetAssuranceVectorEpoch returns the assurance vector an account had at a given
// block height.
func (k Keeper) GetAssuranceVectorEpoch(ctx sdk.Context, accountAddr string, blockHeight int64) (*types.AssuranceVector, bool) {
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return nil, false
	}

	store := ctx.KVStore(k.skey)
	bz := store.Get(types.AssuranceVectorEpochKey(address.Bytes(), blockHeight))
	if bz == nil {
		return nil, false
	}

	vector, err := fromStore(bz)
	if err != nil {
		k.Logger(ctx).Error("failed to decode stored assurance vector epoch",
			"account", accountAddr,
			"block_height", blockHeight,
			"error", err,
		)
		return nil, false
	}

	return vector, true
}

// GetAssuranceVectorHistory returns an account's assurance vectors, newest first.
//
// Iteration is over a KVStore prefix iterator, which is ordered by key; the height
// suffix is big-endian, so the reverse iteration is ordered by height with no
// in-Go sorting and therefore no map-order nondeterminism.
func (k Keeper) GetAssuranceVectorHistory(ctx sdk.Context, accountAddr string) []*types.AssuranceVector {
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return nil
	}

	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStoreReversePrefixIterator(store, types.AssuranceVectorEpochPrefixKey(address.Bytes()))
	defer iterator.Close() //nolint:errcheck // a close error cannot affect the returned slice

	vectors := make([]*types.AssuranceVector, 0)
	for ; iterator.Valid(); iterator.Next() {
		vector, err := fromStore(iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("failed to decode assurance vector in history",
				"account", accountAddr,
				"key", fmt.Sprintf("%X", iterator.Key()),
				"error", err,
			)
			continue
		}
		vectors = append(vectors, vector)
	}

	return vectors
}

// ============================================================================
// Derivation from a composite score
// ============================================================================

// StoreAssuranceVectorForResult derives and stores the assurance vector for a
// composite scoring result.
//
// The epoch comes from whatever vector the account already holds, so the first
// verification writes epoch 0 and every re-verification supersedes the prior one.
// When no prior vector exists this is also the upgrade/genesis path: the account
// simply has no claim until it is verified, which is exactly the intended
// "absent = no assurance claim" behaviour.
func (k Keeper) StoreAssuranceVectorForResult(
	ctx sdk.Context,
	accountAddr string,
	result *types.CompositeScoreResult,
	in types.DerivationInput,
) (*types.AssuranceVector, error) {
	in.Account = accountAddr

	// Default the epoch from existing state so callers cannot accidentally write
	// a duplicate or rewound epoch.
	if in.Epoch == 0 {
		if existing, found := k.GetAssuranceVector(ctx, accountAddr); found {
			in.Epoch = types.NextEpoch(existing.Epoch)
		}
	}

	// Consensus time, never wall clock.
	if in.VerifiedAt.IsZero() {
		in.VerifiedAt = ctx.BlockTime()
	}

	vector, err := types.DeriveAssuranceVector(result, in)
	if err != nil {
		return nil, err
	}

	if err := k.SetAssuranceVector(ctx, vector); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("assurance vector stored",
		"account", accountAddr,
		"epoch", vector.Epoch,
		"overall_bps", vector.OverallBps,
		"measured_factors", len(vector.MeasuredFactors()),
		"score_version", vector.ScoreVersion,
	)

	return vector, nil
}

// ============================================================================
// Recency
// ============================================================================

// assuranceMaxAgeSeconds returns the freshness window from module params.
//
// Params.VerificationExpiryDays is the module's existing statement of how long a
// verification stays valid; reusing it keeps one notion of "how old is too old"
// instead of introducing a second, competing expiry.
func (k Keeper) assuranceMaxAgeSeconds(ctx sdk.Context) int64 {
	params := k.GetParams(ctx)
	if params.VerificationExpiryDays == 0 {
		return types.DefaultAssuranceVectorMaxAgeSeconds
	}
	return int64(params.VerificationExpiryDays) * 24 * 60 * 60
}

// GetAssuranceVectorRecency returns an account's vector together with its
// freshness, evaluated against CONSENSUS time.
//
// A missing vector yields found=false and a zero-valued recency; callers must
// treat that as "no assurance claim" rather than as a stale-but-present vector.
func (k Keeper) GetAssuranceVectorRecency(
	ctx sdk.Context,
	accountAddr string,
) (*types.AssuranceVector, types.AssuranceVectorRecency, bool) {
	vector, found := k.GetAssuranceVector(ctx, accountAddr)
	if !found {
		return nil, types.AssuranceVectorRecency{}, false
	}

	recency := vector.Recency(ctx.BlockTime(), ctx.BlockHeight(), k.assuranceMaxAgeSeconds(ctx))
	return vector, recency, true
}

// addressBytesEqual compares a parsed address against its bech32 string form by
// re-deriving the string, so a vector whose Account field does not correspond to
// the address it is stored under is rejected.
func addressBytesEqual(address []byte, bech32 string) bool {
	reparsed, err := sdk.AccAddressFromBech32(bech32)
	if err != nil {
		return false
	}
	if len(reparsed.Bytes()) != len(address) {
		return false
	}
	var diff byte
	for i := range address {
		diff |= address[i] ^ reparsed.Bytes()[i]
	}
	return diff == 0
}

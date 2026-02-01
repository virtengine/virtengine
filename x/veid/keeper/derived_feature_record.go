package keeper

import (
	"encoding/json"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Derived Feature Verification Record Management (VE-217)
// ============================================================================

// derivedFeatureRecordStore is the storage format for verification records
type derivedFeatureRecordStore struct {
	RecordID          string                        `json:"record_id"`
	AccountAddress    string                        `json:"account_address"`
	Version           uint32                        `json:"version"`
	RequestID         string                        `json:"request_id"`
	FeatureReferences []featureReferenceStore       `json:"feature_references"`
	CompositeHash     []byte                        `json:"composite_hash"`
	ModelVersion      string                        `json:"model_version"`
	ModelHash         string                        `json:"model_hash"`
	Score             uint32                        `json:"score"`
	Confidence        uint32                        `json:"confidence"`
	Status            types.VerificationResultStatus `json:"status"`
	ReasonCodes       []types.ReasonCode            `json:"reason_codes,omitempty"`
	ComputedAt        int64                         `json:"computed_at"`
	BlockHeight       int64                         `json:"block_height"`
	ComputedBy        string                        `json:"computed_by"`
	ConsensusVotes    []consensusVoteStore          `json:"consensus_votes,omitempty"`
	Finalized         bool                          `json:"finalized"`
	FinalizedAt       *int64                        `json:"finalized_at,omitempty"`
	FinalizedAtBlock  *int64                        `json:"finalized_at_block,omitempty"`
}

// featureReferenceStore is the storage format for feature references
type featureReferenceStore struct {
	FeatureType   string                   `json:"feature_type"`
	FeatureHash   []byte                   `json:"feature_hash"`
	SourceScopeID string                   `json:"source_scope_id"`
	EnvelopeID    string                   `json:"envelope_id,omitempty"`
	FieldKey      string                   `json:"field_key,omitempty"`
	Weight        uint32                   `json:"weight"`
	MatchResult   types.FeatureMatchResult `json:"match_result"`
	MatchScore    uint32                   `json:"match_score"`
}

// consensusVoteStore is the storage format for consensus votes
type consensusVoteStore struct {
	ValidatorAddress string `json:"validator_address"`
	Agreed           bool   `json:"agreed"`
	ComputedHash     []byte `json:"computed_hash"`
	ComputedScore    uint32 `json:"computed_score"`
	VotedAt          int64  `json:"voted_at"`
	BlockHeight      int64  `json:"block_height"`
	Signature        []byte `json:"signature"`
}

func featureReferenceToStore(ref types.DerivedFeatureReference) featureReferenceStore {
	return featureReferenceStore{
		FeatureType:   ref.FeatureType,
		FeatureHash:   ref.FeatureHash,
		SourceScopeID: ref.SourceScopeID,
		EnvelopeID:    ref.EnvelopeID,
		FieldKey:      ref.FieldKey,
		Weight:        ref.Weight,
		MatchResult:   ref.MatchResult,
		MatchScore:    ref.MatchScore,
	}
}

func featureReferenceFromStore(s featureReferenceStore) types.DerivedFeatureReference {
	return types.DerivedFeatureReference{
		FeatureType:   s.FeatureType,
		FeatureHash:   s.FeatureHash,
		SourceScopeID: s.SourceScopeID,
		EnvelopeID:    s.EnvelopeID,
		FieldKey:      s.FieldKey,
		Weight:        s.Weight,
		MatchResult:   s.MatchResult,
		MatchScore:    s.MatchScore,
	}
}

func consensusVoteToStore(v types.ConsensusVote) consensusVoteStore {
	return consensusVoteStore{
		ValidatorAddress: v.ValidatorAddress,
		Agreed:           v.Agreed,
		ComputedHash:     v.ComputedHash,
		ComputedScore:    v.ComputedScore,
		VotedAt:          v.VotedAt.Unix(),
		BlockHeight:      v.BlockHeight,
		Signature:        v.Signature,
	}
}

func consensusVoteFromStore(s consensusVoteStore) types.ConsensusVote {
	return types.ConsensusVote{
		ValidatorAddress: s.ValidatorAddress,
		Agreed:           s.Agreed,
		ComputedHash:     s.ComputedHash,
		ComputedScore:    s.ComputedScore,
		VotedAt:          time.Unix(s.VotedAt, 0),
		BlockHeight:      s.BlockHeight,
		Signature:        s.Signature,
	}
}

// SetDerivedFeatureRecord stores a derived feature verification record
func (k Keeper) SetDerivedFeatureRecord(ctx sdk.Context, record types.DerivedFeatureVerificationRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	// Convert to storage format
	refs := make([]featureReferenceStore, len(record.FeatureReferences))
	for i, ref := range record.FeatureReferences {
		refs[i] = featureReferenceToStore(ref)
	}

	votes := make([]consensusVoteStore, len(record.ConsensusVotes))
	for i, vote := range record.ConsensusVotes {
		votes[i] = consensusVoteToStore(vote)
	}

	rs := derivedFeatureRecordStore{
		RecordID:          record.RecordID,
		AccountAddress:    record.AccountAddress,
		Version:           record.Version,
		RequestID:         record.RequestID,
		FeatureReferences: refs,
		CompositeHash:     record.CompositeHash,
		ModelVersion:      record.ModelVersion,
		ModelHash:         record.ModelHash,
		Score:             record.Score,
		Confidence:        record.Confidence,
		Status:            record.Status,
		ReasonCodes:       record.ReasonCodes,
		ComputedAt:        record.ComputedAt.Unix(),
		BlockHeight:       record.BlockHeight,
		ComputedBy:        record.ComputedBy,
		ConsensusVotes:    votes,
		Finalized:         record.Finalized,
	}

	if record.FinalizedAt != nil {
		ts := record.FinalizedAt.Unix()
		rs.FinalizedAt = &ts
	}
	if record.FinalizedAtBlock != nil {
		rs.FinalizedAtBlock = record.FinalizedAtBlock
	}

	bz, err := json.Marshal(&rs)
	if err != nil {
		return err
	}

	store.Set(types.DerivedFeatureRecordKey(record.RecordID), bz)

	// Add to account index
	addr, err := sdk.AccAddressFromBech32(record.AccountAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	k.addRecordToAccountIndex(ctx, addr.Bytes(), record.BlockHeight, record.RecordID)

	return nil
}

// GetDerivedFeatureRecord retrieves a derived feature verification record
func (k Keeper) GetDerivedFeatureRecord(ctx sdk.Context, recordID string) (types.DerivedFeatureVerificationRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.DerivedFeatureRecordKey(recordID))
	if bz == nil {
		return types.DerivedFeatureVerificationRecord{}, false
	}

	var rs derivedFeatureRecordStore
	if err := json.Unmarshal(bz, &rs); err != nil {
		return types.DerivedFeatureVerificationRecord{}, false
	}

	// Convert from storage format
	refs := make([]types.DerivedFeatureReference, len(rs.FeatureReferences))
	for i, ref := range rs.FeatureReferences {
		refs[i] = featureReferenceFromStore(ref)
	}

	votes := make([]types.ConsensusVote, len(rs.ConsensusVotes))
	for i, vote := range rs.ConsensusVotes {
		votes[i] = consensusVoteFromStore(vote)
	}

	record := types.DerivedFeatureVerificationRecord{
		RecordID:          rs.RecordID,
		AccountAddress:    rs.AccountAddress,
		Version:           rs.Version,
		RequestID:         rs.RequestID,
		FeatureReferences: refs,
		CompositeHash:     rs.CompositeHash,
		ModelVersion:      rs.ModelVersion,
		ModelHash:         rs.ModelHash,
		Score:             rs.Score,
		Confidence:        rs.Confidence,
		Status:            rs.Status,
		ReasonCodes:       rs.ReasonCodes,
		ComputedAt:        time.Unix(rs.ComputedAt, 0),
		BlockHeight:       rs.BlockHeight,
		ComputedBy:        rs.ComputedBy,
		ConsensusVotes:    votes,
		Finalized:         rs.Finalized,
	}

	if rs.FinalizedAt != nil {
		t := time.Unix(*rs.FinalizedAt, 0)
		record.FinalizedAt = &t
	}
	if rs.FinalizedAtBlock != nil {
		record.FinalizedAtBlock = rs.FinalizedAtBlock
	}

	return record, true
}

// GetDerivedFeatureRecordsByAccount retrieves all records for an account
func (k Keeper) GetDerivedFeatureRecordsByAccount(ctx sdk.Context, address sdk.AccAddress) []types.DerivedFeatureVerificationRecord {
	var records []types.DerivedFeatureVerificationRecord

	store := ctx.KVStore(k.skey)
	prefix := types.DerivedFeatureRecordByAccountPrefixKey(address.Bytes())
	iter := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var recordID string
		if err := json.Unmarshal(iter.Value(), &recordID); err != nil {
			continue
		}

		if record, found := k.GetDerivedFeatureRecord(ctx, recordID); found {
			records = append(records, record)
		}
	}

	return records
}

// GetLatestDerivedFeatureRecord returns the most recent record for an account
func (k Keeper) GetLatestDerivedFeatureRecord(ctx sdk.Context, address sdk.AccAddress) (types.DerivedFeatureVerificationRecord, bool) {
	records := k.GetDerivedFeatureRecordsByAccount(ctx, address)

	if len(records) == 0 {
		return types.DerivedFeatureVerificationRecord{}, false
	}

	// Find the record with the highest block height
	latest := records[0]
	for i := 1; i < len(records); i++ {
		if records[i].BlockHeight > latest.BlockHeight {
			latest = records[i]
		}
	}

	return latest, true
}

// DeleteDerivedFeatureRecord deletes a verification record
func (k Keeper) DeleteDerivedFeatureRecord(ctx sdk.Context, recordID string) error {
	record, found := k.GetDerivedFeatureRecord(ctx, recordID)
	if !found {
		return types.ErrVerificationRequestNotFound.Wrapf("record not found: %s", recordID)
	}

	store := ctx.KVStore(k.skey)
	store.Delete(types.DerivedFeatureRecordKey(recordID))

	// Remove from account index
	addr, err := sdk.AccAddressFromBech32(record.AccountAddress)
	if err == nil {
		k.removeRecordFromAccountIndex(ctx, addr.Bytes(), record.BlockHeight, recordID)
	}

	return nil
}

// AddConsensusVote adds a consensus vote to a record
func (k Keeper) AddConsensusVote(ctx sdk.Context, recordID string, vote types.ConsensusVote) error {
	record, found := k.GetDerivedFeatureRecord(ctx, recordID)
	if !found {
		return types.ErrVerificationRequestNotFound.Wrapf("record not found: %s", recordID)
	}

	if record.Finalized {
		return types.ErrInvalidVerificationResult.Wrap("record already finalized")
	}

	// Check if validator already voted
	for _, existingVote := range record.ConsensusVotes {
		if existingVote.ValidatorAddress == vote.ValidatorAddress {
			return types.ErrInvalidVerificationResult.Wrap("validator already voted")
		}
	}

	record.AddConsensusVote(vote)

	return k.SetDerivedFeatureRecord(ctx, record)
}

// FinalizeRecord finalizes a verification record
func (k Keeper) FinalizeRecord(ctx sdk.Context, recordID string) error {
	record, found := k.GetDerivedFeatureRecord(ctx, recordID)
	if !found {
		return types.ErrVerificationRequestNotFound.Wrapf("record not found: %s", recordID)
	}

	if record.Finalized {
		return types.ErrInvalidVerificationResult.Wrap("record already finalized")
	}

	record.Finalize(ctx.BlockTime(), ctx.BlockHeight())

	return k.SetDerivedFeatureRecord(ctx, record)
}

// Helper functions for account index management

func (k Keeper) addRecordToAccountIndex(ctx sdk.Context, address []byte, blockHeight int64, recordID string) {
	store := ctx.KVStore(k.skey)
	key := types.DerivedFeatureRecordByAccountKey(address, blockHeight)

	bz, _ := json.Marshal(recordID) //nolint:errchkjson // string cannot fail to marshal
	store.Set(key, bz)
}

func (k Keeper) removeRecordFromAccountIndex(ctx sdk.Context, address []byte, blockHeight int64, recordID string) {
	store := ctx.KVStore(k.skey)
	key := types.DerivedFeatureRecordByAccountKey(address, blockHeight)
	store.Delete(key)
}

// ============================================================================
// Derived Feature Processing (VE-217)
// ============================================================================

// CreateDerivedFeatureRecord creates a new verification record with derived features
func (k Keeper) CreateDerivedFeatureRecord(
	ctx sdk.Context,
	address sdk.AccAddress,
	requestID string,
	modelVersion string,
	modelHash string,
	validatorAddress string,
) (*types.DerivedFeatureVerificationRecord, error) {
	recordID := generateRecordID(address.String(), ctx.BlockHeight(), requestID)

	record := types.NewDerivedFeatureVerificationRecord(
		recordID,
		address.String(),
		requestID,
		modelVersion,
		modelHash,
		ctx.BlockTime(),
		ctx.BlockHeight(),
		validatorAddress,
	)

	return record, nil
}

// AddFeatureReferenceToRecord adds a derived feature reference to a record
func (k Keeper) AddFeatureReferenceToRecord(
	record *types.DerivedFeatureVerificationRecord,
	featureType string,
	featureHash []byte,
	sourceScopeID string,
	envelopeID string,
	fieldKey string,
	weight uint32,
) {
	ref := types.NewDerivedFeatureReference(featureType, featureHash, sourceScopeID, weight)
	ref.EnvelopeID = envelopeID
	ref.FieldKey = fieldKey
	record.AddFeatureReference(ref)
}

// Helper function to generate record IDs
func generateRecordID(address string, blockHeight int64, requestID string) string {
	// Use a deterministic ID based on input parameters
	// In production, this could use a hash of the inputs
	return address[:8] + "-" + requestID
}

// GetFaceEmbeddingHash retrieves the face embedding hash for an account
// Returns the hash from the most recent active embedding envelope
func (k Keeper) GetFaceEmbeddingHash(ctx sdk.Context, address sdk.AccAddress) ([]byte, bool) {
	envelope, found := k.GetActiveEmbeddingEnvelope(ctx, address, types.EmbeddingTypeFace)
	if !found {
		return nil, false
	}
	return envelope.EmbeddingHash, true
}

// GetDocumentFaceEmbeddingHash retrieves the document face embedding hash
func (k Keeper) GetDocumentFaceEmbeddingHash(ctx sdk.Context, address sdk.AccAddress) ([]byte, bool) {
	envelope, found := k.GetActiveEmbeddingEnvelope(ctx, address, types.EmbeddingTypeDocumentFace)
	if !found {
		return nil, false
	}
	return envelope.EmbeddingHash, true
}

// VerifyEmbeddingMatch verifies that a computed embedding hash matches the stored hash
// This is used by validators to verify derived features
func (k Keeper) VerifyEmbeddingMatch(ctx sdk.Context, address sdk.AccAddress, embeddingType types.EmbeddingType, computedHash []byte) (bool, error) {
	envelope, found := k.GetActiveEmbeddingEnvelope(ctx, address, embeddingType)
	if !found {
		return false, types.ErrScopeNotFound.Wrapf("no active envelope for type: %s", embeddingType)
	}

	// Compare hashes
	if len(envelope.EmbeddingHash) != len(computedHash) {
		return false, nil
	}

	// Use constant-time comparison
	var result byte
	for i := 0; i < len(envelope.EmbeddingHash); i++ {
		result |= envelope.EmbeddingHash[i] ^ computedHash[i]
	}

	return result == 0, nil
}

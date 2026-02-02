package keeper

import (
	"encoding/json"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Identity Score Keeper Interface
// ============================================================================

// IdentityScoreKeeper defines the interface for other modules to interact with identity scores
// This interface allows modules like Marketplace and MFA to gate actions based on identity status
type IdentityScoreKeeper interface {
	// GetScore returns the current score, status, and whether a record was found
	GetScore(ctx sdk.Context, addr string) (uint32, types.AccountStatus, bool)

	// IsScoreAboveThreshold checks if an account's score meets or exceeds a threshold
	// Returns false if the account doesn't exist or isn't verified
	IsScoreAboveThreshold(ctx sdk.Context, addr string, threshold uint32) bool

	// GetAccountTier returns the account's tier (0-3) or an error if not found
	GetAccountTier(ctx sdk.Context, addr string) (int, error)

	// GetVerificationStatus returns the verification status for an account
	GetVerificationStatus(ctx sdk.Context, addr string) types.AccountStatus

	// CheckEligibility checks if an account is eligible for an offering type
	CheckEligibility(ctx sdk.Context, addr string, offeringType types.OfferingType) types.EligibilityResult
}

// Ensure Keeper implements IdentityScoreKeeper
var _ IdentityScoreKeeper = Keeper{}

// ============================================================================
// Score Storage
// ============================================================================

// scoreStore is the stored format of an identity score
type scoreStore struct {
	AccountAddress   string `json:"account_address"`
	Score            uint32 `json:"score"`
	Status           string `json:"status"`
	ModelVersion     string `json:"model_version"`
	ComputedAt       int64  `json:"computed_at"`
	ExpiresAt        *int64 `json:"expires_at,omitempty"`
	VerificationHash []byte `json:"verification_hash,omitempty"`
	BlockHeight      int64  `json:"block_height"`
}

// scoreHistoryStore is the stored format of a score history entry
type scoreHistoryStore struct {
	Score            uint32 `json:"score"`
	Status           string `json:"status"`
	ModelVersion     string `json:"model_version"`
	ComputedAt       int64  `json:"computed_at"`
	BlockHeight      int64  `json:"block_height"`
	Reason           string `json:"reason,omitempty"`
	VerificationHash []byte `json:"verification_hash,omitempty"`
}

// Error message constants for score operations
const errMsgNoScoreFound = "no score found for %s"

// ScoreDetails contains optional details for setting a score
type ScoreDetails struct {
	Status           types.AccountStatus
	ModelVersion     string
	VerificationHash []byte
	ExpiresAt        *time.Time
	Reason           string
}

// DefaultScoreDetails returns default score details for verified status
func DefaultScoreDetails(modelVersion string) ScoreDetails {
	return ScoreDetails{
		Status:       types.AccountStatusVerified,
		ModelVersion: modelVersion,
	}
}

// ============================================================================
// Score Management Methods
// ============================================================================

// SetScore sets the identity score for an account and records it in history
func (k Keeper) SetScore(
	ctx sdk.Context,
	accountAddr string,
	score uint32,
	modelVersion string,
) error {
	return k.SetScoreWithDetails(ctx, accountAddr, score, DefaultScoreDetails(modelVersion))
}

// SetScoreWithDetails sets the identity score with full details
func (k Keeper) SetScoreWithDetails(
	ctx sdk.Context,
	accountAddr string,
	score uint32,
	details ScoreDetails,
) error {
	// Validate score
	if score > types.MaxScore {
		return types.ErrInvalidScore.Wrapf("score %d exceeds maximum %d", score, types.MaxScore)
	}

	// Validate account address
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	// Get current time and block height
	now := ctx.BlockTime()
	blockHeight := ctx.BlockHeight()

	// Create score store entry
	ss := scoreStore{
		AccountAddress:   accountAddr,
		Score:            score,
		Status:           string(details.Status),
		ModelVersion:     details.ModelVersion,
		ComputedAt:       now.Unix(),
		VerificationHash: details.VerificationHash,
		BlockHeight:      blockHeight,
	}

	if details.ExpiresAt != nil {
		expTs := details.ExpiresAt.Unix()
		ss.ExpiresAt = &expTs
	}

	// Store the score
	bz, err := json.Marshal(&ss)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ScoreKey(address.Bytes()), bz)

	// Add to score history
	historyEntry := scoreHistoryStore{
		Score:            score,
		Status:           string(details.Status),
		ModelVersion:     details.ModelVersion,
		ComputedAt:       now.Unix(),
		BlockHeight:      blockHeight,
		Reason:           details.Reason,
		VerificationHash: details.VerificationHash,
	}

	historyBz, err := json.Marshal(&historyEntry)
	if err != nil {
		return err
	}

	store.Set(types.ScoreHistoryKey(address.Bytes(), now.Unix(), blockHeight), historyBz)

	// Update the identity record if it exists
	record, found := k.GetIdentityRecord(ctx, address)
	if found {
		record.CurrentScore = score
		record.ScoreVersion = details.ModelVersion
		record.UpdatedAt = now
		if details.Status == types.AccountStatusVerified {
			record.LastVerifiedAt = &now
		}
		record.UpdateTier()
		if err := k.SetIdentityRecord(ctx, record); err != nil {
			return err
		}
	}

	return nil
}

// GetVEIDScore returns the VEID score for an account address.
// This method is used by other modules (e.g., MFA) to get the score.
// Implements the VEIDKeeper interface expected by x/mfa.
func (k Keeper) GetVEIDScore(ctx sdk.Context, address sdk.AccAddress) (uint32, bool) {
	score, _, found := k.GetScore(ctx, address.String())
	return score, found
}

// GetScore returns the score, status, and whether a score was found for an account
// Implements IdentityScoreKeeper interface
func (k Keeper) GetScore(ctx sdk.Context, addr string) (uint32, types.AccountStatus, bool) {
	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return 0, types.AccountStatusUnknown, false
	}

	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ScoreKey(address.Bytes()))
	if bz == nil {
		// Check if identity record exists with a score
		record, found := k.GetIdentityRecord(ctx, address)
		if found && record.CurrentScore > 0 {
			status := types.AccountStatusFromVerificationStatus(types.VerificationStatusVerified)
			if record.Locked {
				status = types.AccountStatusRejected
			}
			return record.CurrentScore, status, true
		}
		return 0, types.AccountStatusUnknown, false
	}

	var ss scoreStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return 0, types.AccountStatusUnknown, false
	}

	// Check if expired
	if ss.ExpiresAt != nil && time.Now().Unix() > *ss.ExpiresAt {
		return ss.Score, types.AccountStatusExpired, true
	}

	return ss.Score, types.AccountStatus(ss.Status), true
}

// GetIdentityScore returns the full IdentityScore struct for an account
func (k Keeper) GetIdentityScore(ctx sdk.Context, addr string) (*types.IdentityScore, bool) {
	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, false
	}

	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ScoreKey(address.Bytes()))
	if bz == nil {
		return nil, false
	}

	var ss scoreStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return nil, false
	}

	score := &types.IdentityScore{
		AccountAddress:   ss.AccountAddress,
		Score:            ss.Score,
		Status:           types.AccountStatus(ss.Status),
		ModelVersion:     ss.ModelVersion,
		ComputedAt:       time.Unix(ss.ComputedAt, 0),
		VerificationHash: ss.VerificationHash,
		BlockHeight:      ss.BlockHeight,
	}

	if ss.ExpiresAt != nil {
		t := time.Unix(*ss.ExpiresAt, 0)
		score.ExpiresAt = &t
	}

	return score, true
}

// GetScoreHistory returns the score history for an account
func (k Keeper) GetScoreHistory(ctx sdk.Context, accountAddr string) []types.ScoreHistoryEntry {
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return nil
	}

	return k.getScoreHistoryForAddress(ctx, address, 0)
}

// GetScoreHistoryPaginated returns paginated score history for an account
func (k Keeper) GetScoreHistoryPaginated(ctx sdk.Context, accountAddr string, limit uint32, offset uint32) []types.ScoreHistoryEntry {
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return nil
	}

	entries := k.getScoreHistoryForAddress(ctx, address, 0)

	// Apply pagination
	if offset >= uint32(len(entries)) {
		return []types.ScoreHistoryEntry{}
	}

	end := offset + limit
	if end > uint32(len(entries)) || limit == 0 {
		end = uint32(len(entries))
	}

	return entries[offset:end]
}

// getScoreHistoryForAddress retrieves score history from storage
func (k Keeper) getScoreHistoryForAddress(ctx sdk.Context, address sdk.AccAddress, limit uint32) []types.ScoreHistoryEntry {
	store := ctx.KVStore(k.skey)
	prefix := types.ScoreHistoryPrefixKey(address.Bytes())
	iter := storetypes.KVStoreReversePrefixIterator(store, prefix) // Reverse for newest first
	defer iter.Close()

	var entries []types.ScoreHistoryEntry
	count := uint32(0)

	for ; iter.Valid(); iter.Next() {
		if limit > 0 && count >= limit {
			break
		}

		var hs scoreHistoryStore
		if err := json.Unmarshal(iter.Value(), &hs); err != nil {
			continue
		}

		entry := types.ScoreHistoryEntry{
			Score:            hs.Score,
			Status:           types.AccountStatus(hs.Status),
			ModelVersion:     hs.ModelVersion,
			ComputedAt:       time.Unix(hs.ComputedAt, 0),
			BlockHeight:      hs.BlockHeight,
			Reason:           hs.Reason,
			VerificationHash: hs.VerificationHash,
		}

		entries = append(entries, entry)
		count++
	}

	return entries
}

// IsScoreAboveThreshold checks if an account's score meets or exceeds a threshold
// Implements IdentityScoreKeeper interface
func (k Keeper) IsScoreAboveThreshold(ctx sdk.Context, addr string, threshold uint32) bool {
	score, status, found := k.GetScore(ctx, addr)
	if !found {
		return false
	}

	// Must be verified to pass threshold checks
	if status != types.AccountStatusVerified {
		return false
	}

	return score >= threshold
}

// GetAccountTier returns the account's tier based on score and status
// Implements IdentityScoreKeeper interface
func (k Keeper) GetAccountTier(ctx sdk.Context, addr string) (int, error) {
	score, status, found := k.GetScore(ctx, addr)
	if !found {
		return types.TierUnverified, types.ErrIdentityRecordNotFound.Wrapf(errMsgNoScoreFound, addr)
	}

	return types.ComputeTierFromScoreValue(score, status), nil
}

// GetVerificationStatus returns the verification status for an account
// Implements IdentityScoreKeeper interface
func (k Keeper) GetVerificationStatus(ctx sdk.Context, addr string) types.AccountStatus {
	_, status, found := k.GetScore(ctx, addr)
	if !found {
		return types.AccountStatusUnknown
	}
	return status
}

// CheckEligibility checks if an account is eligible for an offering type
// Implements IdentityScoreKeeper interface
func (k Keeper) CheckEligibility(ctx sdk.Context, addr string, offeringType types.OfferingType) types.EligibilityResult {
	requirements := types.GetRequiredScopesForOffering(offeringType)
	result := types.EligibilityResult{
		Eligible:       false,
		AccountAddress: addr,
		OfferingType:   offeringType,
		RequiredScore:  requirements.MinimumScore,
		RequiresMFA:    requirements.RequiresMFA,
	}

	// Get score and status
	score, status, found := k.GetScore(ctx, addr)
	if !found {
		result.CurrentStatus = types.AccountStatusUnknown
		result.Reason = "Account not found or not verified"
		return result
	}

	result.CurrentScore = score
	result.CurrentStatus = status

	// Check verification status
	if status != types.AccountStatusVerified {
		result.Reason = "Account is not verified"
		return result
	}

	// Check score threshold
	if score < requirements.MinimumScore {
		result.Reason = "Score does not meet minimum requirement"
		return result
	}

	// Check required scopes
	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		result.Reason = "Invalid account address"
		return result
	}

	record, recordFound := k.GetIdentityRecord(ctx, address)
	if !recordFound {
		result.Reason = "Identity record not found"
		return result
	}

	// Check for missing scopes
	var missingScopes []types.ScopeType
	for _, requiredScope := range requirements.RequiredScopeTypes {
		if !record.HasVerifiedScope(requiredScope) {
			missingScopes = append(missingScopes, requiredScope)
		}
	}

	if len(missingScopes) > 0 {
		result.MissingScopes = missingScopes
		result.Reason = "Required scopes are not verified"
		return result
	}

	// All checks passed
	result.Eligible = true
	result.Reason = "Account meets all eligibility requirements"
	return result
}

// ============================================================================
// Tier-based Account Retrieval
// ============================================================================

// GetAccountsByTier returns all accounts at a specific tier
func (k Keeper) GetAccountsByTier(ctx sdk.Context, tier int) []types.AccountScoreSummary {
	var accounts []types.AccountScoreSummary

	k.WithIdentityRecords(ctx, func(record types.IdentityRecord) bool {
		score, status, found := k.GetScore(ctx, record.AccountAddress)
		if !found {
			score = record.CurrentScore
			status = types.AccountStatusVerified
		}

		accountTier := types.ComputeTierFromScoreValue(score, status)
		if accountTier == tier {
			summary := types.AccountScoreSummary{
				AccountAddress: record.AccountAddress,
				Score:          score,
				Status:         status,
				Tier:           accountTier,
				ModelVersion:   record.ScoreVersion,
				ComputedAt:     record.UpdatedAt,
			}
			accounts = append(accounts, summary)
		}
		return false
	})

	return accounts
}

// GetAccountsByScoreRange returns accounts within a score range
func (k Keeper) GetAccountsByScoreRange(ctx sdk.Context, minScore, maxScore uint32) []types.AccountScoreSummary {
	var accounts []types.AccountScoreSummary

	k.WithIdentityRecords(ctx, func(record types.IdentityRecord) bool {
		score, status, found := k.GetScore(ctx, record.AccountAddress)
		if !found {
			score = record.CurrentScore
			status = types.AccountStatusVerified
		}

		if score >= minScore && score <= maxScore {
			summary := types.AccountScoreSummary{
				AccountAddress: record.AccountAddress,
				Score:          score,
				Status:         status,
				Tier:           types.ComputeTierFromScoreValue(score, status),
				ModelVersion:   record.ScoreVersion,
				ComputedAt:     record.UpdatedAt,
			}
			accounts = append(accounts, summary)
		}
		return false
	})

	return accounts
}

// ============================================================================
// Score Expiration Management
// ============================================================================

// ExpireScore marks a score as expired
func (k Keeper) ExpireScore(ctx sdk.Context, accountAddr string, reason string) error {
	score, _, found := k.GetScore(ctx, accountAddr)
	if !found {
		return types.ErrIdentityRecordNotFound.Wrapf(errMsgNoScoreFound, accountAddr)
	}

	return k.SetScoreWithDetails(ctx, accountAddr, score, ScoreDetails{
		Status: types.AccountStatusExpired,
		Reason: reason,
	})
}

// RefreshScoreExpiration extends the expiration time for a score
func (k Keeper) RefreshScoreExpiration(ctx sdk.Context, accountAddr string, newExpiration time.Time) error {
	identityScore, found := k.GetIdentityScore(ctx, accountAddr)
	if !found {
		return types.ErrIdentityRecordNotFound.Wrapf(errMsgNoScoreFound, accountAddr)
	}

	return k.SetScoreWithDetails(ctx, accountAddr, identityScore.Score, ScoreDetails{
		Status:           identityScore.Status,
		ModelVersion:     identityScore.ModelVersion,
		VerificationHash: identityScore.VerificationHash,
		ExpiresAt:        &newExpiration,
		Reason:           "expiration refreshed",
	})
}

// ============================================================================
// Score Statistics
// ============================================================================

// GetScoreStatistics computes aggregate statistics for all scores
func (k Keeper) GetScoreStatistics(ctx sdk.Context) types.ScoreStatistics {
	stats := types.ScoreStatistics{
		TierCounts:   make(map[int]uint64),
		StatusCounts: make(map[types.AccountStatus]uint64),
		ComputedAt:   ctx.BlockTime(),
	}

	var totalScore uint64
	var verifiedCount uint64
	var scores []uint32

	k.WithIdentityRecords(ctx, func(record types.IdentityRecord) bool {
		score, status, found := k.GetScore(ctx, record.AccountAddress)
		if !found {
			score = record.CurrentScore
			status = types.AccountStatusUnknown
		}

		stats.TotalAccounts++
		stats.StatusCounts[status]++

		tier := types.ComputeTierFromScoreValue(score, status)
		stats.TierCounts[tier]++

		if status == types.AccountStatusVerified {
			totalScore += uint64(score)
			verifiedCount++
			scores = append(scores, score)
		}

		return false
	})

	// Calculate average
	if verifiedCount > 0 {
		stats.AverageScore = float64(totalScore) / float64(verifiedCount)
	}

	// Calculate median (simple approach for now)
	if len(scores) > 0 {
		// Sort scores
		for i := 0; i < len(scores)-1; i++ {
			for j := i + 1; j < len(scores); j++ {
				if scores[i] > scores[j] {
					scores[i], scores[j] = scores[j], scores[i]
				}
			}
		}
		mid := len(scores) / 2
		if len(scores)%2 == 0 {
			stats.MedianScore = (scores[mid-1] + scores[mid]) / 2
		} else {
			stats.MedianScore = scores[mid]
		}
	}

	return stats
}

// maxBlockRangeForMetrics is the maximum number of blocks to iterate for metrics queries.
// This prevents performance issues from large range queries.
const maxBlockRangeForMetrics = 10000

// GetValidatorVerificationCount returns the count of verifications performed by a validator
// in the given block height range. For performance, the range is capped at maxBlockRangeForMetrics blocks.
func (k Keeper) GetValidatorVerificationCount(ctx sdk.Context, validatorAddr string, startHeight, endHeight int64) int64 {
	var count int64

	// Cap the range to prevent performance issues
	if endHeight-startHeight > maxBlockRangeForMetrics {
		startHeight = endHeight - maxBlockRangeForMetrics
	}

	// Iterate through verification metrics in the block range
	for height := startHeight; height <= endHeight; height++ {
		metrics := k.GetBlockMetrics(ctx, height)
		for _, m := range metrics {
			if m.ValidatorAddress == validatorAddr {
				count++
			}
		}
	}

	return count
}

// GetValidatorAverageVerificationScore returns the average verification score for a validator
// in the given block height range. For performance, the range is capped at maxBlockRangeForMetrics blocks.
func (k Keeper) GetValidatorAverageVerificationScore(ctx sdk.Context, validatorAddr string, startHeight, endHeight int64) int64 {
	var totalScore int64
	var count int64

	// Cap the range to prevent performance issues
	if endHeight-startHeight > maxBlockRangeForMetrics {
		startHeight = endHeight - maxBlockRangeForMetrics
	}

	// Iterate through verification metrics in the block range
	for height := startHeight; height <= endHeight; height++ {
		metrics := k.GetBlockMetrics(ctx, height)
		for _, m := range metrics {
			if m.ValidatorAddress == validatorAddr {
				totalScore += int64(m.ComputedScore)
				count++
			}
		}
	}

	if count == 0 {
		return 0
	}

	return totalScore / count
}

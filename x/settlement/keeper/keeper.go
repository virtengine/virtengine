package keeper

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

// IKeeper defines the interface for the settlement keeper
type IKeeper interface {
	// Escrow management
	CreateEscrow(ctx sdk.Context, orderID string, depositor sdk.AccAddress, amount sdk.Coins, expiresIn time.Duration, conditions []types.ReleaseCondition) (string, error)
	ActivateEscrow(ctx sdk.Context, escrowID string, leaseID string, recipient sdk.AccAddress) error
	ReleaseEscrow(ctx sdk.Context, escrowID string, reason string) error
	RefundEscrow(ctx sdk.Context, escrowID string, reason string) error
	DisputeEscrow(ctx sdk.Context, escrowID string, reason string) error
	GetEscrow(ctx sdk.Context, escrowID string) (types.EscrowAccount, bool)
	GetEscrowByOrder(ctx sdk.Context, orderID string) (types.EscrowAccount, bool)
	SetEscrow(ctx sdk.Context, escrow types.EscrowAccount) error

	// Settlement management
	SettleOrder(ctx sdk.Context, orderID string, usageRecordIDs []string, isFinal bool) (*types.SettlementRecord, error)
	GetSettlement(ctx sdk.Context, settlementID string) (types.SettlementRecord, bool)
	GetSettlementsByOrder(ctx sdk.Context, orderID string) []types.SettlementRecord

	// Usage records
	RecordUsage(ctx sdk.Context, record *types.UsageRecord) error
	AcknowledgeUsage(ctx sdk.Context, usageID string, customerSignature []byte) error
	GetUsageRecord(ctx sdk.Context, usageID string) (types.UsageRecord, bool)
	GetUsageRecordsByOrder(ctx sdk.Context, orderID string) []types.UsageRecord
	GetUnsettledUsageRecords(ctx sdk.Context, orderID string) []types.UsageRecord

	// Rewards
	DistributeStakingRewards(ctx sdk.Context, epoch uint64) (*types.RewardDistribution, error)
	DistributeProviderRewards(ctx sdk.Context, usageRecords []types.UsageRecord) (*types.RewardDistribution, error)
	DistributeUsageRewards(ctx sdk.Context, usageRecords []types.UsageRecord) (*types.RewardDistribution, error)
	DistributeUsageRewardsForSettlement(ctx sdk.Context, settlementID string, usageRecords []types.UsageRecord) (*types.RewardDistribution, error)
	DistributeVerificationRewards(ctx sdk.Context, verificationResults []VerificationResult) (*types.RewardDistribution, error)
	AddClaimableReward(ctx sdk.Context, address sdk.AccAddress, entry types.RewardEntry) error
	GetClaimableRewards(ctx sdk.Context, address sdk.AccAddress) (types.ClaimableRewards, bool)
	ClaimRewards(ctx sdk.Context, claimer sdk.AccAddress, source string) (sdk.Coins, error)
	GetRewardDistribution(ctx sdk.Context, distributionID string) (types.RewardDistribution, bool)
	GetRewardsByEpoch(ctx sdk.Context, epoch uint64) []types.RewardDistribution

	// Payout management
	ExecutePayout(ctx sdk.Context, invoiceID string, settlementID string) (*types.PayoutRecord, error)
	GetPayout(ctx sdk.Context, payoutID string) (types.PayoutRecord, bool)
	GetPayoutByInvoice(ctx sdk.Context, invoiceID string) (types.PayoutRecord, bool)
	GetPayoutBySettlement(ctx sdk.Context, settlementID string) (types.PayoutRecord, bool)
	HoldPayout(ctx sdk.Context, payoutID string, disputeID string, reason string) error
	ReleasePayoutHold(ctx sdk.Context, payoutID string) error
	RefundPayout(ctx sdk.Context, payoutID string, reason string) error
	ProcessPendingPayouts(ctx sdk.Context) error
	RetryFailedPayouts(ctx sdk.Context) error
	SetPayout(ctx sdk.Context, payout types.PayoutRecord) error
	WithPayouts(ctx sdk.Context, fn func(types.PayoutRecord) bool)
	WithPayoutsByState(ctx sdk.Context, state types.PayoutState, fn func(types.PayoutRecord) bool)

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error
	GetAuthority() string

	// Iterators
	WithEscrows(ctx sdk.Context, fn func(types.EscrowAccount) bool)
	WithEscrowsByState(ctx sdk.Context, state types.EscrowState, fn func(types.EscrowAccount) bool)
	WithSettlements(ctx sdk.Context, fn func(types.SettlementRecord) bool)
	WithUsageRecords(ctx sdk.Context, fn func(types.UsageRecord) bool)
	WithRewardDistributions(ctx sdk.Context, fn func(types.RewardDistribution) bool)
	WithClaimableRewards(ctx sdk.Context, fn func(types.ClaimableRewards) bool)

	// Genesis sequence setters
	SetNextEscrowSequence(ctx sdk.Context, seq uint64)
	SetNextSettlementSequence(ctx sdk.Context, seq uint64)
	SetNextUsageSequence(ctx sdk.Context, seq uint64)
	SetNextDistributionSequence(ctx sdk.Context, seq uint64)
	SetNextPayoutSequence(ctx sdk.Context, seq uint64)

	// Block hooks
	ProcessExpiredEscrows(ctx sdk.Context) error
	SatisfyTimelockConditions(ctx sdk.Context) error
	AutoSettle(ctx sdk.Context) error
	EndBlockerRewards(ctx sdk.Context) error

	// Storage setters
	SetSettlement(ctx sdk.Context, settlement types.SettlementRecord) error
	SetUsageRecord(ctx sdk.Context, usage types.UsageRecord) error
	SetRewardDistribution(ctx sdk.Context, dist types.RewardDistribution) error
	SetClaimableRewards(ctx sdk.Context, addr sdk.AccAddress, rewards types.ClaimableRewards) error

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// VerificationResult represents a verification result for reward distribution
type VerificationResult struct {
	ValidatorAddress string
	AccountAddress   string
	Score            uint32
	BlockHeight      int64
}

// Keeper of the settlement store
type Keeper struct {
	skey          storetypes.StoreKey
	cdc           codec.BinaryCodec
	bankKeeper    BankKeeper
	billingKeeper BillingKeeper

	// The address capable of executing a MsgUpdateParams message.
	// This should be the x/gov module account.
	authority string
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// NewKeeper creates and returns an instance for settlement keeper
func NewKeeper(cdc codec.BinaryCodec, skey storetypes.StoreKey, bankKeeper BankKeeper, authority string) Keeper {
	return Keeper{
		cdc:           cdc,
		skey:          skey,
		bankKeeper:    bankKeeper,
		billingKeeper: nil,
		authority:     authority,
	}
}

// SetBillingKeeper configures the billing integration keeper.
func (k *Keeper) SetBillingKeeper(billingKeeper BillingKeeper) {
	k.billingKeeper = billingKeeper
}

// Codec returns keeper codec
func (k Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns store key
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.skey
}

// GetAuthority returns the module's authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// ============================================================================
// Sequence Management
// ============================================================================

func (k Keeper) getNextSequence(ctx sdk.Context, key []byte) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(key)
	if bz == nil {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) setNextSequence(ctx sdk.Context, key []byte, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	store.Set(key, bz)
}

func (k Keeper) getNextEscrowSequence(ctx sdk.Context) uint64 {
	return k.getNextSequence(ctx, types.EscrowSequenceKey())
}

func (k Keeper) incrementEscrowSequence(ctx sdk.Context) uint64 {
	seq := k.getNextEscrowSequence(ctx)
	k.setNextSequence(ctx, types.EscrowSequenceKey(), seq+1)
	return seq
}

func (k Keeper) getNextSettlementSequence(ctx sdk.Context) uint64 {
	return k.getNextSequence(ctx, types.SettlementSequenceKey())
}

func (k Keeper) incrementSettlementSequence(ctx sdk.Context) uint64 {
	seq := k.getNextSettlementSequence(ctx)
	k.setNextSequence(ctx, types.SettlementSequenceKey(), seq+1)
	return seq
}

func (k Keeper) getNextDistributionSequence(ctx sdk.Context) uint64 {
	return k.getNextSequence(ctx, types.DistributionSequenceKey())
}

func (k Keeper) incrementDistributionSequence(ctx sdk.Context) uint64 {
	seq := k.getNextDistributionSequence(ctx)
	k.setNextSequence(ctx, types.DistributionSequenceKey(), seq+1)
	return seq
}

func (k Keeper) getNextUsageSequence(ctx sdk.Context) uint64 {
	return k.getNextSequence(ctx, types.UsageSequenceKey())
}

func (k Keeper) incrementUsageSequence(ctx sdk.Context) uint64 {
	seq := k.getNextUsageSequence(ctx)
	k.setNextSequence(ctx, types.UsageSequenceKey(), seq+1)
	return seq
}

// ============================================================================
// Parameters
// ============================================================================

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey(), bz)
	return nil
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ParamsKey())
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}

	return params
}

// getPlatformFeeRate parses and returns the platform fee rate
func (k Keeper) getPlatformFeeRate(ctx sdk.Context) sdkmath.LegacyDec {
	params := k.GetParams(ctx)
	rate, err := sdkmath.LegacyNewDecFromStr(params.PlatformFeeRate)
	if err != nil {
		return sdkmath.LegacyNewDecWithPrec(5, 2) // Default 5%
	}
	return rate
}

// getValidatorFeeRate parses and returns the validator fee rate
func (k Keeper) getValidatorFeeRate(ctx sdk.Context) sdkmath.LegacyDec {
	params := k.GetParams(ctx)
	rate, err := sdkmath.LegacyNewDecFromStr(params.ValidatorFeeRate)
	if err != nil {
		return sdkmath.LegacyNewDecWithPrec(1, 2) // Default 1%
	}
	return rate
}

// ============================================================================
// Escrow Storage
// ============================================================================

// SetEscrow saves an escrow account to the store
func (k Keeper) SetEscrow(ctx sdk.Context, escrow types.EscrowAccount) error {
	if err := escrow.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(&escrow)
	if err != nil {
		return err
	}

	// Store by escrow ID
	store.Set(types.EscrowKey(escrow.EscrowID), bz)

	// Store by order ID
	store.Set(types.EscrowByOrderKey(escrow.OrderID), []byte(escrow.EscrowID))

	// Store by state
	store.Set(types.EscrowByStateKey(escrow.State, escrow.EscrowID), []byte{})

	return nil
}

// GetEscrow retrieves an escrow account by ID
func (k Keeper) GetEscrow(ctx sdk.Context, escrowID string) (types.EscrowAccount, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.EscrowKey(escrowID))
	if bz == nil {
		return types.EscrowAccount{}, false
	}

	var escrow types.EscrowAccount
	if err := json.Unmarshal(bz, &escrow); err != nil {
		return types.EscrowAccount{}, false
	}

	return escrow, true
}

// GetEscrowByOrder retrieves an escrow account by order ID
func (k Keeper) GetEscrowByOrder(ctx sdk.Context, orderID string) (types.EscrowAccount, bool) {
	store := ctx.KVStore(k.skey)
	escrowID := store.Get(types.EscrowByOrderKey(orderID))
	if escrowID == nil {
		return types.EscrowAccount{}, false
	}

	return k.GetEscrow(ctx, string(escrowID))
}

// updateEscrowState updates the state index for an escrow
func (k Keeper) updateEscrowState(ctx sdk.Context, escrow types.EscrowAccount, oldState types.EscrowState) {
	store := ctx.KVStore(k.skey)

	// Remove old state index
	store.Delete(types.EscrowByStateKey(oldState, escrow.EscrowID))

	// Add new state index
	store.Set(types.EscrowByStateKey(escrow.State, escrow.EscrowID), []byte{})
}

// WithEscrows iterates over all escrows
func (k Keeper) WithEscrows(ctx sdk.Context, fn func(types.EscrowAccount) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixEscrow)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var escrow types.EscrowAccount
		if err := json.Unmarshal(iter.Value(), &escrow); err != nil {
			continue
		}
		if fn(escrow) {
			break
		}
	}
}

// WithEscrowsByState iterates over escrows in a specific state
func (k Keeper) WithEscrowsByState(ctx sdk.Context, state types.EscrowState, fn func(types.EscrowAccount) bool) {
	store := ctx.KVStore(k.skey)
	prefix := types.EscrowByStatePrefixKey(state)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// Extract escrow ID from key
		key := iter.Key()
		escrowID := string(key[len(prefix):])

		escrow, found := k.GetEscrow(ctx, escrowID)
		if !found {
			continue
		}
		if fn(escrow) {
			break
		}
	}
}

// ============================================================================
// Settlement Storage
// ============================================================================

// SetSettlement saves a settlement record to the store
func (k Keeper) SetSettlement(ctx sdk.Context, settlement types.SettlementRecord) error {
	if err := settlement.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(&settlement)
	if err != nil {
		return err
	}

	// Store by settlement ID
	store.Set(types.SettlementKey(settlement.SettlementID), bz)

	// Store by order ID (append to list)
	k.appendSettlementToOrder(ctx, settlement.OrderID, settlement.SettlementID)

	// Store by escrow ID
	store.Set(types.SettlementByEscrowKey(settlement.EscrowID), []byte(settlement.SettlementID))

	return nil
}

// appendSettlementToOrder appends a settlement ID to an order's settlement list
func (k Keeper) appendSettlementToOrder(ctx sdk.Context, orderID, settlementID string) {
	store := ctx.KVStore(k.skey)
	key := types.SettlementByOrderKey(orderID)

	settlementIDs := make([]string, 0, 1)
	bz := store.Get(key)
	if bz != nil {
		_ = json.Unmarshal(bz, &settlementIDs)
	}

	settlementIDs = append(settlementIDs, settlementID)
	bz, _ = json.Marshal(settlementIDs) //nolint:errchkjson // string slice cannot fail to marshal
	store.Set(key, bz)
}

// GetSettlement retrieves a settlement record by ID
func (k Keeper) GetSettlement(ctx sdk.Context, settlementID string) (types.SettlementRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SettlementKey(settlementID))
	if bz == nil {
		return types.SettlementRecord{}, false
	}

	var settlement types.SettlementRecord
	if err := json.Unmarshal(bz, &settlement); err != nil {
		return types.SettlementRecord{}, false
	}

	return settlement, true
}

// GetSettlementsByOrder retrieves all settlements for an order
func (k Keeper) GetSettlementsByOrder(ctx sdk.Context, orderID string) []types.SettlementRecord {
	store := ctx.KVStore(k.skey)
	key := types.SettlementByOrderKey(orderID)

	bz := store.Get(key)
	if bz == nil {
		return []types.SettlementRecord{}
	}

	var settlementIDs []string
	if err := json.Unmarshal(bz, &settlementIDs); err != nil {
		return []types.SettlementRecord{}
	}

	settlements := make([]types.SettlementRecord, 0, len(settlementIDs))
	for _, id := range settlementIDs {
		if settlement, found := k.GetSettlement(ctx, id); found {
			settlements = append(settlements, settlement)
		}
	}

	return settlements
}

// WithSettlements iterates over all settlements
func (k Keeper) WithSettlements(ctx sdk.Context, fn func(types.SettlementRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixSettlement)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var settlement types.SettlementRecord
		if err := json.Unmarshal(iter.Value(), &settlement); err != nil {
			continue
		}
		if fn(settlement) {
			break
		}
	}
}

// ============================================================================
// Usage Record Storage
// ============================================================================

// SetUsageRecord saves a usage record to the store
func (k Keeper) SetUsageRecord(ctx sdk.Context, usage types.UsageRecord) error {
	if err := usage.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(&usage)
	if err != nil {
		return err
	}

	// Store by usage ID
	store.Set(types.UsageRecordKey(usage.UsageID), bz)

	// Append to order's usage list
	k.appendUsageToOrder(ctx, usage.OrderID, usage.UsageID)

	return nil
}

// appendUsageToOrder appends a usage ID to an order's usage list
func (k Keeper) appendUsageToOrder(ctx sdk.Context, orderID, usageID string) {
	store := ctx.KVStore(k.skey)
	key := types.UsageByOrderKey(orderID)

	usageIDs := make([]string, 0, 1)
	bz := store.Get(key)
	if bz != nil {
		_ = json.Unmarshal(bz, &usageIDs)
	}

	usageIDs = append(usageIDs, usageID)
	bz, _ = json.Marshal(usageIDs) //nolint:errchkjson // string slice cannot fail to marshal
	store.Set(key, bz)
}

// GetUsageRecord retrieves a usage record by ID
func (k Keeper) GetUsageRecord(ctx sdk.Context, usageID string) (types.UsageRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.UsageRecordKey(usageID))
	if bz == nil {
		return types.UsageRecord{}, false
	}

	var usage types.UsageRecord
	if err := json.Unmarshal(bz, &usage); err != nil {
		return types.UsageRecord{}, false
	}

	return usage, true
}

// GetUsageRecordsByOrder retrieves all usage records for an order
func (k Keeper) GetUsageRecordsByOrder(ctx sdk.Context, orderID string) []types.UsageRecord {
	store := ctx.KVStore(k.skey)
	key := types.UsageByOrderKey(orderID)

	bz := store.Get(key)
	if bz == nil {
		return []types.UsageRecord{}
	}

	var usageIDs []string
	if err := json.Unmarshal(bz, &usageIDs); err != nil {
		return []types.UsageRecord{}
	}

	usages := make([]types.UsageRecord, 0, len(usageIDs))
	for _, id := range usageIDs {
		if usage, found := k.GetUsageRecord(ctx, id); found {
			usages = append(usages, usage)
		}
	}

	return usages
}

// GetUnsettledUsageRecords retrieves unsettled usage records for an order
func (k Keeper) GetUnsettledUsageRecords(ctx sdk.Context, orderID string) []types.UsageRecord {
	allUsage := k.GetUsageRecordsByOrder(ctx, orderID)
	unsettled := make([]types.UsageRecord, 0)

	for _, usage := range allUsage {
		if !usage.Settled {
			unsettled = append(unsettled, usage)
		}
	}

	return unsettled
}

// WithUsageRecords iterates over all usage records
func (k Keeper) WithUsageRecords(ctx sdk.Context, fn func(types.UsageRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixUsageRecord)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var usage types.UsageRecord
		if err := json.Unmarshal(iter.Value(), &usage); err != nil {
			continue
		}
		if fn(usage) {
			break
		}
	}
}

// ============================================================================
// Reward Storage
// ============================================================================

// SetRewardDistribution saves a reward distribution to the store
func (k Keeper) SetRewardDistribution(ctx sdk.Context, dist types.RewardDistribution) error {
	if err := dist.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(&dist)
	if err != nil {
		return err
	}

	// Store by distribution ID
	store.Set(types.RewardDistributionKey(dist.DistributionID), bz)

	// Append to epoch's distribution list
	k.appendDistributionToEpoch(ctx, dist.EpochNumber, dist.DistributionID)

	// Update claimable rewards for each recipient
	for _, recipient := range dist.Recipients {
		addr, err := sdk.AccAddressFromBech32(recipient.Address)
		if err != nil {
			continue
		}

		entry := types.RewardEntry{
			DistributionID: dist.DistributionID,
			Source:         dist.Source,
			Amount:         recipient.Amount,
			CreatedAt:      dist.DistributedAt,
			Reason:         recipient.Reason,
		}

		if err := k.AddClaimableReward(ctx, addr, entry); err != nil {
			k.Logger(ctx).Error("failed to add claimable reward", "error", err, "address", recipient.Address)
		}
	}

	return nil
}

// appendDistributionToEpoch appends a distribution ID to an epoch's distribution list
func (k Keeper) appendDistributionToEpoch(ctx sdk.Context, epoch uint64, distributionID string) {
	store := ctx.KVStore(k.skey)
	key := types.RewardByEpochKey(epoch)

	distributionIDs := make([]string, 0, 1)
	bz := store.Get(key)
	if bz != nil {
		_ = json.Unmarshal(bz, &distributionIDs)
	}

	distributionIDs = append(distributionIDs, distributionID)
	bz, _ = json.Marshal(distributionIDs) //nolint:errchkjson // string slice cannot fail to marshal
	store.Set(key, bz)
}

// GetRewardDistribution retrieves a reward distribution by ID
func (k Keeper) GetRewardDistribution(ctx sdk.Context, distributionID string) (types.RewardDistribution, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.RewardDistributionKey(distributionID))
	if bz == nil {
		return types.RewardDistribution{}, false
	}

	var dist types.RewardDistribution
	if err := json.Unmarshal(bz, &dist); err != nil {
		return types.RewardDistribution{}, false
	}

	return dist, true
}

// GetRewardsByEpoch retrieves all reward distributions for an epoch
func (k Keeper) GetRewardsByEpoch(ctx sdk.Context, epoch uint64) []types.RewardDistribution {
	store := ctx.KVStore(k.skey)
	key := types.RewardByEpochKey(epoch)

	bz := store.Get(key)
	if bz == nil {
		return []types.RewardDistribution{}
	}

	var distributionIDs []string
	if err := json.Unmarshal(bz, &distributionIDs); err != nil {
		return []types.RewardDistribution{}
	}

	distributions := make([]types.RewardDistribution, 0, len(distributionIDs))
	for _, id := range distributionIDs {
		if dist, found := k.GetRewardDistribution(ctx, id); found {
			distributions = append(distributions, dist)
		}
	}

	return distributions
}

// ============================================================================
// Claimable Rewards
// ============================================================================

// SetClaimableRewards saves claimable rewards for an address
func (k Keeper) SetClaimableRewards(ctx sdk.Context, addr sdk.AccAddress, rewards types.ClaimableRewards) error {
	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(&rewards)
	if err != nil {
		return err
	}

	store.Set(types.ClaimableRewardsKey(addr), bz)
	return nil
}

// GetClaimableRewards retrieves claimable rewards for an address
func (k Keeper) GetClaimableRewards(ctx sdk.Context, address sdk.AccAddress) (types.ClaimableRewards, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ClaimableRewardsKey(address))
	if bz == nil {
		return types.ClaimableRewards{}, false
	}

	var rewards types.ClaimableRewards
	if err := json.Unmarshal(bz, &rewards); err != nil {
		return types.ClaimableRewards{}, false
	}

	return rewards, true
}

// AddClaimableReward adds a reward entry to an address's claimable rewards
func (k Keeper) AddClaimableReward(ctx sdk.Context, address sdk.AccAddress, entry types.RewardEntry) error {
	rewards, found := k.GetClaimableRewards(ctx, address)
	if !found {
		rewards = *types.NewClaimableRewards(address.String(), ctx.BlockTime())
	}

	rewards.AddReward(entry)
	rewards.LastUpdated = ctx.BlockTime()

	return k.SetClaimableRewards(ctx, address, rewards)
}

// generateID generates a unique ID with a prefix
func generateID(prefix string, seq uint64) string {
	return fmt.Sprintf("%s-%d", prefix, seq)
}

// generateIDWithTimestamp generates a unique ID with a prefix and timestamp
func generateIDWithTimestamp(prefix string, seq uint64, timestamp int64) string {
	return fmt.Sprintf("%s-%d-%d", prefix, timestamp, seq)
}

// calculateCurrentEpoch calculates the current epoch number based on block height
func (k Keeper) calculateCurrentEpoch(ctx sdk.Context) uint64 {
	params := k.GetParams(ctx)
	if params.StakingRewardEpochLength == 0 {
		return 1
	}
	height := ctx.BlockHeight()
	if height < 0 {
		return 1
	}
	return uint64(height) / params.StakingRewardEpochLength
}

// ============================================================================
// Genesis Sequence Setters
// ============================================================================

// SetNextEscrowSequence sets the next escrow sequence
func (k Keeper) SetNextEscrowSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.EscrowSequenceKey(), seq)
}

// SetNextSettlementSequence sets the next settlement sequence
func (k Keeper) SetNextSettlementSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.SettlementSequenceKey(), seq)
}

// SetNextUsageSequence sets the next usage sequence
func (k Keeper) SetNextUsageSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.UsageSequenceKey(), seq)
}

// SetNextDistributionSequence sets the next distribution sequence
func (k Keeper) SetNextDistributionSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.DistributionSequenceKey(), seq)
}

// ============================================================================
// Additional Iterators
// ============================================================================

// WithRewardDistributions iterates over all reward distributions
func (k Keeper) WithRewardDistributions(ctx sdk.Context, fn func(types.RewardDistribution) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixRewardDistribution)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var dist types.RewardDistribution
		if err := json.Unmarshal(iter.Value(), &dist); err != nil {
			continue
		}
		if fn(dist) {
			break
		}
	}
}

// WithClaimableRewards iterates over all claimable rewards
func (k Keeper) WithClaimableRewards(ctx sdk.Context, fn func(types.ClaimableRewards) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixClaimableRewards)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var rewards types.ClaimableRewards
		if err := json.Unmarshal(iter.Value(), &rewards); err != nil {
			continue
		}
		if fn(rewards) {
			break
		}
	}
}

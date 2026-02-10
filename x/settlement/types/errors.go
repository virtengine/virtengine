package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Error codes for the settlement module
// NOTE: Error codes start at 100 to avoid conflicts with Cosmos SDK core, IBC-Go, and CosmWasm modules
var (
	// ErrInvalidEscrow is returned when an escrow is malformed
	ErrInvalidEscrow = errorsmod.Register(ModuleName, 1500, "invalid escrow")

	// ErrEscrowNotFound is returned when an escrow is not found
	ErrEscrowNotFound = errorsmod.Register(ModuleName, 1501, "escrow not found")

	// ErrEscrowExists is returned when an escrow already exists
	ErrEscrowExists = errorsmod.Register(ModuleName, 1502, "escrow already exists")

	// ErrInvalidStateTransition is returned when a state transition is invalid
	ErrInvalidStateTransition = errorsmod.Register(ModuleName, 1503, "invalid state transition")

	// ErrInsufficientFunds is returned when there are insufficient funds
	ErrInsufficientFunds = errorsmod.Register(ModuleName, 1504, "insufficient funds")

	// ErrUnauthorized is returned when an action is unauthorized
	ErrUnauthorized = errorsmod.Register(ModuleName, 1505, "unauthorized")

	// ErrInvalidSettlement is returned when a settlement is malformed
	ErrInvalidSettlement = errorsmod.Register(ModuleName, 1506, "invalid settlement")

	// ErrSettlementNotFound is returned when a settlement is not found
	ErrSettlementNotFound = errorsmod.Register(ModuleName, 1507, "settlement not found")

	// ErrSettlementExists is returned when a settlement already exists
	ErrSettlementExists = errorsmod.Register(ModuleName, 1508, "settlement already exists")

	// ErrInvalidReward is returned when a reward is malformed
	ErrInvalidReward = errorsmod.Register(ModuleName, 1509, "invalid reward")

	// ErrRewardNotFound is returned when a reward distribution is not found
	ErrRewardNotFound = errorsmod.Register(ModuleName, 1510, "reward distribution not found")

	// ErrNoClaimableRewards is returned when there are no claimable rewards
	ErrNoClaimableRewards = errorsmod.Register(ModuleName, 1511, "no claimable rewards")

	// ErrInvalidUsageRecord is returned when a usage record is malformed
	ErrInvalidUsageRecord = errorsmod.Register(ModuleName, 1512, "invalid usage record")

	// ErrUsageRecordNotFound is returned when a usage record is not found
	ErrUsageRecordNotFound = errorsmod.Register(ModuleName, 1513, "usage record not found")

	// ErrUsageRecordExists is returned when a usage record already exists
	ErrUsageRecordExists = errorsmod.Register(ModuleName, 1514, "usage record already exists")

	// ErrUsageAlreadySettled is returned when usage has already been settled
	ErrUsageAlreadySettled = errorsmod.Register(ModuleName, 1515, "usage already settled")

	// ErrInvalidCondition is returned when a release condition is invalid
	ErrInvalidCondition = errorsmod.Register(ModuleName, 1516, "invalid release condition")

	// ErrConditionsNotMet is returned when release conditions are not met
	ErrConditionsNotMet = errorsmod.Register(ModuleName, 1517, "release conditions not met")

	// ErrEscrowExpired is returned when an escrow has expired
	ErrEscrowExpired = errorsmod.Register(ModuleName, 1518, "escrow has expired")

	// ErrEscrowNotActive is returned when an escrow is not active
	ErrEscrowNotActive = errorsmod.Register(ModuleName, 1519, "escrow not active")

	// ErrOrderNotFound is returned when an order is not found
	ErrOrderNotFound = errorsmod.Register(ModuleName, 1520, "order not found")

	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errorsmod.Register(ModuleName, 1521, "invalid address")

	// ErrInvalidAmount is returned when an amount is invalid
	ErrInvalidAmount = errorsmod.Register(ModuleName, 1522, "invalid amount")

	// ErrEscrowDisputed is returned when an escrow is in disputed state
	ErrEscrowDisputed = errorsmod.Register(ModuleName, 1523, "escrow is disputed")

	// ErrInvalidSignature is returned when a signature is invalid
	ErrInvalidSignature = errorsmod.Register(ModuleName, 1524, "invalid signature")

	// ErrInvalidParams is returned when module parameters are invalid
	ErrInvalidParams = errorsmod.Register(ModuleName, 1525, "invalid params")

	// ErrRewardClaimFailed is returned when reward claim fails
	ErrRewardClaimFailed = errorsmod.Register(ModuleName, 1526, "reward claim failed")

	// ErrInvalidEpoch is returned when an epoch number is invalid
	ErrInvalidEpoch = errorsmod.Register(ModuleName, 1527, "invalid epoch")

	// ErrDistributionFailed is returned when distribution fails
	ErrDistributionFailed = errorsmod.Register(ModuleName, 1528, "distribution failed")

	// ErrLeaseNotFound is returned when a lease is not found
	ErrLeaseNotFound = errorsmod.Register(ModuleName, 1529, "lease not found")

	// ErrInvalidPayout is returned when a payout is malformed
	ErrInvalidPayout = errorsmod.Register(ModuleName, 1530, "invalid payout")

	// ErrPayoutNotFound is returned when a payout is not found
	ErrPayoutNotFound = errorsmod.Register(ModuleName, 1531, "payout not found")

	// ErrPayoutExists is returned when a payout already exists
	ErrPayoutExists = errorsmod.Register(ModuleName, 1532, "payout already exists")

	// ErrPayoutIdempotent is returned when a payout has already been processed
	ErrPayoutIdempotent = errorsmod.Register(ModuleName, 1533, "payout already processed (idempotent)")

	// ErrPayoutHeld is returned when a payout is on hold
	ErrPayoutHeld = errorsmod.Register(ModuleName, 1534, "payout is on hold")

	// ErrDisputeActive is returned when there's an active dispute
	ErrDisputeActive = errorsmod.Register(ModuleName, 1535, "active dispute prevents payout")

	// ErrInvoiceNotPaid is returned when invoice is not paid
	ErrInvoiceNotPaid = errorsmod.Register(ModuleName, 1536, "invoice not paid")

	// ErrPayoutExecutionFailed is returned when payout execution fails
	ErrPayoutExecutionFailed = errorsmod.Register(ModuleName, 1537, "payout execution failed")

	// ErrFiatConversionNotFound is returned when conversion record is missing
	ErrFiatConversionNotFound = errorsmod.Register(ModuleName, 1538, "fiat conversion not found")

	// ErrFiatConversionNotAllowed is returned when conversion is not permitted
	ErrFiatConversionNotAllowed = errorsmod.Register(ModuleName, 1539, "fiat conversion not allowed")

	// ErrFiatConversionFailed is returned when conversion fails
	ErrFiatConversionFailed = errorsmod.Register(ModuleName, 1540, "fiat conversion failed")

	// ErrComplianceRequired is returned when compliance requirements are not met
	ErrComplianceRequired = errorsmod.Register(ModuleName, 1541, "compliance requirements not met")

	// ErrDexUnavailable is returned when DEX integration is missing
	ErrDexUnavailable = errorsmod.Register(ModuleName, 1542, "dex integration unavailable")

	// ErrOffRampUnavailable is returned when off-ramp integration is missing
	ErrOffRampUnavailable = errorsmod.Register(ModuleName, 1543, "off-ramp integration unavailable")

	// ErrFiatLimitExceeded is returned when conversion limits are exceeded
	ErrFiatLimitExceeded = errorsmod.Register(ModuleName, 1544, "fiat conversion limit exceeded")

	// ErrOracleUnavailable is returned when oracle sources are unavailable
	ErrOracleUnavailable = errorsmod.Register(ModuleName, 1545, "oracle sources unavailable")

	// ErrOracleStalePrice is returned when oracle prices are stale
	ErrOracleStalePrice = errorsmod.Register(ModuleName, 1546, "oracle price stale")

	// ErrOracleInsufficientSources is returned when not enough oracle sources are available
	ErrOracleInsufficientSources = errorsmod.Register(ModuleName, 1547, "insufficient oracle sources")

	// ErrRateUnavailable is returned when settlement rates cannot be locked
	ErrRateUnavailable = errorsmod.Register(ModuleName, 1548, "settlement rate unavailable")
)

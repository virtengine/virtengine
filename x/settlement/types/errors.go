package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Error codes for the settlement module
var (
	// ErrInvalidEscrow is returned when an escrow is malformed
	ErrInvalidEscrow = errorsmod.Register(ModuleName, 1, "invalid escrow")

	// ErrEscrowNotFound is returned when an escrow is not found
	ErrEscrowNotFound = errorsmod.Register(ModuleName, 2, "escrow not found")

	// ErrEscrowExists is returned when an escrow already exists
	ErrEscrowExists = errorsmod.Register(ModuleName, 3, "escrow already exists")

	// ErrInvalidStateTransition is returned when a state transition is invalid
	ErrInvalidStateTransition = errorsmod.Register(ModuleName, 4, "invalid state transition")

	// ErrInsufficientFunds is returned when there are insufficient funds
	ErrInsufficientFunds = errorsmod.Register(ModuleName, 5, "insufficient funds")

	// ErrUnauthorized is returned when an action is unauthorized
	ErrUnauthorized = errorsmod.Register(ModuleName, 6, "unauthorized")

	// ErrInvalidSettlement is returned when a settlement is malformed
	ErrInvalidSettlement = errorsmod.Register(ModuleName, 7, "invalid settlement")

	// ErrSettlementNotFound is returned when a settlement is not found
	ErrSettlementNotFound = errorsmod.Register(ModuleName, 8, "settlement not found")

	// ErrSettlementExists is returned when a settlement already exists
	ErrSettlementExists = errorsmod.Register(ModuleName, 9, "settlement already exists")

	// ErrInvalidReward is returned when a reward is malformed
	ErrInvalidReward = errorsmod.Register(ModuleName, 10, "invalid reward")

	// ErrRewardNotFound is returned when a reward distribution is not found
	ErrRewardNotFound = errorsmod.Register(ModuleName, 11, "reward distribution not found")

	// ErrNoClaimableRewards is returned when there are no claimable rewards
	ErrNoClaimableRewards = errorsmod.Register(ModuleName, 12, "no claimable rewards")

	// ErrInvalidUsageRecord is returned when a usage record is malformed
	ErrInvalidUsageRecord = errorsmod.Register(ModuleName, 13, "invalid usage record")

	// ErrUsageRecordNotFound is returned when a usage record is not found
	ErrUsageRecordNotFound = errorsmod.Register(ModuleName, 14, "usage record not found")

	// ErrUsageRecordExists is returned when a usage record already exists
	ErrUsageRecordExists = errorsmod.Register(ModuleName, 15, "usage record already exists")

	// ErrUsageAlreadySettled is returned when usage has already been settled
	ErrUsageAlreadySettled = errorsmod.Register(ModuleName, 16, "usage already settled")

	// ErrInvalidCondition is returned when a release condition is invalid
	ErrInvalidCondition = errorsmod.Register(ModuleName, 17, "invalid release condition")

	// ErrConditionsNotMet is returned when release conditions are not met
	ErrConditionsNotMet = errorsmod.Register(ModuleName, 18, "release conditions not met")

	// ErrEscrowExpired is returned when an escrow has expired
	ErrEscrowExpired = errorsmod.Register(ModuleName, 19, "escrow has expired")

	// ErrEscrowNotActive is returned when an escrow is not active
	ErrEscrowNotActive = errorsmod.Register(ModuleName, 20, "escrow not active")

	// ErrOrderNotFound is returned when an order is not found
	ErrOrderNotFound = errorsmod.Register(ModuleName, 21, "order not found")

	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errorsmod.Register(ModuleName, 22, "invalid address")

	// ErrInvalidAmount is returned when an amount is invalid
	ErrInvalidAmount = errorsmod.Register(ModuleName, 23, "invalid amount")

	// ErrEscrowDisputed is returned when an escrow is in disputed state
	ErrEscrowDisputed = errorsmod.Register(ModuleName, 24, "escrow is disputed")

	// ErrInvalidSignature is returned when a signature is invalid
	ErrInvalidSignature = errorsmod.Register(ModuleName, 25, "invalid signature")

	// ErrInvalidParams is returned when module parameters are invalid
	ErrInvalidParams = errorsmod.Register(ModuleName, 26, "invalid params")

	// ErrRewardClaimFailed is returned when reward claim fails
	ErrRewardClaimFailed = errorsmod.Register(ModuleName, 27, "reward claim failed")

	// ErrInvalidEpoch is returned when an epoch number is invalid
	ErrInvalidEpoch = errorsmod.Register(ModuleName, 28, "invalid epoch")

	// ErrDistributionFailed is returned when distribution fails
	ErrDistributionFailed = errorsmod.Register(ModuleName, 29, "distribution failed")

	// ErrLeaseNotFound is returned when a lease is not found
	ErrLeaseNotFound = errorsmod.Register(ModuleName, 30, "lease not found")
)

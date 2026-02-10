// Package types contains types for the delegation module.
//
// VE-922: Delegation module events
package types

// Delegation module event types
const (
	// EventTypeDelegate is emitted when tokens are delegated
	EventTypeDelegate = "delegate"

	// EventTypeUndelegate is emitted when tokens are undelegated
	EventTypeUndelegate = "undelegate"

	// EventTypeRedelegate is emitted when tokens are redelegated
	EventTypeRedelegate = "redelegate"

	// EventTypeCompleteUnbonding is emitted when unbonding completes
	EventTypeCompleteUnbonding = "complete_unbonding"

	// EventTypeCompleteRedelegation is emitted when redelegation completes
	EventTypeCompleteRedelegation = "complete_redelegation"

	// EventTypeClaimReward is emitted when rewards are claimed
	EventTypeClaimReward = "claim_reward"

	// EventTypeDistributeReward is emitted when rewards are distributed
	EventTypeDistributeReward = "distribute_reward"

	// EventTypeDelegatorSlashed is emitted when a delegator is slashed
	EventTypeDelegatorSlashed = "delegator_slashed"
)

// Delegation module event attribute keys
const (
	// AttributeKeyDelegator is the attribute key for delegator address
	AttributeKeyDelegator = "delegator"

	// AttributeKeyValidator is the attribute key for validator address
	AttributeKeyValidator = "validator"

	// AttributeKeySrcValidator is the attribute key for source validator address
	AttributeKeySrcValidator = "src_validator"

	// AttributeKeyDstValidator is the attribute key for destination validator address
	AttributeKeyDstValidator = "dst_validator"

	// AttributeKeyAmount is the attribute key for amount
	AttributeKeyAmount = "amount"

	// AttributeKeyShares is the attribute key for shares
	AttributeKeyShares = "shares"

	// AttributeKeyCompletionTime is the attribute key for completion time
	AttributeKeyCompletionTime = "completion_time"

	// AttributeKeyReward is the attribute key for reward amount
	AttributeKeyReward = "reward"

	// AttributeKeyEpoch is the attribute key for epoch number
	AttributeKeyEpoch = "epoch"

	// AttributeKeyUnbondingID is the attribute key for unbonding ID
	AttributeKeyUnbondingID = "unbonding_id"

	// AttributeKeyRedelegationID is the attribute key for redelegation ID
	AttributeKeyRedelegationID = "redelegation_id"

	// AttributeKeySlashAmount is the attribute key for slashed amount
	AttributeKeySlashAmount = "slash_amount"

	// AttributeKeySlashFraction is the attribute key for slashing fraction
	AttributeKeySlashFraction = "slash_fraction"

	// AttributeKeySlashShares is the attribute key for slashed shares
	AttributeKeySlashShares = "slash_shares"

	// AttributeKeyInfractionHeight is the attribute key for infraction height
	AttributeKeyInfractionHeight = "infraction_height"
)

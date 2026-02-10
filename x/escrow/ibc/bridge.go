// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
)

// CrossChainDeposit represents a deposit initiated from another chain via IBC.
type CrossChainDeposit struct {
	// SourceChain is the originating chain ID
	SourceChain string `json:"source_chain"`

	// SourceChannel is the IBC channel from which the deposit originated
	SourceChannel string `json:"source_channel"`

	// OriginalDenom is the denomination on the source chain
	OriginalDenom string `json:"original_denom"`

	// IBCDenom is the IBC-wrapped denomination on the destination chain
	IBCDenom string `json:"ibc_denom"`

	// Amount is the deposit amount
	Amount sdkmath.Int `json:"amount"`

	// Sender is the sender address on the source chain
	Sender string `json:"sender"`

	// DepositorOnDest is the depositor address on the destination chain
	DepositorOnDest string `json:"depositor_on_dest"`

	// TimeoutHeight is the IBC timeout height
	TimeoutHeight clienttypes.Height `json:"timeout_height"`

	// TimeoutTimestamp is the IBC timeout timestamp (Unix nanoseconds)
	TimeoutTimestamp uint64 `json:"timeout_timestamp"`
}

// Validate performs basic validation on a cross-chain deposit.
func (d CrossChainDeposit) Validate() error {
	if d.SourceChain == "" {
		return fmt.Errorf("source chain cannot be empty")
	}
	if d.SourceChannel == "" {
		return fmt.Errorf("source channel cannot be empty")
	}
	if d.OriginalDenom == "" {
		return fmt.Errorf("original denom cannot be empty")
	}
	if d.IBCDenom == "" {
		return fmt.Errorf("IBC denom cannot be empty")
	}
	if d.Amount.IsNil() || !d.Amount.IsPositive() {
		return fmt.Errorf("amount must be positive")
	}
	if d.Sender == "" {
		return fmt.Errorf("sender cannot be empty")
	}
	if d.DepositorOnDest == "" {
		return fmt.Errorf("depositor on destination cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(d.DepositorOnDest); err != nil {
		return fmt.Errorf("invalid depositor address: %w", err)
	}
	return nil
}

// CrossChainSettlement represents an atomic cross-chain settlement operation.
type CrossChainSettlement struct {
	// SettlementID is the unique identifier for this settlement
	SettlementID string `json:"settlement_id"`

	// SourceChain is the chain where the original order exists
	SourceChain string `json:"source_chain"`

	// DestChain is the chain where the provider is being paid
	DestChain string `json:"dest_chain"`

	// OrderID is the original order identifier
	OrderID string `json:"order_id"`

	// ProviderAddress is the provider's address on the destination chain
	ProviderAddress string `json:"provider_address"`

	// CustomerAddress is the customer's address on the source chain
	CustomerAddress string `json:"customer_address"`

	// TotalAmount is the total settlement amount
	TotalAmount sdk.Coins `json:"total_amount"`

	// ProviderShare is the provider's share
	ProviderShare sdk.Coins `json:"provider_share"`

	// PlatformFee is the platform fee
	PlatformFee sdk.Coins `json:"platform_fee"`

	// Channel is the IBC channel used
	Channel string `json:"channel"`

	// Status is the settlement status
	Status CrossChainSettlementStatus `json:"status"`

	// CreatedAt is the time the settlement was created
	CreatedAt time.Time `json:"created_at"`

	// CompletedAt is the time the settlement was completed (if applicable)
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// CrossChainSettlementStatus represents the status of a cross-chain settlement.
type CrossChainSettlementStatus string

const (
	// SettlementStatusPending indicates the settlement is pending IBC confirmation
	SettlementStatusPending CrossChainSettlementStatus = "pending"

	// SettlementStatusConfirmed indicates the settlement was confirmed via IBC ack
	SettlementStatusConfirmed CrossChainSettlementStatus = "confirmed"

	// SettlementStatusFailed indicates the settlement failed
	SettlementStatusFailed CrossChainSettlementStatus = "failed"

	// SettlementStatusTimedOut indicates the settlement timed out
	SettlementStatusTimedOut CrossChainSettlementStatus = "timed_out"

	// SettlementStatusRefunded indicates the settlement was refunded
	SettlementStatusRefunded CrossChainSettlementStatus = "refunded"
)

// IsValidStatus checks if the settlement status is valid.
func IsValidStatus(s CrossChainSettlementStatus) bool {
	switch s {
	case SettlementStatusPending, SettlementStatusConfirmed,
		SettlementStatusFailed, SettlementStatusTimedOut, SettlementStatusRefunded:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the status is a terminal state.
func (s CrossChainSettlementStatus) IsTerminal() bool {
	switch s {
	case SettlementStatusConfirmed, SettlementStatusFailed, SettlementStatusRefunded:
		return true
	default:
		return false
	}
}

// Validate validates the cross-chain settlement.
func (s CrossChainSettlement) Validate() error {
	if s.SettlementID == "" {
		return fmt.Errorf("settlement ID cannot be empty")
	}
	if s.SourceChain == "" {
		return fmt.Errorf("source chain cannot be empty")
	}
	if s.DestChain == "" {
		return fmt.Errorf("dest chain cannot be empty")
	}
	if s.OrderID == "" {
		return fmt.Errorf("order ID cannot be empty")
	}
	if s.ProviderAddress == "" {
		return fmt.Errorf("provider address cannot be empty")
	}
	if s.CustomerAddress == "" {
		return fmt.Errorf("customer address cannot be empty")
	}
	if !IsValidStatus(s.Status) {
		return fmt.Errorf("invalid status: %s", s.Status)
	}
	return nil
}

// SupportedIBCDenom returns whether a denomination is supported for cross-chain transfers.
func SupportedIBCDenom(denom string) bool {
	supportedDenoms := map[string]bool{
		"uve":   true, // VirtEngine native token
		"uatom": true, // Cosmos Hub
		"uosmo": true, // Osmosis
	}
	return supportedDenoms[denom]
}

// IBCDenomTrace returns the IBC denomination trace for a given source channel and denom.
func IBCDenomTrace(sourceChannel, denom string) string {
	return fmt.Sprintf("ibc/%s/%s", sourceChannel, denom)
}

// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
)

// DefaultTimeoutHeight is the default timeout height for IBC transfers (100 blocks on dest chain).
var DefaultTimeoutHeight = clienttypes.NewHeight(0, 100)

// DefaultTimeoutTimestamp is the default timeout timestamp (10 minutes from current time).
const DefaultTimeoutTimestamp = uint64(10 * 60 * 1_000_000_000) // 10 minutes in nanoseconds

// TransferParams holds the parameters for an IBC token transfer.
type TransferParams struct {
	// SourcePort is the source port (usually "transfer")
	SourcePort string

	// SourceChannel is the source channel identifier
	SourceChannel string

	// Token is the token to transfer
	Token sdk.Coin

	// Sender is the sender address
	Sender string

	// Receiver is the receiver address on the destination chain
	Receiver string

	// TimeoutHeight is the timeout block height on the destination chain
	TimeoutHeight clienttypes.Height

	// TimeoutTimestamp is the timeout timestamp in nanoseconds
	TimeoutTimestamp uint64

	// Memo is an optional memo for the transfer
	Memo string
}

// Validate validates the transfer parameters.
func (p TransferParams) Validate() error {
	if p.SourcePort == "" {
		return fmt.Errorf("source port cannot be empty")
	}
	if p.SourceChannel == "" {
		return fmt.Errorf("source channel cannot be empty")
	}
	if !p.Token.IsValid() || p.Token.IsZero() {
		return fmt.Errorf("token must be valid and non-zero")
	}
	if p.Sender == "" {
		return fmt.Errorf("sender cannot be empty")
	}
	if p.Receiver == "" {
		return fmt.Errorf("receiver cannot be empty")
	}
	if p.TimeoutHeight.IsZero() && p.TimeoutTimestamp == 0 {
		return fmt.Errorf("either timeout height or timeout timestamp must be set")
	}
	return nil
}

// NewTransferMsg creates a new IBC MsgTransfer from the given parameters.
func NewTransferMsg(params TransferParams) (*ibctransfertypes.MsgTransfer, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid transfer params: %w", err)
	}

	msg := ibctransfertypes.NewMsgTransfer(
		params.SourcePort,
		params.SourceChannel,
		params.Token,
		params.Sender,
		params.Receiver,
		params.TimeoutHeight,
		params.TimeoutTimestamp,
		params.Memo,
	)

	return msg, nil
}

// MultiHopRoute represents a multi-hop transfer route for tokens
// that need to traverse multiple IBC channels.
type MultiHopRoute struct {
	// Hops is the ordered list of channel hops
	Hops []RouteHop `json:"hops"`
}

// RouteHop represents a single hop in a multi-hop transfer route.
type RouteHop struct {
	// SourcePort is the port on the current chain
	SourcePort string `json:"source_port"`

	// SourceChannel is the channel on the current chain
	SourceChannel string `json:"source_channel"`

	// DestChainID is the destination chain ID for this hop
	DestChainID string `json:"dest_chain_id"`
}

// Validate validates the multi-hop route.
func (r MultiHopRoute) Validate() error {
	if len(r.Hops) == 0 {
		return fmt.Errorf("route must have at least one hop")
	}
	for i, hop := range r.Hops {
		if hop.SourcePort == "" {
			return fmt.Errorf("hop %d: source port cannot be empty", i)
		}
		if hop.SourceChannel == "" {
			return fmt.Errorf("hop %d: source channel cannot be empty", i)
		}
		if hop.DestChainID == "" {
			return fmt.Errorf("hop %d: dest chain ID cannot be empty", i)
		}
	}
	return nil
}

// ComputeIBCDenom computes the IBC denomination for a token after traversing a route.
func ComputeIBCDenom(baseDenom string, route MultiHopRoute) string {
	denom := baseDenom
	for _, hop := range route.Hops {
		denom = fmt.Sprintf("%s/%s/%s", ibctransfertypes.ModuleName, hop.SourceChannel, denom)
	}
	return denom
}

// EstimateTransferFee estimates the fee for an IBC transfer.
// This is a simplified estimate; actual fees depend on gas prices on each chain.
func EstimateTransferFee(amount sdkmath.Int, numHops int) sdk.Coin {
	// Base fee is 0.1% per hop, minimum 1 unit
	feePerHop := amount.Quo(sdkmath.NewInt(1000))
	if feePerHop.IsZero() {
		feePerHop = sdkmath.OneInt()
	}
	totalFee := feePerHop.Mul(sdkmath.NewInt(int64(numHops)))
	return sdk.NewCoin("uve", totalFee)
}

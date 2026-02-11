// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

const (
	// Version defines the current IBC settlement module version.
	Version = "settlement-1"

	// PortID is the default port id for the settlement IBC module.
	PortID = "settlement"

	// ModuleName is the IBC module name for settlement.
	ModuleName = "settlement-ibc"
)

const (
	// DefaultTimeoutHeightDelta is the default height delta for outgoing packets.
	DefaultTimeoutHeightDelta uint64 = 100

	// DefaultTimeoutTimestampDelta is the default timeout timestamp delta (10 minutes).
	DefaultTimeoutTimestampDelta uint64 = uint64(10 * time.Minute)
)

// PacketType identifies settlement IBC packet types.
type PacketType string

const (
	PacketTypeEscrowDeposit    PacketType = "escrow_deposit"
	PacketTypeEscrowRelease    PacketType = "escrow_release"
	PacketTypeSettlementRecord PacketType = "settlement_record"
)

// SettlementPacketData wraps packet type and data payload.
type SettlementPacketData struct {
	Type PacketType      `json:"type"`
	Data json.RawMessage `json:"data"`
}

// NewPacketData creates a new SettlementPacketData.
func NewPacketData(packetType PacketType, data interface{}) (SettlementPacketData, error) {
	bz, err := json.Marshal(data)
	if err != nil {
		return SettlementPacketData{}, err
	}
	return SettlementPacketData{
		Type: packetType,
		Data: bz,
	}, nil
}

// GetBytes returns the JSON marshaled bytes of the packet data.
func (p SettlementPacketData) GetBytes() []byte {
	bz, _ := json.Marshal(p) //nolint:errchkjson // simple struct cannot fail
	return bz
}

// Validate validates the packet data wrapper.
func (p SettlementPacketData) Validate() error {
	switch p.Type {
	case PacketTypeEscrowDeposit, PacketTypeEscrowRelease, PacketTypeSettlementRecord:
		// valid
	default:
		return fmt.Errorf("unknown packet type: %s", p.Type)
	}

	if len(p.Data) == 0 {
		return fmt.Errorf("packet data cannot be empty")
	}

	return nil
}

// EscrowDepositPacket represents a cross-chain escrow deposit request.
type EscrowDepositPacket struct {
	DepositID       string                   `json:"deposit_id"`
	OrderID         string                   `json:"order_id"`
	Depositor       string                   `json:"depositor"`
	Amount          sdk.Coins                `json:"amount"`
	ExpiresInSeconds uint64                  `json:"expires_in_seconds"`
	Conditions      []types.ReleaseCondition `json:"conditions,omitempty"`
	SourceChainID   string                   `json:"source_chain_id"`
	SourceChannel   string                   `json:"source_channel"`
	RequestedAt     time.Time                `json:"requested_at"`
	TimeoutHeight   clienttypes.Height       `json:"timeout_height"`
	TimeoutTimestamp uint64                  `json:"timeout_timestamp"`
}

// Validate validates the escrow deposit packet.
func (p EscrowDepositPacket) Validate() error {
	if p.OrderID == "" {
		return fmt.Errorf("order_id cannot be empty")
	}
	if p.Depositor == "" {
		return fmt.Errorf("depositor cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(p.Depositor); err != nil {
		return fmt.Errorf("invalid depositor address: %w", err)
	}
	if !p.Amount.IsValid() || p.Amount.IsZero() {
		return fmt.Errorf("amount must be valid and non-zero")
	}
	if p.ExpiresInSeconds == 0 {
		return fmt.Errorf("expires_in_seconds must be greater than 0")
	}
	if p.SourceChainID == "" {
		return fmt.Errorf("source_chain_id cannot be empty")
	}
	if p.SourceChannel == "" {
		return fmt.Errorf("source_channel cannot be empty")
	}
	for i, cond := range p.Conditions {
		if err := cond.Validate(); err != nil {
			return fmt.Errorf("invalid condition %d: %w", i, err)
		}
	}
	return nil
}

// EscrowReleaseType defines the release action for an escrow.
type EscrowReleaseType string

const (
	ReleaseTypeRelease EscrowReleaseType = "release"
	ReleaseTypeRefund  EscrowReleaseType = "refund"
)

// EscrowReleasePacket represents a cross-chain escrow release/refund request.
type EscrowReleasePacket struct {
	EscrowID      string            `json:"escrow_id"`
	OrderID       string            `json:"order_id"`
	ReleaseType   EscrowReleaseType `json:"release_type"`
	Reason        string            `json:"reason"`
	SourceChainID string            `json:"source_chain_id"`
	SourceChannel string            `json:"source_channel"`
	RequestedAt   time.Time         `json:"requested_at"`
}

// Validate validates the escrow release packet.
func (p EscrowReleasePacket) Validate() error {
	if p.EscrowID == "" {
		return fmt.Errorf("escrow_id cannot be empty")
	}
	if p.OrderID == "" {
		return fmt.Errorf("order_id cannot be empty")
	}
	if p.ReleaseType == "" {
		return fmt.Errorf("release_type cannot be empty")
	}
	switch p.ReleaseType {
	case ReleaseTypeRelease, ReleaseTypeRefund:
		// ok
	default:
		return fmt.Errorf("unknown release_type: %s", p.ReleaseType)
	}
	if p.SourceChainID == "" {
		return fmt.Errorf("source_chain_id cannot be empty")
	}
	if p.SourceChannel == "" {
		return fmt.Errorf("source_channel cannot be empty")
	}
	return nil
}

// SettlementRecordPacket represents a cross-chain settlement record.
type SettlementRecordPacket struct {
	Record        types.SettlementRecord `json:"record"`
	SourceChainID string                 `json:"source_chain_id"`
	SourceChannel string                 `json:"source_channel"`
}

// Validate validates the settlement record packet.
func (p SettlementRecordPacket) Validate() error {
	if p.SourceChainID == "" {
		return fmt.Errorf("source_chain_id cannot be empty")
	}
	if p.SourceChannel == "" {
		return fmt.Errorf("source_channel cannot be empty")
	}
	if err := p.Record.Validate(); err != nil {
		return fmt.Errorf("invalid settlement record: %w", err)
	}
	return nil
}

// EscrowDepositAck is returned after processing a deposit packet.
type EscrowDepositAck struct {
	EscrowID string `json:"escrow_id"`
	OrderID  string `json:"order_id"`
	Status   string `json:"status"`
}

// EscrowReleaseAck is returned after processing a release packet.
type EscrowReleaseAck struct {
	EscrowID string `json:"escrow_id"`
	Status   string `json:"status"`
}

// SettlementRecordAck is returned after processing a settlement record.
type SettlementRecordAck struct {
	SettlementID string `json:"settlement_id"`
	Status       string `json:"status"`
}

// Acknowledgement defines the IBC acknowledgement structure.
type Acknowledgement struct {
	Result []byte `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// NewResultAcknowledgement creates a successful acknowledgement.
func NewResultAcknowledgement(result interface{}) Acknowledgement {
	bz, _ := json.Marshal(result) //nolint:errchkjson // caller provides valid result
	return Acknowledgement{
		Result: bz,
	}
}

// NewErrorAcknowledgement creates an error acknowledgement.
func NewErrorAcknowledgement(err error) Acknowledgement {
	return Acknowledgement{
		Error: err.Error(),
	}
}

// Success returns true if acknowledgement is successful.
func (a Acknowledgement) Success() bool {
	return a.Error == ""
}

// GetBytes returns the JSON marshaled bytes of the acknowledgement.
func (a Acknowledgement) GetBytes() []byte {
	bz, _ := json.Marshal(a) //nolint:errchkjson // simple struct cannot fail
	return bz
}

// Acknowledgement implements the exported.Acknowledgement interface.
func (a Acknowledgement) Acknowledgement() []byte {
	return a.GetBytes()
}

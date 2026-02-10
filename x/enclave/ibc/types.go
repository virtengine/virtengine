// Package ibc implements IBC support for cross-chain enclave identity federation.
//
// This package provides IBC module interfaces for querying and syncing enclave
// identities and measurement allowlists across IBC-connected chains.
package ibc

import (
	"encoding/json"
	"fmt"

	"github.com/virtengine/virtengine/x/enclave/types"
)

const (
	// Version defines the current version the IBC enclave module supports
	Version = "enclave-1"

	// PortID is the default port id for the enclave IBC module
	PortID = "enclave"

	// ModuleName is the IBC module name for enclave
	ModuleName = "enclave-ibc"
)

// PacketType defines the type of IBC packet
type PacketType string

const (
	// PacketTypeQueryEnclaveIdentity is a query for enclave identity
	PacketTypeQueryEnclaveIdentity PacketType = "query_enclave_identity"

	// PacketTypeQueryMeasurementAllowlist is a query for measurement allowlist
	PacketTypeQueryMeasurementAllowlist PacketType = "query_measurement_allowlist"

	// PacketTypeSyncMeasurement is a sync request for measurements
	PacketTypeSyncMeasurement PacketType = "sync_measurement"

	// PacketTypeEnclaveIdentityResponse is a response containing enclave identity
	PacketTypeEnclaveIdentityResponse PacketType = "enclave_identity_response"

	// PacketTypeMeasurementAllowlistResponse is a response containing measurement allowlist
	PacketTypeMeasurementAllowlistResponse PacketType = "measurement_allowlist_response"

	// PacketTypeSyncMeasurementAck is an acknowledgement for measurement sync
	PacketTypeSyncMeasurementAck PacketType = "sync_measurement_ack"
)

// EnclavePacketData is the base structure for all enclave IBC packets
type EnclavePacketData struct {
	Type PacketType      `json:"type"`
	Data json.RawMessage `json:"data"`
}

// NewEnclavePacketData creates a new EnclavePacketData
func NewEnclavePacketData(packetType PacketType, data interface{}) (EnclavePacketData, error) {
	bz, err := json.Marshal(data)
	if err != nil {
		return EnclavePacketData{}, err
	}
	return EnclavePacketData{
		Type: packetType,
		Data: bz,
	}, nil
}

// GetBytes returns the JSON marshaled bytes of the packet data
func (p EnclavePacketData) GetBytes() []byte {
	bz, _ := json.Marshal(p) //nolint:errchkjson // panics not expected for simple struct
	return bz
}

// Validate validates the packet data
func (p EnclavePacketData) Validate() error {
	switch p.Type {
	case PacketTypeQueryEnclaveIdentity,
		PacketTypeQueryMeasurementAllowlist,
		PacketTypeSyncMeasurement,
		PacketTypeEnclaveIdentityResponse,
		PacketTypeMeasurementAllowlistResponse,
		PacketTypeSyncMeasurementAck:
		// Valid packet types
	default:
		return fmt.Errorf("unknown packet type: %s", p.Type)
	}

	if len(p.Data) == 0 {
		return fmt.Errorf("packet data cannot be empty")
	}

	return nil
}

// QueryEnclaveIdentityPacket is the data for querying a remote chain's enclave identity
type QueryEnclaveIdentityPacket struct {
	// ValidatorAddress is the validator address to query (optional - if empty, returns all active)
	ValidatorAddress string `json:"validator_address,omitempty"`

	// TEEType filters by TEE type (optional)
	TEEType string `json:"tee_type,omitempty"`

	// IncludeExpired includes expired identities in the response
	IncludeExpired bool `json:"include_expired,omitempty"`
}

// Validate validates the query packet
func (p QueryEnclaveIdentityPacket) Validate() error {
	// All fields are optional, so this is always valid
	return nil
}

// QueryMeasurementAllowlistPacket is the data for querying a remote chain's measurement allowlist
type QueryMeasurementAllowlistPacket struct {
	// TEEType filters by TEE type (optional)
	TEEType string `json:"tee_type,omitempty"`

	// IncludeRevoked includes revoked measurements in the response
	IncludeRevoked bool `json:"include_revoked,omitempty"`
}

// Validate validates the query packet
func (p QueryMeasurementAllowlistPacket) Validate() error {
	return nil
}

// SyncMeasurementPacket is the data for syncing a measurement to a remote chain
type SyncMeasurementPacket struct {
	// Measurement is the measurement to sync
	Measurement types.MeasurementRecord `json:"measurement"`

	// SourceChainID is the chain ID of the source chain
	SourceChainID string `json:"source_chain_id"`

	// ProposalID is the governance proposal that approved this sync (if applicable)
	ProposalID uint64 `json:"proposal_id,omitempty"`
}

// Validate validates the sync packet
func (p SyncMeasurementPacket) Validate() error {
	if err := types.ValidateMeasurementRecord(&p.Measurement); err != nil {
		return fmt.Errorf("invalid measurement: %w", err)
	}
	if p.SourceChainID == "" {
		return fmt.Errorf("source chain ID cannot be empty")
	}
	return nil
}

// EnclaveIdentityResponse is the response data for enclave identity queries
type EnclaveIdentityResponse struct {
	// Identities is the list of enclave identities
	Identities []types.EnclaveIdentity `json:"identities"`

	// ChainID is the chain ID of the responding chain
	ChainID string `json:"chain_id"`

	// BlockHeight is the block height at which this data was read
	BlockHeight int64 `json:"block_height"`
}

// MeasurementAllowlistResponse is the response data for measurement allowlist queries
type MeasurementAllowlistResponse struct {
	// Measurements is the list of measurements
	Measurements []types.MeasurementRecord `json:"measurements"`

	// ChainID is the chain ID of the responding chain
	ChainID string `json:"chain_id"`

	// BlockHeight is the block height at which this data was read
	BlockHeight int64 `json:"block_height"`
}

// SyncMeasurementAck is the acknowledgement data for measurement sync
type SyncMeasurementAck struct {
	// Success indicates if the sync was successful
	Success bool `json:"success"`

	// Error is the error message if sync failed
	Error string `json:"error,omitempty"`

	// MeasurementHash is the hash of the synced measurement
	MeasurementHash string `json:"measurement_hash"`
}

// Acknowledgement defines the IBC acknowledgement structure
type Acknowledgement struct {
	// Result is the successful result data (JSON encoded)
	Result []byte `json:"result,omitempty"`

	// Error is the error message if the packet processing failed
	Error string `json:"error,omitempty"`
}

// NewResultAcknowledgement creates a successful acknowledgement
func NewResultAcknowledgement(result interface{}) Acknowledgement {
	bz, _ := json.Marshal(result) //nolint:errchkjson // intentional - caller should provide valid result
	return Acknowledgement{
		Result: bz,
	}
}

// NewErrorAcknowledgement creates an error acknowledgement
func NewErrorAcknowledgement(err error) Acknowledgement {
	return Acknowledgement{
		Error: err.Error(),
	}
}

// Success returns true if the acknowledgement is successful
func (a Acknowledgement) Success() bool {
	return a.Error == ""
}

// GetBytes returns the JSON marshaled bytes of the acknowledgement
func (a Acknowledgement) GetBytes() []byte {
	bz, _ := json.Marshal(a) //nolint:errchkjson // simple struct cannot fail to marshal
	return bz
}

// Acknowledgement implements the exported.Acknowledgement interface
func (a Acknowledgement) Acknowledgement() []byte {
	return a.GetBytes()
}

// FederatedIdentity represents an enclave identity from a remote chain
type FederatedIdentity struct {
	// Identity is the enclave identity
	Identity types.EnclaveIdentity `json:"identity"`

	// SourceChainID is the chain ID where this identity originates
	SourceChainID string `json:"source_chain_id"`

	// SourceChannelID is the IBC channel through which this identity was received
	SourceChannelID string `json:"source_channel_id"`

	// ReceivedHeight is the local block height when this identity was received
	ReceivedHeight int64 `json:"received_height"`

	// Verified indicates if this identity has been verified locally
	Verified bool `json:"verified"`
}

// FederatedMeasurement represents a measurement synced from a remote chain
type FederatedMeasurement struct {
	// Measurement is the measurement record
	Measurement types.MeasurementRecord `json:"measurement"`

	// SourceChainID is the chain ID where this measurement originates
	SourceChainID string `json:"source_chain_id"`

	// SourceChannelID is the IBC channel through which this measurement was received
	SourceChannelID string `json:"source_channel_id"`

	// ReceivedHeight is the local block height when this measurement was received
	ReceivedHeight int64 `json:"received_height"`

	// Trusted indicates if this measurement is trusted for local verification
	Trusted bool `json:"trusted"`
}

// ChannelMetadata stores metadata about an IBC channel
type ChannelMetadata struct {
	// CounterpartyChainID is the chain ID of the counterparty
	CounterpartyChainID string `json:"counterparty_chain_id"`

	// Version is the IBC version negotiated
	Version string `json:"version"`

	// Trusted indicates if this channel is trusted for measurement sync
	Trusted bool `json:"trusted"`
}

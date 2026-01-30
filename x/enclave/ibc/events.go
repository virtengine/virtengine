package ibc

// IBC event types for enclave module
const (
	// EventTypePacketSent is emitted when a packet is sent
	EventTypePacketSent = "enclave_ibc_packet_sent"

	// EventTypePacketReceived is emitted when a packet is received
	EventTypePacketReceived = "enclave_ibc_packet_received"

	// EventTypePacketAcknowledged is emitted when a packet acknowledgement is processed
	EventTypePacketAcknowledged = "enclave_ibc_packet_acknowledged"

	// EventTypePacketTimeout is emitted when a packet times out
	EventTypePacketTimeout = "enclave_ibc_packet_timeout"

	// EventTypeIdentityQueried is emitted when an identity is queried via IBC
	EventTypeIdentityQueried = "enclave_ibc_identity_queried"

	// EventTypeMeasurementQueried is emitted when measurements are queried via IBC
	EventTypeMeasurementQueried = "enclave_ibc_measurement_queried"

	// EventTypeMeasurementSynced is emitted when a measurement is synced via IBC
	EventTypeMeasurementSynced = "enclave_ibc_measurement_synced"

	// EventTypeFederatedIdentityReceived is emitted when a federated identity is received
	EventTypeFederatedIdentityReceived = "enclave_ibc_federated_identity_received"

	// EventTypeChannelOpened is emitted when an IBC channel is opened
	EventTypeChannelOpened = "enclave_ibc_channel_opened"

	// EventTypeChannelClosed is emitted when an IBC channel is closed
	EventTypeChannelClosed = "enclave_ibc_channel_closed"
)

// IBC event attribute keys
const (
	// AttributeKeyPacketType is the packet type
	AttributeKeyPacketType = "packet_type"

	// AttributeKeySourceChannel is the source channel ID
	AttributeKeySourceChannel = "source_channel"

	// AttributeKeyDestinationChannel is the destination channel ID
	AttributeKeyDestinationChannel = "destination_channel"

	// AttributeKeySourcePort is the source port ID
	AttributeKeySourcePort = "source_port"

	// AttributeKeyDestinationPort is the destination port ID
	AttributeKeyDestinationPort = "destination_port"

	// AttributeKeySequence is the packet sequence number
	AttributeKeySequence = "sequence"

	// AttributeKeyCounterpartyChainID is the counterparty chain ID
	AttributeKeyCounterpartyChainID = "counterparty_chain_id"

	// AttributeKeyValidatorAddress is the validator address
	AttributeKeyValidatorAddress = "validator_address"

	// AttributeKeyTEEType is the TEE type
	AttributeKeyTEEType = "tee_type"

	// AttributeKeyMeasurementHash is the measurement hash
	AttributeKeyMeasurementHash = "measurement_hash"

	// AttributeKeyIdentityCount is the count of identities
	AttributeKeyIdentityCount = "identity_count"

	// AttributeKeyMeasurementCount is the count of measurements
	AttributeKeyMeasurementCount = "measurement_count"

	// AttributeKeySuccess is whether the operation was successful
	AttributeKeySuccess = "success"

	// AttributeKeyError is the error message
	AttributeKeyError = "error"

	// AttributeKeyAckSuccess is whether the acknowledgement indicates success
	AttributeKeyAckSuccess = "ack_success"

	// AttributeKeyChannelID is the channel ID
	AttributeKeyChannelID = "channel_id"

	// AttributeKeyPortID is the port ID
	AttributeKeyPortID = "port_id"

	// AttributeKeyVersion is the IBC version
	AttributeKeyVersion = "version"

	// AttributeKeySourceChainID is the source chain ID
	AttributeKeySourceChainID = "source_chain_id"

	// AttributeKeyTrusted is whether the channel/measurement is trusted
	AttributeKeyTrusted = "trusted"
)

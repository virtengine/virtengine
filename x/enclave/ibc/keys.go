package ibc

import "encoding/binary"

// IBC store key prefixes
var (
	// PrefixFederatedIdentity is the prefix for federated identity storage
	// Key: PrefixFederatedIdentity | source_chain_id | validator_address -> FederatedIdentity
	PrefixFederatedIdentity = []byte{0x10}

	// PrefixFederatedMeasurement is the prefix for federated measurement storage
	// Key: PrefixFederatedMeasurement | source_chain_id | measurement_hash -> FederatedMeasurement
	PrefixFederatedMeasurement = []byte{0x11}

	// PrefixChannelMetadata is the prefix for channel metadata storage
	// Key: PrefixChannelMetadata | channel_id -> ChannelMetadata
	PrefixChannelMetadata = []byte{0x12}

	// PrefixPortBinding is the prefix for port binding state
	// Key: PrefixPortBinding -> bool (whether port is bound)
	PrefixPortBinding = []byte{0x13}

	// PrefixPendingPacket is the prefix for pending packets awaiting acknowledgement
	// Key: PrefixPendingPacket | channel_id | sequence -> packet data
	PrefixPendingPacket = []byte{0x14}

	// PrefixTrustedChain is the prefix for trusted chain IDs
	// Key: PrefixTrustedChain | chain_id -> bool
	PrefixTrustedChain = []byte{0x15}
)

// FederatedIdentityKey returns the store key for a federated identity
func FederatedIdentityKey(sourceChainID string, validatorAddr []byte) []byte {
	chainIDBytes := []byte(sourceChainID)
	key := make([]byte, 0, len(PrefixFederatedIdentity)+len(chainIDBytes)+1+len(validatorAddr))
	key = append(key, PrefixFederatedIdentity...)
	key = append(key, chainIDBytes...)
	key = append(key, 0x00) // separator
	key = append(key, validatorAddr...)
	return key
}

// FederatedMeasurementKey returns the store key for a federated measurement
func FederatedMeasurementKey(sourceChainID string, measurementHash []byte) []byte {
	chainIDBytes := []byte(sourceChainID)
	key := make([]byte, 0, len(PrefixFederatedMeasurement)+len(chainIDBytes)+1+len(measurementHash))
	key = append(key, PrefixFederatedMeasurement...)
	key = append(key, chainIDBytes...)
	key = append(key, 0x00) // separator
	key = append(key, measurementHash...)
	return key
}

// ChannelMetadataKey returns the store key for channel metadata
func ChannelMetadataKey(channelID string) []byte {
	channelIDBytes := []byte(channelID)
	key := make([]byte, 0, len(PrefixChannelMetadata)+len(channelIDBytes))
	key = append(key, PrefixChannelMetadata...)
	key = append(key, channelIDBytes...)
	return key
}

// PortBindingKey returns the store key for port binding state
func PortBindingKey() []byte {
	return PrefixPortBinding
}

// PendingPacketKey returns the store key for a pending packet
func PendingPacketKey(channelID string, sequence uint64) []byte {
	channelIDBytes := []byte(channelID)
	key := make([]byte, 0, len(PrefixPendingPacket)+len(channelIDBytes)+1+8)
	key = append(key, PrefixPendingPacket...)
	key = append(key, channelIDBytes...)
	key = append(key, 0x00) // separator
	seqBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(seqBytes, sequence)
	key = append(key, seqBytes...)
	return key
}

// TrustedChainKey returns the store key for a trusted chain
func TrustedChainKey(chainID string) []byte {
	chainIDBytes := []byte(chainID)
	key := make([]byte, 0, len(PrefixTrustedChain)+len(chainIDBytes))
	key = append(key, PrefixTrustedChain...)
	key = append(key, chainIDBytes...)
	return key
}

// FederatedIdentityPrefix returns the prefix for iterating federated identities from a chain
func FederatedIdentityPrefix(sourceChainID string) []byte {
	chainIDBytes := []byte(sourceChainID)
	key := make([]byte, 0, len(PrefixFederatedIdentity)+len(chainIDBytes)+1)
	key = append(key, PrefixFederatedIdentity...)
	key = append(key, chainIDBytes...)
	key = append(key, 0x00) // separator
	return key
}

// FederatedMeasurementPrefix returns the prefix for iterating federated measurements from a chain
func FederatedMeasurementPrefix(sourceChainID string) []byte {
	chainIDBytes := []byte(sourceChainID)
	key := make([]byte, 0, len(PrefixFederatedMeasurement)+len(chainIDBytes)+1)
	key = append(key, PrefixFederatedMeasurement...)
	key = append(key, chainIDBytes...)
	key = append(key, 0x00) // separator
	return key
}

package ibc

import (
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/enclave/types"
)

func TestEnclavePacketData_Validate(t *testing.T) {
	tests := []struct {
		name      string
		data      EnclavePacketData
		expectErr bool
	}{
		{
			name: "valid query identity packet",
			data: EnclavePacketData{
				Type: PacketTypeQueryEnclaveIdentity,
				Data: json.RawMessage(`{"validator_address":"cosmos1..."}`),
			},
			expectErr: false,
		},
		{
			name: "valid query measurement packet",
			data: EnclavePacketData{
				Type: PacketTypeQueryMeasurementAllowlist,
				Data: json.RawMessage(`{"tee_type":"SGX"}`),
			},
			expectErr: false,
		},
		{
			name: "invalid packet type",
			data: EnclavePacketData{
				Type: "invalid_type",
				Data: json.RawMessage(`{}`),
			},
			expectErr: true,
		},
		{
			name: "empty packet data",
			data: EnclavePacketData{
				Type: PacketTypeQueryEnclaveIdentity,
				Data: nil,
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.data.Validate()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewEnclavePacketData(t *testing.T) {
	query := QueryEnclaveIdentityPacket{
		ValidatorAddress: "cosmos1abc...",
		TEEType:          "SGX",
	}

	packet, err := NewEnclavePacketData(PacketTypeQueryEnclaveIdentity, query)
	require.NoError(t, err)
	require.Equal(t, PacketTypeQueryEnclaveIdentity, packet.Type)
	require.NotEmpty(t, packet.Data)

	// Verify the data can be unmarshaled back
	var decoded QueryEnclaveIdentityPacket
	err = json.Unmarshal(packet.Data, &decoded)
	require.NoError(t, err)
	require.Equal(t, query.ValidatorAddress, decoded.ValidatorAddress)
	require.Equal(t, query.TEEType, decoded.TEEType)
}

func TestQueryEnclaveIdentityPacket_Validate(t *testing.T) {
	tests := []struct {
		name      string
		packet    QueryEnclaveIdentityPacket
		expectErr bool
	}{
		{
			name:      "empty packet is valid",
			packet:    QueryEnclaveIdentityPacket{},
			expectErr: false,
		},
		{
			name: "with validator address",
			packet: QueryEnclaveIdentityPacket{
				ValidatorAddress: "cosmos1abc...",
			},
			expectErr: false,
		},
		{
			name: "with tee type filter",
			packet: QueryEnclaveIdentityPacket{
				TEEType: "SGX",
			},
			expectErr: false,
		},
		{
			name: "with include expired",
			packet: QueryEnclaveIdentityPacket{
				IncludeExpired: true,
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.packet.Validate()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSyncMeasurementPacket_Validate(t *testing.T) {
	validMeasurement := types.MeasurementRecord{
		MeasurementHash: make([]byte, 32),
		TEEType:         types.TEETypeSGX,
		Description:     "Test measurement",
		MinISVSVN:       1,
	}

	tests := []struct {
		name      string
		packet    SyncMeasurementPacket
		expectErr bool
	}{
		{
			name: "valid packet",
			packet: SyncMeasurementPacket{
				Measurement:   validMeasurement,
				SourceChainID: "testchain-1",
			},
			expectErr: false,
		},
		{
			name: "empty source chain id",
			packet: SyncMeasurementPacket{
				Measurement:   validMeasurement,
				SourceChainID: "",
			},
			expectErr: true,
		},
		{
			name: "invalid measurement",
			packet: SyncMeasurementPacket{
				Measurement: types.MeasurementRecord{
					MeasurementHash: nil,
				},
				SourceChainID: "testchain-1",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.packet.Validate()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAcknowledgement(t *testing.T) {
	t.Run("success acknowledgement", func(t *testing.T) {
		result := EnclaveIdentityResponse{
			ChainID:     "testchain-1",
			BlockHeight: 100,
		}
		ack := NewResultAcknowledgement(result)
		require.True(t, ack.Success())
		require.NotEmpty(t, ack.Result)
		require.Empty(t, ack.Error)
	})

	t.Run("error acknowledgement", func(t *testing.T) {
		ack := NewErrorAcknowledgement(ErrInvalidPacket)
		require.False(t, ack.Success())
		require.Empty(t, ack.Result)
		require.NotEmpty(t, ack.Error)
	})
}

func TestFederatedIdentityKey(t *testing.T) {
	validatorAddr := []byte("validator123")
	chainID := "cosmoshub-4"

	key := FederatedIdentityKey(chainID, validatorAddr)
	require.NotEmpty(t, key)
	require.True(t, len(key) > len(PrefixFederatedIdentity))

	// Keys for different chains should be different
	key2 := FederatedIdentityKey("osmosis-1", validatorAddr)
	require.NotEqual(t, key, key2)

	// Keys for different validators should be different
	key3 := FederatedIdentityKey(chainID, []byte("validator456"))
	require.NotEqual(t, key, key3)
}

func TestFederatedMeasurementKey(t *testing.T) {
	measurementHash := make([]byte, 32)
	copy(measurementHash, []byte("measurement"))
	chainID := "cosmoshub-4"

	key := FederatedMeasurementKey(chainID, measurementHash)
	require.NotEmpty(t, key)
	require.True(t, len(key) > len(PrefixFederatedMeasurement))
}

func TestChannelMetadataKey(t *testing.T) {
	key := ChannelMetadataKey("channel-0")
	require.NotEmpty(t, key)
	require.True(t, len(key) > len(PrefixChannelMetadata))
}

func TestPendingPacketKey(t *testing.T) {
	key := PendingPacketKey("channel-0", 42)
	require.NotEmpty(t, key)
	require.True(t, len(key) > len(PrefixPendingPacket)+8)
}

func TestTrustedChainKey(t *testing.T) {
	key := TrustedChainKey("cosmoshub-4")
	require.NotEmpty(t, key)
	require.True(t, len(key) > len(PrefixTrustedChain))
}

func TestEnclaveIdentityResponse(t *testing.T) {
	response := EnclaveIdentityResponse{
		Identities: []types.EnclaveIdentity{
			{
				ValidatorAddress: "cosmos1abc...",
				TEEType:          types.TEETypeSGX,
				MeasurementHash:  make([]byte, 32),
				EncryptionPubKey: []byte("pubkey"),
				SigningPubKey:    []byte("sigkey"),
				AttestationQuote: []byte("quote"),
				ExpiryHeight:     1000,
				Status:           types.EnclaveIdentityStatusActive,
			},
		},
		ChainID:     "testchain-1",
		BlockHeight: 500,
	}

	bz, err := json.Marshal(response)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	var decoded EnclaveIdentityResponse
	err = json.Unmarshal(bz, &decoded)
	require.NoError(t, err)
	require.Equal(t, response.ChainID, decoded.ChainID)
	require.Equal(t, response.BlockHeight, decoded.BlockHeight)
	require.Len(t, decoded.Identities, 1)
}

func TestMeasurementAllowlistResponse(t *testing.T) {
	response := MeasurementAllowlistResponse{
		Measurements: []types.MeasurementRecord{
			{
				MeasurementHash: make([]byte, 32),
				TEEType:         types.TEETypeSGX,
				Description:     "Test measurement",
				MinISVSVN:       1,
				AddedAt:         time.Now(),
			},
		},
		ChainID:     "testchain-1",
		BlockHeight: 500,
	}

	bz, err := json.Marshal(response)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	var decoded MeasurementAllowlistResponse
	err = json.Unmarshal(bz, &decoded)
	require.NoError(t, err)
	require.Equal(t, response.ChainID, decoded.ChainID)
	require.Len(t, decoded.Measurements, 1)
}

func TestSyncMeasurementAck(t *testing.T) {
	ack := SyncMeasurementAck{
		Success:         true,
		MeasurementHash: hex.EncodeToString(make([]byte, 32)),
	}

	bz, err := json.Marshal(ack)
	require.NoError(t, err)

	var decoded SyncMeasurementAck
	err = json.Unmarshal(bz, &decoded)
	require.NoError(t, err)
	require.True(t, decoded.Success)
	require.Equal(t, ack.MeasurementHash, decoded.MeasurementHash)
}

func TestChannelMetadata(t *testing.T) {
	metadata := ChannelMetadata{
		CounterpartyChainID: "osmosis-1",
		Version:             Version,
		Trusted:             true,
	}

	bz, err := json.Marshal(metadata)
	require.NoError(t, err)

	var decoded ChannelMetadata
	err = json.Unmarshal(bz, &decoded)
	require.NoError(t, err)
	require.Equal(t, metadata.CounterpartyChainID, decoded.CounterpartyChainID)
	require.Equal(t, metadata.Version, decoded.Version)
	require.True(t, decoded.Trusted)
}

func TestFederatedIdentity(t *testing.T) {
	identity := FederatedIdentity{
		Identity: types.EnclaveIdentity{
			ValidatorAddress: "cosmos1abc...",
			TEEType:          types.TEETypeSGX,
			MeasurementHash:  make([]byte, 32),
			EncryptionPubKey: []byte("pubkey"),
			SigningPubKey:    []byte("sigkey"),
			AttestationQuote: []byte("quote"),
			ExpiryHeight:     1000,
			Status:           types.EnclaveIdentityStatusActive,
		},
		SourceChainID:   "cosmoshub-4",
		SourceChannelID: "channel-0",
		ReceivedHeight:  500,
		Verified:        true,
	}

	bz, err := json.Marshal(identity)
	require.NoError(t, err)

	var decoded FederatedIdentity
	err = json.Unmarshal(bz, &decoded)
	require.NoError(t, err)
	require.Equal(t, identity.SourceChainID, decoded.SourceChainID)
	require.Equal(t, identity.SourceChannelID, decoded.SourceChannelID)
	require.True(t, decoded.Verified)
}

func TestFederatedMeasurement(t *testing.T) {
	measurement := FederatedMeasurement{
		Measurement: types.MeasurementRecord{
			MeasurementHash: make([]byte, 32),
			TEEType:         types.TEETypeSGX,
			Description:     "Test measurement",
			MinISVSVN:       1,
		},
		SourceChainID:   "cosmoshub-4",
		SourceChannelID: "channel-0",
		ReceivedHeight:  500,
		Trusted:         true,
	}

	bz, err := json.Marshal(measurement)
	require.NoError(t, err)

	var decoded FederatedMeasurement
	err = json.Unmarshal(bz, &decoded)
	require.NoError(t, err)
	require.Equal(t, measurement.SourceChainID, decoded.SourceChainID)
	require.True(t, decoded.Trusted)
}

func TestPacketDataGetBytes(t *testing.T) {
	packet := EnclavePacketData{
		Type: PacketTypeQueryEnclaveIdentity,
		Data: json.RawMessage(`{"validator_address":"cosmos1..."}`),
	}

	bz := packet.GetBytes()
	require.NotEmpty(t, bz)

	var decoded EnclavePacketData
	err := json.Unmarshal(bz, &decoded)
	require.NoError(t, err)
	require.Equal(t, packet.Type, decoded.Type)
}

func TestAcknowledgementGetBytes(t *testing.T) {
	ack := NewResultAcknowledgement(map[string]string{"key": "value"})
	bz := ack.GetBytes()
	require.NotEmpty(t, bz)

	var decoded Acknowledgement
	err := json.Unmarshal(bz, &decoded)
	require.NoError(t, err)
	require.True(t, decoded.Success())
}

func TestValidateEnclavePacket(t *testing.T) {
	validPacket := EnclavePacketData{
		Type: PacketTypeQueryEnclaveIdentity,
		Data: json.RawMessage(`{}`),
	}
	require.NoError(t, ValidateEnclavePacket(validPacket))

	invalidPacket := EnclavePacketData{
		Type: "invalid",
		Data: json.RawMessage(`{}`),
	}
	require.Error(t, ValidateEnclavePacket(invalidPacket))
}

func TestAllPacketTypes(t *testing.T) {
	packetTypes := []PacketType{
		PacketTypeQueryEnclaveIdentity,
		PacketTypeQueryMeasurementAllowlist,
		PacketTypeSyncMeasurement,
		PacketTypeEnclaveIdentityResponse,
		PacketTypeMeasurementAllowlistResponse,
		PacketTypeSyncMeasurementAck,
	}

	for _, pt := range packetTypes {
		packet := EnclavePacketData{
			Type: pt,
			Data: json.RawMessage(`{}`),
		}
		require.NoError(t, packet.Validate(), "packet type %s should be valid", pt)
	}
}

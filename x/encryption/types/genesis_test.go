package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisState(t *testing.T) {
	gs := DefaultGenesisState()

	require.NotNil(t, gs)
	assert.Empty(t, gs.RecipientKeys)
	assert.NotNil(t, gs.Params)
}

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	assert.Equal(t, uint32(10), params.MaxRecipientsPerEnvelope)
	assert.Equal(t, uint32(5), params.MaxKeysPerAccount)
	assert.True(t, params.RequireSignature)
	assert.NotEmpty(t, params.AllowedAlgorithms)
}

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		name      string
		state     GenesisState
		expectErr bool
	}{
		{
			name:      "valid default state",
			state:     *DefaultGenesisState(),
			expectErr: false,
		},
		{
			name: "valid state with keys",
			state: GenesisState{
				RecipientKeys: []RecipientKeyRecord{
					{
						Address:        "cosmos1xyz...",
						PublicKey:      make([]byte, 32),
						KeyFingerprint: "abc123",
						AlgorithmID:    AlgorithmX25519XSalsa20Poly1305,
					},
				},
				Params: DefaultParams(),
			},
			expectErr: false,
		},
		{
			name: "duplicate key fingerprints",
			state: GenesisState{
				RecipientKeys: []RecipientKeyRecord{
					{
						Address:        "cosmos1xyz...",
						PublicKey:      make([]byte, 32),
						KeyFingerprint: "abc123",
						AlgorithmID:    AlgorithmX25519XSalsa20Poly1305,
					},
					{
						Address:        "cosmos1abc...",
						PublicKey:      make([]byte, 32),
						KeyFingerprint: "abc123", // Duplicate
						AlgorithmID:    AlgorithmX25519XSalsa20Poly1305,
					},
				},
				Params: DefaultParams(),
			},
			expectErr: true,
		},
		{
			name: "invalid key record",
			state: GenesisState{
				RecipientKeys: []RecipientKeyRecord{
					{
						Address:        "", // Invalid
						PublicKey:      make([]byte, 32),
						KeyFingerprint: "abc123",
						AlgorithmID:    AlgorithmX25519XSalsa20Poly1305,
					},
				},
				Params: DefaultParams(),
			},
			expectErr: true,
		},
		{
			name: "invalid params",
			state: GenesisState{
				RecipientKeys: []RecipientKeyRecord{},
				Params: Params{
					MaxRecipientsPerEnvelope: 0, // Invalid
					MaxKeysPerAccount:        5,
				},
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.state.Validate()

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParams_Validate(t *testing.T) {
	tests := []struct {
		name      string
		params    Params
		expectErr bool
	}{
		{
			name:      "valid default params",
			params:    DefaultParams(),
			expectErr: false,
		},
		{
			name: "zero max recipients",
			params: Params{
				MaxRecipientsPerEnvelope: 0,
				MaxKeysPerAccount:        5,
			},
			expectErr: true,
		},
		{
			name: "zero max keys",
			params: Params{
				MaxRecipientsPerEnvelope: 10,
				MaxKeysPerAccount:        0,
			},
			expectErr: true,
		},
		{
			name: "invalid algorithm in list",
			params: Params{
				MaxRecipientsPerEnvelope: 10,
				MaxKeysPerAccount:        5,
				AllowedAlgorithms:        []string{"INVALID-ALGO"},
			},
			expectErr: true,
		},
		{
			name: "valid custom algorithms",
			params: Params{
				MaxRecipientsPerEnvelope: 10,
				MaxKeysPerAccount:        5,
				AllowedAlgorithms:        []string{AlgorithmX25519XSalsa20Poly1305},
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParams_IsAlgorithmAllowed(t *testing.T) {
	// With empty allowed list, all supported algorithms are allowed
	emptyParams := Params{
		MaxRecipientsPerEnvelope: 10,
		MaxKeysPerAccount:        5,
		AllowedAlgorithms:        []string{},
	}

	assert.True(t, emptyParams.IsAlgorithmAllowed(AlgorithmX25519XSalsa20Poly1305))
	assert.False(t, emptyParams.IsAlgorithmAllowed("UNKNOWN"))

	// With specific list, only those are allowed
	restrictedParams := Params{
		MaxRecipientsPerEnvelope: 10,
		MaxKeysPerAccount:        5,
		AllowedAlgorithms:        []string{AlgorithmX25519XSalsa20Poly1305},
	}

	assert.True(t, restrictedParams.IsAlgorithmAllowed(AlgorithmX25519XSalsa20Poly1305))
	assert.False(t, restrictedParams.IsAlgorithmAllowed(AlgorithmAgeX25519)) // Not in list
}

// Package types contains tests for the delegation module types.
//
// VE-922: Delegation types tests
package types

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Test address constants - valid bech32 addresses
var (
	testDelegatorAddr1 = sdk.AccAddress([]byte("delegator_address_01")).String()
	testValidatorAddr1 = sdk.AccAddress([]byte("validator_address_01")).String()
	testValidatorAddr2 = sdk.AccAddress([]byte("validator_address_02")).String()
)

// TestDelegationValidation tests delegation validation
func TestDelegationValidation(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name        string
		delegation  Delegation
		expectError bool
	}{
		{
			name: "valid delegation",
			delegation: Delegation{
				DelegatorAddress: testDelegatorAddr1,
				ValidatorAddress: testValidatorAddr1,
				Shares:           "1000000000000000000",
				InitialAmount:    "1000000",
				CreatedAt:        now,
				UpdatedAt:        now,
				Height:           100,
			},
			expectError: false,
		},
		{
			name: "empty delegator",
			delegation: Delegation{
				DelegatorAddress: "",
				ValidatorAddress: testValidatorAddr1,
				Shares:           "1000000000000000000",
				InitialAmount:    "1000000",
			},
			expectError: true,
		},
		{
			name: "empty validator",
			delegation: Delegation{
				DelegatorAddress: testDelegatorAddr1,
				ValidatorAddress: "",
				Shares:           "1000000000000000000",
				InitialAmount:    "1000000",
			},
			expectError: true,
		},
		{
			name: "negative shares",
			delegation: Delegation{
				DelegatorAddress: testDelegatorAddr1,
				ValidatorAddress: testValidatorAddr1,
				Shares:           "-1000",
				InitialAmount:    "1000000",
			},
			expectError: true,
		},
		{
			name: "invalid shares format",
			delegation: Delegation{
				DelegatorAddress: testDelegatorAddr1,
				ValidatorAddress: testValidatorAddr1,
				Shares:           "not-a-number",
				InitialAmount:    "1000000",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.delegation.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestDelegationShareOperations tests share add/subtract operations
func TestDelegationShareOperations(t *testing.T) {
	now := time.Now().UTC()

	del := NewDelegation("delegator", "validator", "1000000000000000000", "1000000", now, 100)

	// Add shares
	err := del.AddShares("500000000000000000", now.Add(time.Hour))
	require.NoError(t, err)
	require.Equal(t, "1500000000000000000", del.Shares)

	// Subtract shares
	err = del.SubtractShares("300000000000000000", now.Add(2*time.Hour))
	require.NoError(t, err)
	require.Equal(t, "1200000000000000000", del.Shares)

	// Try to subtract more than available
	err = del.SubtractShares("2000000000000000000", now)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient shares")
}

// TestValidatorSharesCalculations tests share calculations
func TestValidatorSharesCalculations(t *testing.T) {
	now := time.Now().UTC()

	vs := NewValidatorShares("validator", now)

	// First delegation - shares = amount * 10^18
	shares1, err := vs.CalculateSharesForAmount("1000000")
	require.NoError(t, err)
	require.Equal(t, "1000000000000000000000000", shares1)

	// Add shares
	err = vs.AddShares(shares1, "1000000", now)
	require.NoError(t, err)

	// Second delegation - same amount should get same shares
	shares2, err := vs.CalculateSharesForAmount("1000000")
	require.NoError(t, err)
	require.Equal(t, shares1, shares2)

	// Calculate amount for shares
	amount, err := vs.CalculateAmountForShares("500000000000000000000000")
	require.NoError(t, err)
	require.Equal(t, "500000", amount) // Half the shares = half the amount
}

// TestValidatorSharesAddSubtract tests add/subtract operations on validator shares
func TestValidatorSharesAddSubtract(t *testing.T) {
	now := time.Now().UTC()

	vs := NewValidatorShares("validator", now)

	// Add shares
	err := vs.AddShares("1000000000000000000", "1000000", now)
	require.NoError(t, err)
	require.Equal(t, "1000000000000000000", vs.TotalShares)
	require.Equal(t, "1000000", vs.TotalStake)

	// Add more
	err = vs.AddShares("500000000000000000", "500000", now)
	require.NoError(t, err)
	require.Equal(t, "1500000000000000000", vs.TotalShares)
	require.Equal(t, "1500000", vs.TotalStake)

	// Subtract
	err = vs.SubtractShares("300000000000000000", "300000", now)
	require.NoError(t, err)
	require.Equal(t, "1200000000000000000", vs.TotalShares)
	require.Equal(t, "1200000", vs.TotalStake)

	// Try to subtract more than available
	err = vs.SubtractShares("2000000000000000000", "2000000", now)
	require.Error(t, err)
}

// TestUnbondingDelegationValidation tests unbonding delegation validation
func TestUnbondingDelegationValidation(t *testing.T) {
	now := time.Now().UTC()
	completionTime := now.Add(21 * 24 * time.Hour)

	tests := []struct {
		name        string
		ubd         UnbondingDelegation
		expectError bool
	}{
		{
			name: "valid unbonding",
			ubd: *NewUnbondingDelegation(
				"ubd-123",
				"delegator",
				"validator",
				100,
				completionTime,
				now,
				"1000000",
				"1000000000000000000",
			),
			expectError: false,
		},
		{
			name: "empty id",
			ubd: UnbondingDelegation{
				ID:               "",
				DelegatorAddress: "delegator",
				ValidatorAddress: "validator",
				Entries:          []UnbondingDelegationEntry{{Balance: "1000"}},
			},
			expectError: true,
		},
		{
			name: "empty entries",
			ubd: UnbondingDelegation{
				ID:               "ubd-123",
				DelegatorAddress: "delegator",
				ValidatorAddress: "validator",
				Entries:          []UnbondingDelegationEntry{},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ubd.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestRedelegationValidation tests redelegation validation
func TestRedelegationValidation(t *testing.T) {
	now := time.Now().UTC()
	completionTime := now.Add(21 * 24 * time.Hour)

	tests := []struct {
		name        string
		red         Redelegation
		expectError bool
	}{
		{
			name: "valid redelegation",
			red: *NewRedelegation(
				"red-123",
				"delegator",
				"validator1",
				"validator2",
				100,
				completionTime,
				now,
				"1000000",
				"1000000000000000000",
			),
			expectError: false,
		},
		{
			name: "same src and dst validator",
			red: *NewRedelegation(
				"red-123",
				"delegator",
				"validator1",
				"validator1",
				100,
				completionTime,
				now,
				"1000000",
				"1000000000000000000",
			),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.red.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestDelegatorRewardValidation tests delegator reward validation
func TestDelegatorRewardValidation(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name        string
		reward      DelegatorReward
		expectError bool
	}{
		{
			name: "valid reward",
			reward: *NewDelegatorReward(
				"delegator",
				"validator",
				10,
				"500000",
				"1000000000000000000",
				"10000000000000000000",
				now,
			),
			expectError: false,
		},
		{
			name: "empty delegator",
			reward: DelegatorReward{
				DelegatorAddress: "",
				ValidatorAddress: "validator",
				Reward:           "500000",
			},
			expectError: true,
		},
		{
			name: "negative reward",
			reward: DelegatorReward{
				DelegatorAddress: "delegator",
				ValidatorAddress: "validator",
				Reward:           "-100",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.reward.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestParamsValidation tests parameter validation
func TestParamsValidation(t *testing.T) {
	tests := []struct {
		name        string
		params      Params
		expectError bool
	}{
		{
			name:        "default params",
			params:      DefaultParams(),
			expectError: false,
		},
		{
			name: "zero unbonding period",
			params: Params{
				UnbondingPeriod:           0,
				MaxValidatorsPerDelegator: 10,
				MinDelegationAmount:       1000000,
				MaxRedelegations:          7,
				ValidatorCommissionRate:   1000,
				RewardDenom:               "uve",
				StakeDenom:                "uve",
			},
			expectError: true,
		},
		{
			name: "negative unbonding period",
			params: Params{
				UnbondingPeriod:           -1,
				MaxValidatorsPerDelegator: 10,
				MinDelegationAmount:       1000000,
				MaxRedelegations:          7,
				ValidatorCommissionRate:   1000,
				RewardDenom:               "uve",
				StakeDenom:                "uve",
			},
			expectError: true,
		},
		{
			name: "commission rate too high",
			params: Params{
				UnbondingPeriod:           DefaultUnbondingPeriod,
				MaxValidatorsPerDelegator: 10,
				MinDelegationAmount:       1000000,
				MaxRedelegations:          7,
				ValidatorCommissionRate:   15000, // > 10000 basis points
				RewardDenom:               "uve",
				StakeDenom:                "uve",
			},
			expectError: true,
		},
		{
			name: "empty reward denom",
			params: Params{
				UnbondingPeriod:           DefaultUnbondingPeriod,
				MaxValidatorsPerDelegator: 10,
				MinDelegationAmount:       1000000,
				MaxRedelegations:          7,
				ValidatorCommissionRate:   1000,
				RewardDenom:               "",
				StakeDenom:                "uve",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateParams(&tc.params)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestGenesisStateValidation tests genesis state validation
func TestGenesisStateValidation(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name        string
		genesis     GenesisState
		expectError bool
	}{
		{
			name:        "default genesis",
			genesis:     *DefaultGenesisState(),
			expectError: false,
		},
		{
			name: "invalid params",
			genesis: GenesisState{
				Params: Params{
					UnbondingPeriod: -1, // Invalid
				},
			},
			expectError: true,
		},
		{
			name: "invalid delegation",
			genesis: GenesisState{
				Params: DefaultParams(),
				Delegations: []Delegation{
					{
						DelegatorAddress: "", // Invalid
						ValidatorAddress: "validator",
					},
				},
			},
			expectError: true,
		},
		{
			name: "valid with data",
			genesis: GenesisState{
				Params: DefaultParams(),
				Delegations: []Delegation{
					*NewDelegation("delegator", "validator", "1000000000000000000", "1000000", now, 100),
				},
				ValidatorShares: []ValidatorShares{
					*NewValidatorShares("validator", now),
				},
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestMsgDelegateValidation tests delegate message validation
func TestMsgDelegateValidation(t *testing.T) {
	tests := []struct {
		name        string
		msg         *MsgDelegate
		expectError bool
	}{
		{
			name: "valid message",
			msg: NewMsgDelegate(
				testDelegatorAddr1,
				testValidatorAddr1,
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectError: false,
		},
		{
			name: "invalid delegator",
			msg: NewMsgDelegate(
				"invalid",
				testValidatorAddr1,
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectError: true,
		},
		{
			name: "zero amount",
			msg: NewMsgDelegate(
				testDelegatorAddr1,
				testValidatorAddr1,
				sdk.NewCoin("uve", sdkmath.NewInt(0)),
			),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestMsgRedelegateValidation tests redelegate message validation
func TestMsgRedelegateValidation(t *testing.T) {
	tests := []struct {
		name        string
		msg         *MsgRedelegate
		expectError bool
	}{
		{
			name: "valid message",
			msg: NewMsgRedelegate(
				testDelegatorAddr1,
				testValidatorAddr1,
				testValidatorAddr2,
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectError: false,
		},
		{
			name: "same source and destination",
			msg: NewMsgRedelegate(
				testDelegatorAddr1,
				testValidatorAddr1,
				testValidatorAddr1,
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestUnbondingDelegationTotalBalance tests total balance calculation
func TestUnbondingDelegationTotalBalance(t *testing.T) {
	now := time.Now().UTC()

	ubd := &UnbondingDelegation{
		ID:               "ubd-123",
		DelegatorAddress: "delegator",
		ValidatorAddress: "validator",
		Entries: []UnbondingDelegationEntry{
			{Balance: "100000"},
			{Balance: "200000"},
			{Balance: "300000"},
		},
		CreatedAt: now,
	}

	total := ubd.TotalBalance()
	require.Equal(t, "600000", total.String())
}

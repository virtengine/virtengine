//go:build ignore
// +build ignore

// TODO: This test file is excluded until settlement events API is stabilized.

package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/settlement/types"
)

func TestEventEmission(t *testing.T) {
	t.Run("escrow created event", func(t *testing.T) {
		event := types.EventEscrowCreated{
			EscrowID:   "escrow-1",
			OrderID:    "order-1",
			Depositor:  "cosmos1depositor...",
			Amount:     "1000uve",
			ExpiresAt:  time.Now().Add(time.Hour * 24).Format(time.RFC3339),
			Conditions: 2,
		}

		require.Equal(t, "escrow-1", event.EscrowID)
		require.Equal(t, "order-1", event.OrderID)
	})

	t.Run("escrow activated event", func(t *testing.T) {
		event := types.EventEscrowActivated{
			EscrowID:    "escrow-1",
			OrderID:     "order-1",
			LeaseID:     "lease-1",
			Recipient:   "cosmos1recipient...",
			ActivatedAt: time.Now().Format(time.RFC3339),
		}

		require.Equal(t, "escrow-1", event.EscrowID)
		require.Equal(t, "lease-1", event.LeaseID)
	})

	t.Run("escrow released event", func(t *testing.T) {
		event := types.EventEscrowReleased{
			EscrowID:   "escrow-1",
			OrderID:    "order-1",
			Recipient:  "cosmos1recipient...",
			Amount:     "1000uve",
			Reason:     "service completed",
			ReleasedAt: time.Now().Format(time.RFC3339),
		}

		require.Equal(t, "escrow-1", event.EscrowID)
		require.Equal(t, "service completed", event.Reason)
	})

	t.Run("order settled event", func(t *testing.T) {
		event := types.EventOrderSettled{
			SettlementID:   "settlement-1",
			OrderID:        "order-1",
			EscrowID:       "escrow-1",
			Provider:       "cosmos1provider...",
			Customer:       "cosmos1customer...",
			TotalAmount:    "1000uve",
			PlatformFee:    "50uve",
			ValidatorFee:   "10uve",
			ProviderPayout: "940uve",
			Type:           string(types.SettlementTypePeriodic),
			SettledAt:      time.Now().Format(time.RFC3339),
		}

		require.Equal(t, "settlement-1", event.SettlementID)
		require.Equal(t, "order-1", event.OrderID)
	})

	t.Run("usage recorded event", func(t *testing.T) {
		event := types.EventUsageRecorded{
			UsageID:     "usage-1",
			OrderID:     "order-1",
			Provider:    "cosmos1provider...",
			Customer:    "cosmos1customer...",
			ComputeUsed: "1000",
			StorageUsed: "500",
			TotalCost:   "100uve",
			PeriodStart: time.Now().Add(-time.Hour).Format(time.RFC3339),
			PeriodEnd:   time.Now().Format(time.RFC3339),
		}

		require.Equal(t, "usage-1", event.UsageID)
		require.Equal(t, "1000", event.ComputeUsed)
	})

	t.Run("rewards distributed event", func(t *testing.T) {
		event := types.EventRewardsDistributed{
			DistributionID: "dist-1",
			Source:         string(types.RewardSourceStaking),
			EpochNumber:    1,
			TotalAmount:    "1000uve",
			RecipientCount: 10,
			DistributedAt:  time.Now().Format(time.RFC3339),
		}

		require.Equal(t, "dist-1", event.DistributionID)
		require.Equal(t, uint64(1), event.EpochNumber)
	})

	t.Run("rewards claimed event", func(t *testing.T) {
		event := types.EventRewardsClaimed{
			Claimer:        "cosmos1claimer...",
			Source:         string(types.RewardSourceStaking),
			Amount:         "500uve",
			EntriesClaimed: 3,
			ClaimedAt:      time.Now().Format(time.RFC3339),
		}

		require.Equal(t, "cosmos1claimer...", event.Claimer)
		require.Equal(t, "500uve", event.Amount)
	})
}

func TestParamsValidation(t *testing.T) {
	testCases := []struct {
		name        string
		params      types.Params
		expectError bool
	}{
		{
			name:        "default params are valid",
			params:      types.DefaultParams(),
			expectError: false,
		},
		{
			name: "valid custom params",
			params: types.Params{
				PlatformFeeRate:          "0.03",
				ValidatorFeeRate:         "0.02",
				MinEscrowDuration:        3600,
				MaxEscrowDuration:        86400 * 365,
				SettlementPeriod:         86400,
				DisputeResolutionTimeout: 86400 * 7,
				StakingRewardPool:        "1000000",
				StakingRewardEpochLength: 100,
				ProviderRewardShare:      "0.85",
			},
			expectError: false,
		},
		{
			name: "invalid platform fee rate",
			params: types.Params{
				PlatformFeeRate:          "1.5", // > 100%
				ValidatorFeeRate:         "0.01",
				MinEscrowDuration:        3600,
				MaxEscrowDuration:        86400 * 365,
				SettlementPeriod:         86400,
				DisputeResolutionTimeout: 86400 * 7,
			},
			expectError: true,
		},
		{
			name: "min duration greater than max",
			params: types.Params{
				PlatformFeeRate:   "0.05",
				ValidatorFeeRate:  "0.01",
				MinEscrowDuration: 86400 * 400,
				MaxEscrowDuration: 86400,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenesisStateValidation(t *testing.T) {
	testCases := []struct {
		name        string
		genesis     types.GenesisState
		expectError bool
	}{
		{
			name:        "default genesis is valid",
			genesis:     *types.DefaultGenesisState(),
			expectError: false,
		},
		{
			name: "genesis with escrows",
			genesis: types.GenesisState{
				Params: types.DefaultParams(),
				Escrows: []types.EscrowAccount{
					{
						EscrowID:  "escrow-1",
						OrderID:   "order-1",
						Depositor: "cosmos1depositor...",
						Amount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
						State:     types.EscrowStatePending,
						CreatedAt: time.Now(),
						ExpiresAt: time.Now().Add(time.Hour * 24),
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid params in genesis",
			genesis: types.GenesisState{
				Params: types.Params{
					PlatformFeeRate: "invalid",
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
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

func TestMsgValidation(t *testing.T) {
	t.Run("MsgCreateEscrow", func(t *testing.T) {
		msg := types.MsgCreateEscrow{
			Depositor:  "cosmos1depositor...",
			OrderID:    "order-1",
			Amount:     sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			Duration:   86400,
			Conditions: nil,
		}
		err := msg.ValidateBasic()
		require.NoError(t, err)

		// Empty depositor
		msg.Depositor = ""
		err = msg.ValidateBasic()
		require.Error(t, err)
	})

	t.Run("MsgReleaseEscrow", func(t *testing.T) {
		msg := types.MsgReleaseEscrow{
			Sender:   "cosmos1sender...",
			EscrowID: "escrow-1",
			Reason:   "service completed",
		}
		err := msg.ValidateBasic()
		require.NoError(t, err)

		// Empty escrow ID
		msg.EscrowID = ""
		err = msg.ValidateBasic()
		require.Error(t, err)
	})

	t.Run("MsgSettleOrder", func(t *testing.T) {
		msg := types.MsgSettleOrder{
			Sender:  "cosmos1sender...",
			OrderID: "order-1",
			IsFinal: true,
		}
		err := msg.ValidateBasic()
		require.NoError(t, err)

		// Empty sender
		msg.Sender = ""
		err = msg.ValidateBasic()
		require.Error(t, err)
	})

	t.Run("MsgRecordUsage", func(t *testing.T) {
		msg := types.MsgRecordUsage{
			Provider:    "cosmos1provider...",
			OrderID:     "order-1",
			ComputeUsed: "1000",
			StorageUsed: "500",
			TotalCost:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100))),
			PeriodStart: time.Now().Add(-time.Hour),
			PeriodEnd:   time.Now(),
		}
		err := msg.ValidateBasic()
		require.NoError(t, err)

		// Invalid period (start after end)
		msg.PeriodStart = time.Now()
		msg.PeriodEnd = time.Now().Add(-time.Hour)
		err = msg.ValidateBasic()
		require.Error(t, err)
	})

	t.Run("MsgClaimRewards", func(t *testing.T) {
		msg := types.MsgClaimRewards{
			Claimer: "cosmos1claimer...",
			Source:  string(types.RewardSourceStaking),
		}
		err := msg.ValidateBasic()
		require.NoError(t, err)

		// Empty claimer
		msg.Claimer = ""
		err = msg.ValidateBasic()
		require.Error(t, err)
	})
}

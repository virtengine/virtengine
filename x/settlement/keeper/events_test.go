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
	now := time.Now()

	t.Run("escrow created event", func(t *testing.T) {
		event := types.EventEscrowCreated{
			EscrowID:    "escrow-1",
			OrderID:     "order-1",
			Depositor:   "cosmos1depositor...",
			Amount:      "1000uve",
			ExpiresAt:   now.Add(time.Hour * 24).Unix(),
			BlockHeight: 100,
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
			ActivatedAt: now.Unix(),
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
			ReleasedAt: now.Unix(),
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
			ProviderShare:  "940uve",
			PlatformFee:    "50uve",
			SettlementType: string(types.SettlementTypePeriodic),
			IsFinal:        false,
			SettledAt:      now.Unix(),
		}

		require.Equal(t, "settlement-1", event.SettlementID)
		require.Equal(t, "order-1", event.OrderID)
	})

	t.Run("usage recorded event", func(t *testing.T) {
		event := types.EventUsageRecorded{
			UsageID:    "usage-1",
			OrderID:    "order-1",
			LeaseID:    "lease-1",
			Provider:   "cosmos1provider...",
			UsageUnits: 1000,
			UsageType:  "compute",
			TotalCost:  "100uve",
			RecordedAt: now.Unix(),
		}

		require.Equal(t, "usage-1", event.UsageID)
		require.Equal(t, uint64(1000), event.UsageUnits)
	})

	t.Run("rewards distributed event", func(t *testing.T) {
		event := types.EventRewardsDistributed{
			DistributionID: "dist-1",
			Source:         string(types.RewardSourceStaking),
			EpochNumber:    1,
			TotalRewards:   "1000uve",
			RecipientCount: 10,
			DistributedAt:  now.Unix(),
		}

		require.Equal(t, "dist-1", event.DistributionID)
		require.Equal(t, uint64(1), event.EpochNumber)
	})

	t.Run("rewards claimed event", func(t *testing.T) {
		event := types.EventRewardsClaimed{
			Claimer:       "cosmos1claimer...",
			Source:        string(types.RewardSourceStaking),
			ClaimedAmount: "500uve",
			ClaimedAt:     now.Unix(),
		}

		require.Equal(t, "cosmos1claimer...", event.Claimer)
		require.Equal(t, "500uve", event.ClaimedAmount)
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
				PlatformFeeRate:                        "0.03",
				ValidatorFeeRate:                       "0.02",
				MinEscrowDuration:                      3600,
				MaxEscrowDuration:                      86400 * 365,
				SettlementPeriod:                       86400,
				RewardClaimExpiry:                      86400 * 30,
				StakingRewardEpochLength:               100,
				DisputeWindowDuration:                  86400 * 7,
				UsageRewardRateBps:                     500,
				UsageRewardCPUMultiplierBps:            10000,
				UsageRewardMemoryMultiplierBps:         10000,
				UsageRewardStorageMultiplierBps:        10000,
				UsageRewardGPUMultiplierBps:            10000,
				UsageRewardNetworkMultiplierBps:        10000,
				UsageRewardSLAOnTimeMultiplierBps:      10000,
				UsageRewardSLALateMultiplierBps:        9000,
				UsageRewardAcknowledgedMultiplierBps:   10000,
				UsageRewardUnacknowledgedMultiplierBps: 9000,
			},
			expectError: false,
		},
		{
			name: "invalid platform fee rate",
			params: types.Params{
				PlatformFeeRate:       "1.5", // > 100%
				ValidatorFeeRate:      "0.01",
				MinEscrowDuration:     3600,
				MaxEscrowDuration:     86400 * 365,
				SettlementPeriod:      86400,
				DisputeWindowDuration: 86400 * 7,
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
	validAddr := sdk.AccAddress([]byte("test_address________")).String()

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
				EscrowAccounts: []types.EscrowAccount{
					{
						EscrowID:  "escrow-1",
						OrderID:   "order-1",
						Depositor: validAddr,
						Amount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
						Balance:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
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
	validAddr := sdk.AccAddress([]byte("test_sender_________")).String()

	t.Run("MsgCreateEscrow", func(t *testing.T) {
		msg := types.NewMsgCreateEscrow(
			validAddr,
			"order-1",
			sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			86400,
		)
		err := msg.ValidateBasic()
		require.NoError(t, err)

		// Empty sender
		msg.Sender = ""
		err = msg.ValidateBasic()
		require.Error(t, err)
	})

	t.Run("MsgReleaseEscrow", func(t *testing.T) {
		msg := types.MsgReleaseEscrow{
			Sender:   validAddr,
			EscrowId: "escrow-1",
			Reason:   "service completed",
		}
		err := msg.ValidateBasic()
		require.NoError(t, err)

		// Empty escrow ID
		msg.EscrowId = ""
		err = msg.ValidateBasic()
		require.Error(t, err)
	})

	t.Run("MsgSettleOrder", func(t *testing.T) {
		msg := types.MsgSettleOrder{
			Sender:  validAddr,
			OrderId: "order-1",
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
			Sender:      validAddr,
			OrderId:     "order-1",
			LeaseId:     "lease-1",
			UsageUnits:  1000,
			UsageType:   "compute",
			PeriodStart: time.Now().Add(-time.Hour).Unix(),
			PeriodEnd:   time.Now().Unix(),
			Signature:   []byte("provider-signature"),
		}
		err := msg.ValidateBasic()
		require.NoError(t, err)

		// Invalid period (start after end)
		msg.PeriodStart = time.Now().Unix()
		msg.PeriodEnd = time.Now().Add(-time.Hour).Unix()
		err = msg.ValidateBasic()
		require.Error(t, err)
	})

	t.Run("MsgClaimRewards", func(t *testing.T) {
		msg := types.MsgClaimRewards{
			Sender: validAddr,
			Source: string(types.RewardSourceStaking),
		}
		err := msg.ValidateBasic()
		require.NoError(t, err)

		// Empty sender
		msg.Sender = ""
		err = msg.ValidateBasic()
		require.Error(t, err)
	})
}

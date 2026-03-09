package types

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalculateRewardsCapsPerfectValidatorAtEpochPool(t *testing.T) {
	perf := NewValidatorPerformance("validator-1", 1)
	perf.BlocksProposed = 100
	perf.BlocksExpected = 100
	perf.VEIDVerificationsCompleted = 10
	perf.VEIDVerificationsExpected = 10
	perf.VEIDVerificationScore = MaxPerformanceScore
	perf.UptimeSeconds = 3600
	ComputeOverallScore(perf)

	reward := CalculateRewards(RewardCalculationInput{
		ValidatorAddress: "validator-1",
		Performance:      perf,
		StakeAmount:      1_000_000,
		TotalStake:       1_000_000,
		EpochRewardPool:  100_000_000,
	}, "uve")

	require.Equal(t, int64(30_000_000), reward.BlockProposalReward.AmountOf("uve").Int64())
	require.Equal(t, int64(40_000_000), reward.VEIDReward.AmountOf("uve").Int64())
	require.Equal(t, int64(30_000_000), reward.UptimeReward.AmountOf("uve").Int64())
	require.Equal(t, int64(100_000_000), reward.TotalReward.AmountOf("uve").Int64())
}

func TestCalculateRewardsAllocatesRemainderDeterministically(t *testing.T) {
	perf := NewValidatorPerformance("validator-remainder", 1)
	perf.BlocksProposed = 10
	perf.BlocksExpected = 10
	perf.VEIDVerificationsCompleted = 10
	perf.VEIDVerificationsExpected = 10
	perf.VEIDVerificationScore = MaxPerformanceScore
	perf.UptimeSeconds = 3600
	ComputeOverallScore(perf)

	reward := CalculateRewards(RewardCalculationInput{
		ValidatorAddress: "validator-remainder",
		Performance:      perf,
		StakeAmount:      1,
		TotalStake:       2,
		EpochRewardPool:  3,
		BlocksInEpoch:    3,
	}, "uve")

	require.Equal(t, int64(1), reward.TotalReward.AmountOf("uve").Int64())
	require.True(t, reward.BlockProposalReward.IsZero())
	require.Equal(t, int64(1), reward.VEIDReward.AmountOf("uve").Int64())
	require.True(t, reward.UptimeReward.IsZero())
}

func TestCalculateIdentityNetworkRewardIsDeterministic(t *testing.T) {
	input := IdentityNetworkRewardInput{
		ValidatorAddress:         "validator-1",
		VerificationsCompleted:   10,
		TotalVerifications:       100,
		AverageVerificationScore: 9000,
		RewardPool:               100_000_000,
	}

	first := CalculateIdentityNetworkReward(input, "uve")
	second := CalculateIdentityNetworkReward(input, "uve")

	require.True(t, first.Equal(second))
	require.Equal(t, int64(9_000_000), first.AmountOf("uve").Int64())
}

func TestCalculateRewardsHandlesLargeStakeWithoutOverflow(t *testing.T) {
	perf := NewValidatorPerformance("validator-large", 1)
	perf.BlocksProposed = 100
	perf.BlocksExpected = 100
	perf.VEIDVerificationsCompleted = 10
	perf.VEIDVerificationsExpected = 10
	perf.VEIDVerificationScore = MaxPerformanceScore
	perf.UptimeSeconds = 3600
	ComputeOverallScore(perf)

	rewardPool := int64(math.MaxInt64 / 4)
	reward := CalculateRewards(RewardCalculationInput{
		ValidatorAddress: "validator-large",
		Performance:      perf,
		StakeAmount:      math.MaxInt64,
		TotalStake:       math.MaxInt64,
		EpochRewardPool:  rewardPool,
	}, "uve")

	require.Equal(t, rewardPool, reward.TotalReward.AmountOf("uve").Int64())
	require.Equal(t, int64(FixedPointScale), mustParseInt64(t, reward.StakeWeight))
}

func TestGetSlashConfigForParamsUsesOnChainEconomics(t *testing.T) {
	params := DefaultParams()
	params.SlashFractionDoubleSign = 55_000
	params.JailDurationDoubleSign = 7200
	params.SlashFractionDowntime = 2_500
	params.JailDurationDowntime = 900
	params.SlashFractionInvalidAttestation = 75_000
	params.JailDurationInvalidAttestation = 1800

	doubleSign := GetSlashConfigForParams(SlashReasonDoubleSigning, params)
	require.Equal(t, int64(55_000), doubleSign.SlashPercent)
	require.Equal(t, int64(7200), doubleSign.JailDuration)
	require.True(t, doubleSign.IsTombstone)

	downtime := GetSlashConfigForParams(SlashReasonDowntime, params)
	require.Equal(t, int64(2_500), downtime.SlashPercent)
	require.Equal(t, int64(900), downtime.JailDuration)
	require.False(t, downtime.IsTombstone)

	attestation := GetSlashConfigForParams(SlashReasonInvalidVEIDAttestation, params)
	require.Equal(t, int64(75_000), attestation.SlashPercent)
	require.Equal(t, int64(1800), attestation.JailDuration)
	require.False(t, attestation.IsTombstone)
}

func mustParseInt64(t *testing.T, value string) int64 {
	t.Helper()
	parsed, ok := new(big.Int).SetString(value, 10)
	require.True(t, ok)
	require.True(t, parsed.IsInt64())
	return parsed.Int64()
}

// Package keeper implements the staking module keeper tests.
//
// VE-921: Staking rewards keeper tests
package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"pkg.akt.dev/node/x/staking/types"
)

// StakingKeeperTestSuite is the test suite for the staking keeper
type StakingKeeperTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper Keeper
	cdc    codec.BinaryCodec
	skey   *storetypes.KVStoreKey
}

// SetupTest sets up the test suite
func (s *StakingKeeperTestSuite) SetupTest() {
	s.skey = storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(s.skey, storetypes.StoreTypeIAVL, db)
	require.NoError(s.T(), stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	s.cdc = codec.NewProtoCodec(registry)

	s.keeper = NewKeeper(
		s.cdc,
		s.skey,
		nil, // bankKeeper
		nil, // veidKeeper
		nil, // stakingKeeper
		"authority",
	)

	s.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Height: 100,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	// Initialize default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	require.NoError(s.T(), err)
}

// TestStakingKeeperTestSuite runs the test suite
func TestStakingKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(StakingKeeperTestSuite))
}

// TestParams tests parameter management
func (s *StakingKeeperTestSuite) TestParams() {
	// Get default params
	params := s.keeper.GetParams(s.ctx)
	s.Require().Equal(uint64(100), params.EpochLength)
	s.Require().Equal(int64(1000000), params.BaseRewardPerBlock)
	s.Require().Equal("uve", params.RewardDenom)

	// Update params
	params.EpochLength = 200
	params.BaseRewardPerBlock = 2000000
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Verify update
	updatedParams := s.keeper.GetParams(s.ctx)
	s.Require().Equal(uint64(200), updatedParams.EpochLength)
	s.Require().Equal(int64(2000000), updatedParams.BaseRewardPerBlock)
}

// TestValidatorPerformance tests validator performance tracking
func (s *StakingKeeperTestSuite) TestValidatorPerformance() {
	validatorAddr := "validator1qperwt9wrnkg39mvs5g6ys"

	// Initially no performance record
	_, found := s.keeper.GetValidatorPerformance(s.ctx, validatorAddr, 1)
	s.Require().False(found)

	// Create performance record
	perf := types.NewValidatorPerformance(validatorAddr, 1)
	perf.BlocksProposed = 10
	perf.BlocksMissed = 2
	perf.VEIDVerificationsCompleted = 5
	perf.VEIDVerificationScore = 9500

	err := s.keeper.SetValidatorPerformance(s.ctx, *perf)
	s.Require().NoError(err)

	// Get performance
	retrieved, found := s.keeper.GetValidatorPerformance(s.ctx, validatorAddr, 1)
	s.Require().True(found)
	s.Require().Equal(int64(10), retrieved.BlocksProposed)
	s.Require().Equal(int64(2), retrieved.BlocksMissed)
	s.Require().Equal(int64(5), retrieved.VEIDVerificationsCompleted)
	s.Require().Equal(int64(9500), retrieved.VEIDVerificationScore)
}

// TestUpdateValidatorPerformance tests performance updates
func (s *StakingKeeperTestSuite) TestUpdateValidatorPerformance() {
	validatorAddr := "validator1qperwt9wrnkg39mvs5g6ys"

	// Update performance
	update := PerformanceUpdate{
		BlockProposed:            true,
		BlockSigned:              true,
		VEIDVerificationComplete: true,
		VEIDVerificationScore:    9000,
	}
	err := s.keeper.UpdateValidatorPerformance(s.ctx, validatorAddr, update)
	s.Require().NoError(err)

	// Get updated performance
	epoch := s.keeper.GetCurrentEpoch(s.ctx)
	retrieved, found := s.keeper.GetValidatorPerformance(s.ctx, validatorAddr, epoch)
	s.Require().True(found)
	s.Require().Equal(int64(1), retrieved.BlocksProposed)
	s.Require().Equal(int64(1), retrieved.TotalSignatures)
	s.Require().Equal(int64(1), retrieved.VEIDVerificationsCompleted)
	s.Require().Equal(int64(9000), retrieved.VEIDVerificationScore)
}

// TestComputeOverallScore tests overall score computation
func (s *StakingKeeperTestSuite) TestComputeOverallScore() {
	perf := types.NewValidatorPerformance("validator1", 1)
	perf.BlocksProposed = 10
	perf.BlocksExpected = 10
	perf.VEIDVerificationsCompleted = 5
	perf.VEIDVerificationsExpected = 5
	perf.VEIDVerificationScore = 10000 // Perfect score
	perf.UptimeSeconds = 86400
	perf.DowntimeSeconds = 0

	score := perf.ComputeOverallScore()

	// Perfect performance should yield max score
	s.Require().Equal(types.MaxPerformanceScore, score)
	s.Require().Equal(types.MaxPerformanceScore, perf.OverallScore)
}

// TestComputeOverallScorePartial tests partial score computation
func (s *StakingKeeperTestSuite) TestComputeOverallScorePartial() {
	perf := types.NewValidatorPerformance("validator1", 1)
	perf.BlocksProposed = 5
	perf.BlocksExpected = 10 // 50% block proposal
	perf.VEIDVerificationsCompleted = 3
	perf.VEIDVerificationsExpected = 5 // 60% completion
	perf.VEIDVerificationScore = 8000  // 80% quality
	perf.UptimeSeconds = 86400
	perf.DowntimeSeconds = 14400 // 83% uptime

	score := perf.ComputeOverallScore()

	// Score should be between 0 and max
	s.Require().Greater(score, int64(0))
	s.Require().LessOrEqual(score, types.MaxPerformanceScore)
}

// TestEpochManagement tests epoch management
func (s *StakingKeeperTestSuite) TestEpochManagement() {
	// Get initial epoch
	epoch := s.keeper.GetCurrentEpoch(s.ctx)
	s.Require().Equal(uint64(1), epoch)

	// Set new epoch
	s.keeper.SetCurrentEpoch(s.ctx, 5)
	epoch = s.keeper.GetCurrentEpoch(s.ctx)
	s.Require().Equal(uint64(5), epoch)
}

// TestRewardEpoch tests reward epoch management
func (s *StakingKeeperTestSuite) TestRewardEpoch() {
	epochInfo := types.NewRewardEpoch(1, 100, s.ctx.BlockTime())
	epochInfo.ValidatorCount = 10

	err := s.keeper.SetRewardEpoch(s.ctx, *epochInfo)
	s.Require().NoError(err)

	retrieved, found := s.keeper.GetRewardEpoch(s.ctx, 1)
	s.Require().True(found)
	s.Require().Equal(int64(100), retrieved.StartHeight)
	s.Require().Equal(int64(10), retrieved.ValidatorCount)
}

// TestValidatorReward tests validator reward management
func (s *StakingKeeperTestSuite) TestValidatorReward() {
	validatorAddr := "validator1qperwt9wrnkg39mvs5g6ys"

	reward := types.NewValidatorReward(validatorAddr, 1)
	reward.PerformanceScore = 9000
	reward.BlockProposalReward = sdk.NewCoins(sdk.NewInt64Coin("uve", 1000))
	reward.VEIDReward = sdk.NewCoins(sdk.NewInt64Coin("uve", 500))
	reward.UptimeReward = sdk.NewCoins(sdk.NewInt64Coin("uve", 300))
	reward.ComputeTotal()

	err := s.keeper.SetValidatorReward(s.ctx, *reward)
	s.Require().NoError(err)

	retrieved, found := s.keeper.GetValidatorReward(s.ctx, validatorAddr, 1)
	s.Require().True(found)
	s.Require().Equal(int64(9000), retrieved.PerformanceScore)
	s.Require().True(retrieved.TotalReward.IsEqual(sdk.NewCoins(sdk.NewInt64Coin("uve", 1800))))
}

// TestSlashRecord tests slashing record management
func (s *StakingKeeperTestSuite) TestSlashRecord() {
	slashRecord := types.NewSlashRecord(
		"slash-001",
		"validator1",
		types.SlashReasonDowntime,
		sdk.NewCoins(sdk.NewInt64Coin("uve", 1000)),
		1000, // 0.1%
		90,
		100,
		s.ctx.BlockTime(),
	)

	err := s.keeper.SetSlashRecord(s.ctx, *slashRecord)
	s.Require().NoError(err)

	retrieved, found := s.keeper.GetSlashRecord(s.ctx, "slash-001")
	s.Require().True(found)
	s.Require().Equal(types.SlashReasonDowntime, retrieved.Reason)
	s.Require().Equal(int64(90), retrieved.InfractionHeight)
}

// TestValidatorSigningInfo tests signing info management
func (s *StakingKeeperTestSuite) TestValidatorSigningInfo() {
	validatorAddr := "validator1qperwt9wrnkg39mvs5g6ys"

	signingInfo := types.NewValidatorSigningInfo(validatorAddr, 100)
	signingInfo.MissedBlocksCounter = 5

	err := s.keeper.SetValidatorSigningInfo(s.ctx, *signingInfo)
	s.Require().NoError(err)

	retrieved, found := s.keeper.GetValidatorSigningInfo(s.ctx, validatorAddr)
	s.Require().True(found)
	s.Require().Equal(int64(100), retrieved.StartHeight)
	s.Require().Equal(int64(5), retrieved.MissedBlocksCounter)
}

// TestIsValidatorJailed tests jailing checks
func (s *StakingKeeperTestSuite) TestIsValidatorJailed() {
	validatorAddr := "validator1qperwt9wrnkg39mvs5g6ys"

	// Initially not jailed
	jailed := s.keeper.IsValidatorJailed(s.ctx, validatorAddr)
	s.Require().False(jailed)

	// Jail validator
	err := s.keeper.JailValidator(s.ctx, validatorAddr, time.Hour)
	s.Require().NoError(err)

	// Should be jailed
	jailed = s.keeper.IsValidatorJailed(s.ctx, validatorAddr)
	s.Require().True(jailed)
}

// TestTombstone tests tombstoning
func (s *StakingKeeperTestSuite) TestTombstone() {
	validatorAddr := "validator1qperwt9wrnkg39mvs5g6ys"

	// Tombstone validator
	err := s.keeper.TombstoneValidator(s.ctx, validatorAddr)
	s.Require().NoError(err)

	// Should be tombstoned
	signingInfo, found := s.keeper.GetValidatorSigningInfo(s.ctx, validatorAddr)
	s.Require().True(found)
	s.Require().True(signingInfo.Tombstoned)

	// Cannot unjail tombstoned validator
	err = s.keeper.UnjailValidator(s.ctx, validatorAddr)
	s.Require().Error(err)
}

// TestHandleValidatorSignature tests signature handling
func (s *StakingKeeperTestSuite) TestHandleValidatorSignature() {
	validatorAddr := "validator1qperwt9wrnkg39mvs5g6ys"

	// Sign block
	err := s.keeper.HandleValidatorSignature(s.ctx, validatorAddr, true)
	s.Require().NoError(err)

	signingInfo, found := s.keeper.GetValidatorSigningInfo(s.ctx, validatorAddr)
	s.Require().True(found)
	s.Require().Equal(int64(0), signingInfo.MissedBlocksCounter)

	// Miss block
	err = s.keeper.HandleValidatorSignature(s.ctx, validatorAddr, false)
	s.Require().NoError(err)

	signingInfo, _ = s.keeper.GetValidatorSigningInfo(s.ctx, validatorAddr)
	s.Require().Equal(int64(1), signingInfo.MissedBlocksCounter)
}

// TestSlashValidator tests slashing
func (s *StakingKeeperTestSuite) TestSlashValidator() {
	validatorAddr := "validator1qperwt9wrnkg39mvs5g6ys"

	slashRecord, err := s.keeper.SlashValidator(
		s.ctx,
		validatorAddr,
		types.SlashReasonDowntime,
		90,
		"evidence-001",
	)
	s.Require().NoError(err)
	s.Require().NotNil(slashRecord)
	s.Require().Equal(types.SlashReasonDowntime, slashRecord.Reason)
	s.Require().True(slashRecord.Jailed)

	// Check signing info updated
	signingInfo, found := s.keeper.GetValidatorSigningInfo(s.ctx, validatorAddr)
	s.Require().True(found)
	s.Require().Equal(int64(1), signingInfo.InfractionCount)
}

// TestSlashEscalation tests slash escalation for repeat offenders
func (s *StakingKeeperTestSuite) TestSlashEscalation() {
	validatorAddr := "validator1qperwt9wrnkg39mvs5g6ys"

	// First slash
	slashRecord1, err := s.keeper.SlashValidator(
		s.ctx,
		validatorAddr,
		types.SlashReasonDowntime,
		90,
		"evidence-001",
	)
	s.Require().NoError(err)
	firstSlashPercent := slashRecord1.SlashPercent

	// Unjail (simulate time passing)
	signingInfo, _ := s.keeper.GetValidatorSigningInfo(s.ctx, validatorAddr)
	signingInfo.JailedUntil = time.Time{}
	_ = s.keeper.SetValidatorSigningInfo(s.ctx, signingInfo)

	// Second slash (should be escalated)
	slashRecord2, err := s.keeper.SlashValidator(
		s.ctx,
		validatorAddr,
		types.SlashReasonDowntime,
		95,
		"evidence-002",
	)
	s.Require().NoError(err)
	secondSlashPercent := slashRecord2.SlashPercent

	// Second slash should be higher due to escalation
	s.Require().Greater(secondSlashPercent, firstSlashPercent)
}

// TestSlashForDoubleSigning tests double signing slash
func (s *StakingKeeperTestSuite) TestSlashForDoubleSigning() {
	validatorAddr := "validator1qperwt9wrnkg39mvs5g6ys"

	evidence := types.DoubleSignEvidence{
		EvidenceID:       "ds-001",
		ValidatorAddress: validatorAddr,
		Height1:          90,
		Height2:          90,
		VoteHash1:        "hash1",
		VoteHash2:        "hash2",
		DetectedAt:       s.ctx.BlockTime(),
		DetectedHeight:   100,
	}

	slashRecord, err := s.keeper.SlashForDoubleSigning(s.ctx, validatorAddr, 90, evidence)
	s.Require().NoError(err)
	s.Require().NotNil(slashRecord)
	s.Require().Equal(types.SlashReasonDoubleSigning, slashRecord.Reason)
	s.Require().True(slashRecord.Tombstoned) // Double signing should tombstone
}

// TestSlashForInvalidAttestation tests invalid attestation slash
func (s *StakingKeeperTestSuite) TestSlashForInvalidAttestation() {
	validatorAddr := "validator1qperwt9wrnkg39mvs5g6ys"

	attestation := types.InvalidVEIDAttestation{
		RecordID:         "ia-001",
		ValidatorAddress: validatorAddr,
		AttestationID:    "att-001",
		Reason:           "score mismatch",
		ExpectedScore:    9000,
		ActualScore:      5000,
		ScoreDifference:  4000,
		DetectedAt:       s.ctx.BlockTime(),
		DetectedHeight:   100,
	}

	slashRecord, err := s.keeper.SlashForInvalidAttestation(s.ctx, validatorAddr, attestation)
	s.Require().NoError(err)
	s.Require().NotNil(slashRecord)
	s.Require().Equal(types.SlashReasonInvalidVEIDAttestation, slashRecord.Reason)
}

// TestIterators tests iteration functions
func (s *StakingKeeperTestSuite) TestIterators() {
	// Create multiple performances
	for i := 1; i <= 3; i++ {
		perf := types.NewValidatorPerformance("validator"+string(rune('0'+i)), 1)
		perf.BlocksProposed = int64(i * 10)
		err := s.keeper.SetValidatorPerformance(s.ctx, *perf)
		s.Require().NoError(err)
	}

	// Iterate performances
	var count int
	s.keeper.WithValidatorPerformances(s.ctx, func(perf types.ValidatorPerformance) bool {
		count++
		return false
	})
	s.Require().Equal(3, count)
}

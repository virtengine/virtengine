// Package keeper implements the delegation module keeper tests.
//
// VE-922: gRPC query server tests
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
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	delegationv1 "github.com/virtengine/virtengine/sdk/go/node/delegation/v1"
	"github.com/virtengine/virtengine/x/delegation/types"
)

var (
	queryDelegatorAddr = sdk.AccAddress([]byte("delegator_query__01")).String()
	queryValidator1    = sdk.AccAddress([]byte("validator_query__01")).String()
	queryValidator2    = sdk.AccAddress([]byte("validator_query__02")).String()
	queryValidator3    = sdk.AccAddress([]byte("validator_query__03")).String()
)

// DelegationGRPCQueryTestSuite is the test suite for delegation gRPC queries.
type DelegationGRPCQueryTestSuite struct {
	suite.Suite
	ctx     sdk.Context
	keeper  Keeper
	querier *Querier
	cdc     codec.BinaryCodec
	skey    *storetypes.KVStoreKey
}

// SetupTest sets up the test suite.
func (s *DelegationGRPCQueryTestSuite) SetupTest() {
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
		nil,
		nil,
		"authority",
	)

	s.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Height: 100,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	require.NoError(s.T(), s.keeper.SetParams(s.ctx, types.DefaultParams()))
	s.querier = NewQuerier(s.keeper)
}

// TestDelegationGRPCQueryTestSuite runs the test suite.
func TestDelegationGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, new(DelegationGRPCQueryTestSuite))
}

// TestEmptyStateQueries verifies empty state responses.
func (s *DelegationGRPCQueryTestSuite) TestEmptyStateQueries() {
	resp, err := s.querier.DelegatorDelegations(s.ctx, &delegationv1.QueryDelegatorDelegationsRequest{
		DelegatorAddress: queryDelegatorAddr,
		Pagination:       &sdkquery.PageRequest{Limit: 10},
	})
	s.Require().NoError(err)
	s.Require().Len(resp.Delegations, 0)

	valResp, err := s.querier.ValidatorDelegations(s.ctx, &delegationv1.QueryValidatorDelegationsRequest{
		ValidatorAddress: queryValidator1,
		Pagination:       &sdkquery.PageRequest{Limit: 10},
	})
	s.Require().NoError(err)
	s.Require().Len(valResp.Delegations, 0)

	rewardResp, err := s.querier.DelegatorRewards(s.ctx, &delegationv1.QueryDelegatorRewardsRequest{
		DelegatorAddress: queryDelegatorAddr,
		ValidatorAddress: queryValidator1,
	})
	s.Require().NoError(err)
	s.Require().Len(rewardResp.Rewards, 0)
	s.Require().Equal("0", rewardResp.TotalReward)
}

// TestDelegationQueriesPopulated verifies populated query responses.
func (s *DelegationGRPCQueryTestSuite) TestDelegationQueriesPopulated() {
	del1 := types.NewDelegation(queryDelegatorAddr, queryValidator1, "1000000000000000000", "1000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	del2 := types.NewDelegation(queryDelegatorAddr, queryValidator2, "2000000000000000000", "2000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	require.NoError(s.T(), s.keeper.SetDelegation(s.ctx, *del1))
	require.NoError(s.T(), s.keeper.SetDelegation(s.ctx, *del2))

	resp, err := s.querier.Delegation(s.ctx, &delegationv1.QueryDelegationRequest{
		DelegatorAddress: queryDelegatorAddr,
		ValidatorAddress: queryValidator1,
	})
	s.Require().NoError(err)
	s.Require().True(resp.Found)
	s.Require().Equal(queryValidator1, resp.Delegation.ValidatorAddress)

	listResp, err := s.querier.DelegatorDelegations(s.ctx, &delegationv1.QueryDelegatorDelegationsRequest{
		DelegatorAddress: queryDelegatorAddr,
		Pagination:       &sdkquery.PageRequest{Limit: 10},
	})
	s.Require().NoError(err)
	s.Require().Len(listResp.Delegations, 2)

	valResp, err := s.querier.ValidatorDelegations(s.ctx, &delegationv1.QueryValidatorDelegationsRequest{
		ValidatorAddress: queryValidator1,
		Pagination:       &sdkquery.PageRequest{Limit: 10},
	})
	s.Require().NoError(err)
	s.Require().Len(valResp.Delegations, 1)

	reward := types.NewDelegatorReward(queryDelegatorAddr, queryValidator1, 1, "500000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	require.NoError(s.T(), s.keeper.SetDelegatorReward(s.ctx, *reward))

	rewardResp, err := s.querier.DelegatorRewards(s.ctx, &delegationv1.QueryDelegatorRewardsRequest{
		DelegatorAddress: queryDelegatorAddr,
		ValidatorAddress: queryValidator1,
	})
	s.Require().NoError(err)
	s.Require().Len(rewardResp.Rewards, 1)
	s.Require().Equal("500000", rewardResp.TotalReward)
}

// TestDelegationPagination verifies offset and key-based pagination.
func (s *DelegationGRPCQueryTestSuite) TestDelegationPagination() {
	del1 := types.NewDelegation(queryDelegatorAddr, queryValidator1, "1000000000000000000", "1000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	del2 := types.NewDelegation(queryDelegatorAddr, queryValidator2, "2000000000000000000", "2000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	del3 := types.NewDelegation(queryDelegatorAddr, queryValidator3, "3000000000000000000", "3000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	require.NoError(s.T(), s.keeper.SetDelegation(s.ctx, *del1))
	require.NoError(s.T(), s.keeper.SetDelegation(s.ctx, *del2))
	require.NoError(s.T(), s.keeper.SetDelegation(s.ctx, *del3))

	offsetResp, err := s.querier.DelegatorDelegations(s.ctx, &delegationv1.QueryDelegatorDelegationsRequest{
		DelegatorAddress: queryDelegatorAddr,
		Pagination: &sdkquery.PageRequest{
			Offset:     1,
			Limit:      1,
			CountTotal: true,
		},
	})
	s.Require().NoError(err)
	s.Require().Len(offsetResp.Delegations, 1)
	s.Require().Equal(uint64(3), offsetResp.Pagination.Total)

	firstPage, err := s.querier.DelegatorDelegations(s.ctx, &delegationv1.QueryDelegatorDelegationsRequest{
		DelegatorAddress: queryDelegatorAddr,
		Pagination: &sdkquery.PageRequest{
			Limit: 1,
		},
	})
	s.Require().NoError(err)
	s.Require().Len(firstPage.Delegations, 1)
	s.Require().NotEmpty(firstPage.Pagination.NextKey)

	secondPage, err := s.querier.DelegatorDelegations(s.ctx, &delegationv1.QueryDelegatorDelegationsRequest{
		DelegatorAddress: queryDelegatorAddr,
		Pagination: &sdkquery.PageRequest{
			Key:   firstPage.Pagination.NextKey,
			Limit: 1,
		},
	})
	s.Require().NoError(err)
	s.Require().Len(secondPage.Delegations, 1)
}

// TestQueryErrors verifies not found and invalid address handling.
func (s *DelegationGRPCQueryTestSuite) TestQueryErrors() {
	notFoundResp, err := s.querier.Delegation(s.ctx, &delegationv1.QueryDelegationRequest{
		DelegatorAddress: queryDelegatorAddr,
		ValidatorAddress: queryValidator1,
	})
	s.Require().NoError(err)
	s.Require().False(notFoundResp.Found)

	ubdResp, err := s.querier.UnbondingDelegation(s.ctx, &delegationv1.QueryUnbondingDelegationRequest{
		UnbondingId: "missing-ubd",
	})
	s.Require().NoError(err)
	s.Require().False(ubdResp.Found)

	_, err = s.querier.DelegatorDelegations(s.ctx, &delegationv1.QueryDelegatorDelegationsRequest{
		DelegatorAddress: "invalid-address",
	})
	s.Require().Error(err)

	_, err = s.querier.ValidatorShares(s.ctx, &delegationv1.QueryValidatorSharesRequest{
		ValidatorAddress: "invalid-address",
	})
	s.Require().Error(err)
}

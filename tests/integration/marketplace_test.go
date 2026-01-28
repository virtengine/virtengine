//go:build e2e.integration

// Package integration contains integration tests for VirtEngine.
// These tests verify end-to-end flows against a running localnet.
package integration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/sdk/go/cli"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	"github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	"github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	"github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	sdktestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	"github.com/virtengine/virtengine/testutil"
)

// MarketplaceIntegrationTestSuite tests marketplace-related flows.
// This suite verifies:
//   - Marketplace offering creation
//   - Order submission and matching
//   - Provider daemon bid/provision flows
//
// Acceptance Criteria (VE-002):
//   - Integration test suite can create marketplace offering + order
//   - Observe daemon bid/provision simulation
type MarketplaceIntegrationTestSuite struct {
	suite.Suite

	*testutil.NetworkTestSuite

	cctx         client.Context
	addrDeployer sdk.AccAddress
	addrProvider sdk.AccAddress
}

// TestMarketplaceIntegration runs the marketplace integration test suite.
func TestMarketplaceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(MarketplaceIntegrationTestSuite))
}

// SetupSuite runs once before all tests in the suite.
func (s *MarketplaceIntegrationTestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()

	ctx := context.Background()
	val := s.Network().Validators[0]
	kb := val.ClientCtx.Keyring

	_, _, err := kb.NewMnemonic("integration-deployer", keyring.English, sdk.FullFundraiserPath, "", hd.Secp256k1)
	s.Require().NoError(err)

	_, _, err = kb.NewMnemonic("integration-provider", keyring.English, sdk.FullFundraiserPath, "", hd.Secp256k1)
	s.Require().NoError(err)

	deployer, err := kb.Key("integration-deployer")
	s.Require().NoError(err)

	provider, err := kb.Key("integration-provider")
	s.Require().NoError(err)

	s.addrDeployer, err = deployer.GetAddress()
	s.Require().NoError(err)

	s.addrProvider, err = provider.GetAddress()
	s.Require().NoError(err)

	s.cctx = val.ClientCtx

	res, err := clitestutil.ExecSend(
		ctx,
		s.cctx,
		cli.TestFlags().
			With(
				val.Address.String(),
				s.addrDeployer.String(),
				sdk.NewCoins(sdk.NewInt64Coin(s.Config().BondDenom, 10000000)).String(),
			).
			WithFrom(val.Address.String()).
			WithGasAutoFlags().
			WithSkipConfirm().
			WithBroadcastModeBlock()...,
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())
	clitestutil.ValidateTxSuccessful(ctx, s.T(), s.cctx, res.Bytes())

	res, err = clitestutil.ExecSend(
		ctx,
		s.cctx,
		cli.TestFlags().
			With(
				val.Address.String(),
				s.addrProvider.String(),
				sdk.NewCoins(sdk.NewInt64Coin(s.Config().BondDenom, 10000000)).String(),
			).
			WithFrom(val.Address.String()).
			WithGasAutoFlags().
			WithSkipConfirm().
			WithBroadcastModeBlock()...,
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())
	clitestutil.ValidateTxSuccessful(ctx, s.T(), s.cctx, res.Bytes())

	_, err = clitestutil.TxGenerateClientExec(
		ctx,
		s.cctx,
		cli.TestFlags().WithFrom(s.addrDeployer.String())...,
	)
	s.Require().NoError(err)

	_, err = clitestutil.TxPublishClientExec(
		ctx,
		s.cctx,
		cli.TestFlags().
			WithFrom(s.addrDeployer.String()).
			WithGasAutoFlags().
			WithSkipConfirm().
			WithBroadcastModeBlock()...,
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())
}

// TestCreateMarketplaceOffering tests creating a marketplace offering.
//
// Flow:
//  1. Create provider account
//  2. Register provider on chain
//  3. Create resource offering (compute/storage specs)
//  4. Verify offering is queryable
func (s *MarketplaceIntegrationTestSuite) TestOrderBidLeaseFlow() {
	ctx := context.Background()

	deploymentPath, err := filepath.Abs("../../x/deployment/testdata/deployment.yaml")
	s.Require().NoError(err)

	providerPath, err := filepath.Abs("../../x/provider/testdata/provider.yaml")
	s.Require().NoError(err)

	// Create deployment/order
	res, err := clitestutil.ExecDeploymentCreate(
		ctx,
		s.cctx,
		deploymentPath,
		cli.TestFlags().
			WithFrom(s.addrDeployer.String()).
			WithDeposit(sdktestutil.VECoin(s.T(), 5000000)).
			WithSkipConfirm().
			WithGasAutoFlags().
			WithBroadcastModeBlock()...,
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())
	clitestutil.ValidateTxSuccessful(ctx, s.T(), s.cctx, res.Bytes())

	// Query orders
	resp, err := clitestutil.ExecQueryOrders(ctx, s.cctx.WithOutputFormat("json"))
	s.Require().NoError(err)

	orders := &v1beta5.QueryOrdersResponse{}
	err = s.cctx.Codec.UnmarshalJSON(resp.Bytes(), orders)
	s.Require().NoError(err)
	s.Require().NotEmpty(orders.Orders)
	order := orders.Orders[0].Order

	// Create provider
	res, err = clitestutil.ExecTxCreateProvider(
		ctx,
		s.cctx,
		providerPath,
		cli.TestFlags().
			WithFrom(s.addrProvider.String()).
			WithSkipConfirm().
			WithGasAutoFlags().
			WithBroadcastModeBlock()...,
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())
	clitestutil.ValidateTxSuccessful(ctx, s.T(), s.cctx, res.Bytes())

	// Simulate daemon bid by submitting bid tx directly
	res, err = clitestutil.ExecCreateBid(
		ctx,
		s.cctx,
		cli.TestFlags().
			WithFrom(s.addrProvider.String()).
			WithOrderID(order.ID).
			WithPrice(sdktestutil.VEDecCoinAmount(s.T(), "1.1")).
			WithDeposit(sdktestutil.VECoin(s.T(), 5000000)).
			WithGasAutoFlags().
			WithSkipConfirm().
			WithBroadcastModeBlock()...,
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())
	clitestutil.ValidateTxSuccessful(ctx, s.T(), s.cctx, res.Bytes())

	// Query bids to locate the provider bid
	resp, err = clitestutil.ExecQueryBids(ctx, s.cctx.WithOutputFormat("json"))
	s.Require().NoError(err)

	bids := &v1beta5.QueryBidsResponse{}
	err = s.cctx.Codec.UnmarshalJSON(resp.Bytes(), bids)
	s.Require().NoError(err)
	s.Require().NotEmpty(bids.Bids)

	// Accept bid -> create lease
	res, err = clitestutil.ExecCreateLease(
		ctx,
		s.cctx,
		cli.TestFlags().
			WithFrom(s.addrDeployer.String()).
			WithBidID(bids.Bids[0].Bid.ID).
			WithGasAutoFlags().
			WithSkipConfirm().
			WithBroadcastModeBlock()...,
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())
	clitestutil.ValidateTxSuccessful(ctx, s.T(), s.cctx, res.Bytes())

	resp, err = clitestutil.ExecQueryLeases(ctx, s.cctx.WithOutputFormat("json"))
	s.Require().NoError(err)

	leases := &v1beta5.QueryLeasesResponse{}
	err = s.cctx.Codec.UnmarshalJSON(resp.Bytes(), leases)
	s.Require().NoError(err)
	s.Require().NotEmpty(leases.Leases)
	s.Require().Equal(s.addrProvider.String(), leases.Leases[0].Lease.ID.Provider)
}


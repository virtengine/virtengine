//go:build e2e.integration

// Package integration contains integration tests for VirtEngine.
// These tests verify end-to-end flows against a running localnet.
package integration

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/sdk/go/cli"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	"github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	deposit "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	sdktestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	"github.com/virtengine/virtengine/testutil"
	"github.com/virtengine/virtengine/testutil/network"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
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

	// providerMnemonic is generated before the test network boots so the
	// provider address is known at genesis time and can carry a seeded VEID
	// score (provider registration requires VEID ≥ 70, MARKET-VEID-002).
	providerMnemonic string

	// seededProviderAddr is the address the genesis VEID record was seeded
	// for; SetupSuite asserts the recovered key matches it.
	seededProviderAddr sdk.AccAddress
}

// seedProviderVEIDScore returns an InterceptState that seeds a VEID identity
// record with the given score for providerAddr in the veid genesis.
//
// NOTE: VEID GenesisState is a plain Go struct, not a protobuf message, so
// the codec (ProtoCodec) cannot unmarshal it — MustUnmarshalJSON panics with
// `unknown field "identity_records"`. The seed is applied with encoding/json
// maps instead, following the tests/e2e/mfa_gating_test.go intercept pattern.
func seedProviderVEIDScore(providerAddr sdk.AccAddress, score uint32) network.InterceptState {
	return func(_ codec.Codec, module string, raw json.RawMessage) json.RawMessage {
		if module != veidtypes.ModuleName {
			return nil
		}
		var state map[string]any
		if err := json.Unmarshal(raw, &state); err != nil {
			return nil
		}
		genesisTime := time.Unix(1_700_000_000, 0).UTC()
		record := veidtypes.NewIdentityRecord(providerAddr.String(), genesisTime)
		record.CurrentScore = score
		record.ScoreVersion = "genesis-seed-marketplace"
		record.Tier = veidtypes.ComputeTierFromScore(score)
		record.LastVerifiedAt = &genesisTime
		record.UpdatedAt = genesisTime
		recordRaw, err := json.Marshal(record)
		if err != nil {
			return nil
		}
		var recordMap map[string]any
		if err := json.Unmarshal(recordRaw, &recordMap); err != nil {
			return nil
		}
		records, _ := state["identity_records"].([]any)
		state["identity_records"] = append(records, recordMap)
		out, err := json.Marshal(state)
		if err != nil {
			return nil
		}
		return out
	}
}

// TestMarketplaceIntegration runs the marketplace integration test suite.
func TestMarketplaceIntegration(t *testing.T) {
	s := &MarketplaceIntegrationTestSuite{}

	// Pre-generate the provider key: its address must be known before the
	// network starts so genesis can seed its VEID score.
	tmpKR := keyring.NewInMemory(sdkutil.MakeEncodingConfig().Codec)
	providerKey, mnemonic, err := tmpKR.NewMnemonic(
		"integration-provider", keyring.English, sdk.FullFundraiserPath, "", hd.Secp256k1,
	)
	require.NoError(t, err)
	providerAddr, err := providerKey.GetAddress()
	require.NoError(t, err)
	s.providerMnemonic = mnemonic
	s.seededProviderAddr = providerAddr

	cfg := network.DefaultConfig(
		testutil.NewTestNetworkFixture,
		network.WithInterceptState(seedProviderVEIDScore(providerAddr, 75)),
	)
	cfg.NumValidators = 1
	s.NetworkTestSuite = testutil.NewNetworkTestSuite(&cfg, s)
	suite.Run(t, s)
}

// SetupSuite runs once before all tests in the suite.
func (s *MarketplaceIntegrationTestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()

	ctx := context.Background()
	val := s.Network().Validators[0]
	kb := val.ClientCtx.Keyring

	_, _, err := kb.NewMnemonic("integration-deployer", keyring.English, sdk.FullFundraiserPath, "", hd.Secp256k1)
	s.Require().NoError(err)

	// Recover the pre-generated provider key so its address matches the
	// genesis-seeded VEID identity record (score 75 satisfies the ≥70
	// provider registration requirement).
	_, err = kb.NewAccount("integration-provider", s.providerMnemonic, "", sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	deployer, err := kb.Key("integration-deployer")
	s.Require().NoError(err)

	provider, err := kb.Key("integration-provider")
	s.Require().NoError(err)

	s.addrDeployer, err = deployer.GetAddress()
	s.Require().NoError(err)

	s.addrProvider, err = provider.GetAddress()
	s.Require().NoError(err)
	s.Require().Equal(
		s.seededProviderAddr.String(), s.addrProvider.String(),
		"recovered provider key must match the genesis-seeded VEID address",
	)

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

// submitProviderTx broadcasts a message signed by the provider key. It
// mirrors the CLI tx flags the suite uses elsewhere (gas auto, 0.0025uve gas
// prices, sync broadcast) via a tx factory, for messages the market CLI
// cannot express (resource offers on bids, resource heartbeats).
func (s *MarketplaceIntegrationTestSuite) submitProviderTx(msg sdk.Msg) {
	cmd := &cobra.Command{Use: "market-bid-tx"}
	flags.AddTxFlagsToCmd(cmd)
	fs := cmd.Flags()
	s.Require().NoError(fs.Set(flags.FlagFrom, s.addrProvider.String()))
	s.Require().NoError(fs.Set(flags.FlagChainID, s.Config().ChainID))
	s.Require().NoError(fs.Set(flags.FlagGas, "auto"))
	s.Require().NoError(fs.Set(flags.FlagGasPrices, "0.0025uve"))
	s.Require().NoError(fs.Set(flags.FlagGasAdjustment, "1.5"))
	s.Require().NoError(fs.Set(flags.FlagBroadcastMode, "sync"))

	// The factory snapshots the signer name at construction (used for the
	// gas-simulation signature), so the client context must carry it first.
	// FromName is the keyring key name ("integration-provider"): the keystore
	// has no address-to-key fallback.
	cctx := s.cctx.
		WithFromName("integration-provider").
		WithFromAddress(s.addrProvider).
		WithSkipConfirmation(true).
		// client.Context.BroadcastTx only supports sync/async; the caller
		// waits a block and queries, which is the commit confirmation.
		WithBroadcastMode("sync")

	txf, err := clienttx.NewFactoryCLI(cctx, fs)
	s.Require().NoError(err)

	s.Require().NoError(clienttx.GenerateOrBroadcastTxWithFactory(cctx, txf, msg))
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
		append([]string{deploymentPath},
			cli.TestFlags().
				WithFrom(s.addrDeployer.String()).
				WithDeposit(sdktestutil.VECoin(s.T(), 5000000)).
				WithSkipConfirm().
				WithGasAutoFlags().
				WithBroadcastModeBlock()...)...,
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
	order := orders.Orders[0]

	// Create provider
	res, err = clitestutil.ExecTxCreateProvider(
		ctx,
		s.cctx,
		append([]string{providerPath},
			cli.TestFlags().
				WithFrom(s.addrProvider.String()).
				WithSkipConfirm().
				WithGasAutoFlags().
				WithBroadcastModeBlock()...)...,
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())
	clitestutil.ValidateTxSuccessful(ctx, s.T(), s.cctx, res.Bytes())

	// Simulate the provider daemon inventory heartbeat: lease creation reserves
	// capacity through x/resources, which requires a fresh heartbeat from an
	// eligible (registered) provider whose available capacity covers the
	// order. The advertised capacity generously covers the test deployment
	// (0.01 CPU, 128Mi RAM, 512Mi storage, no GPU).
	heartbeatCapacity := resourcesv1.ResourceCapacity{
		CpuCores:  16,
		MemoryGb:  64,
		StorageGb: 64,
	}
	s.submitProviderTx(&resourcesv1.MsgProviderHeartbeat{
		ProviderAddress: s.addrProvider.String(),
		ResourceClass:   resourcesv1.ResourceClass_RESOURCE_CLASS_COMPUTE,
		Total:           heartbeatCapacity,
		Available:       heartbeatCapacity,
		Sequence:        1,
	})
	s.Require().NoError(s.Network().WaitForNextBlock())

	// Simulate the provider daemon bid by submitting MsgCreateBid directly with
	// a resource offer derived from the order spec — the same construction the
	// provider daemon uses (pkg/provider_daemon/chain_client.go). The market
	// CLI bid-create cannot carry resource offers, and lease creation now
	// reserves capacity through x/resources (marketResourceCapacity rejects
	// bids with no reservable capacity), so a CLI-created bid can never reach
	// lease.
	offer := v1beta5.ResourceOfferFromRU(order.Spec.Resources)
	s.Require().NotEmpty(offer, "order spec must carry resources for the bid offer")
	bidMsg := v1beta5.NewMsgCreateBid(
		mv1.MakeBidID(order.ID, s.addrProvider),
		sdktestutil.VEDecCoinAmount(s.T(), "1.1"),
		deposit.Deposit{
			Amount:  sdktestutil.VECoin(s.T(), 5000000),
			Sources: deposit.Sources{deposit.SourceGrant, deposit.SourceBalance},
		},
		offer,
	)
	s.submitProviderTx(bidMsg)
	s.Require().NoError(s.Network().WaitForNextBlock())

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

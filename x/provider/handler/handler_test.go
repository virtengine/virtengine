package handler_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktestdata "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	mtypes "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	vetypes "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
	depositv1 "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
	"github.com/virtengine/virtengine/sdk/go/testutil"

	"github.com/virtengine/virtengine/testutil/state"
	mkeeper "github.com/virtengine/virtengine/x/market/keeper"
	"github.com/virtengine/virtengine/x/provider/handler"
	"github.com/virtengine/virtengine/x/provider/keeper"
)

const (
	emailValid = "test@example.com"
)

type testSuite struct {
	t       testing.TB
	ctx     sdk.Context
	keeper  keeper.IKeeper
	mkeeper mkeeper.IKeeper
	handler baseapp.MsgServiceHandler
}

func setupTestSuite(t *testing.T) *testSuite {
	ssuite := state.SetupTestSuite(t)
	suite := &testSuite{
		t:       t,
		ctx:     ssuite.Context(),
		keeper:  ssuite.ProviderKeeper(),
		mkeeper: ssuite.MarketKeeper(),
	}

	// Pass nil for VEIDKeeper and MFAKeeper - these are optional for basic provider operations
	suite.handler = handler.NewHandler(suite.keeper, suite.mkeeper, nil, nil)

	return suite
}

func TestProviderBadMessageType(t *testing.T) {
	suite := setupTestSuite(t)

	_, err := suite.handler(suite.ctx, sdk.Msg(sdktestdata.NewTestMsg()))
	require.Error(t, err)
	require.True(t, errors.Is(err, sdkerrors.ErrUnknownRequest))
}

func TestProviderCreate(t *testing.T) {
	suite := setupTestSuite(t)

	msg := &types.MsgCreateProvider{
		Owner:   testutil.AccAddress(t).String(),
		HostURI: testutil.ProviderHostname(t),
	}

	res, err := suite.handler(suite.ctx, msg)
	require.NotNil(t, res)
	require.NoError(t, err)

	t.Run("ensure event created", func(t *testing.T) {
		ev, err := sdk.ParseTypedEvent(res.Events[0])
		require.NoError(t, err)

		require.IsType(t, &types.EventProviderCreated{}, ev)

		dev := ev.(*types.EventProviderCreated)

		require.Equal(t, msg.Owner, dev.Owner)
	})

	res, err = suite.handler(suite.ctx, msg)
	require.Nil(t, res)
	require.Error(t, err)
	require.True(t, errors.Is(err, types.ErrProviderExists))
}

func TestProviderCreateWithInfo(t *testing.T) {
	suite := setupTestSuite(t)

	msg := &types.MsgCreateProvider{
		Owner:   testutil.AccAddress(t).String(),
		HostURI: testutil.ProviderHostname(t),
		Info: types.Info{
			EMail:   emailValid,
			Website: testutil.Hostname(t),
		},
	}

	res, err := suite.handler(suite.ctx, msg)
	require.NotNil(t, res)
	require.NoError(t, err)

	t.Run("ensure event created", func(t *testing.T) {
		ev, err := sdk.ParseTypedEvent(res.Events[0])
		require.NoError(t, err)

		require.IsType(t, &types.EventProviderCreated{}, ev)

		dev := ev.(*types.EventProviderCreated)

		require.Equal(t, msg.Owner, dev.Owner)
	})

	res, err = suite.handler(suite.ctx, msg)
	require.Nil(t, res)
	require.Error(t, err)
	require.True(t, errors.Is(err, types.ErrProviderExists))
}

func TestProviderCreateWithDuplicated(t *testing.T) {
	suite := setupTestSuite(t)

	msg := &types.MsgCreateProvider{
		Owner:      testutil.AccAddress(t).String(),
		HostURI:    testutil.ProviderHostname(t),
		Attributes: testutil.Attributes(t),
	}

	msg.Attributes = append(msg.Attributes, msg.Attributes[0])

	res, err := suite.handler(suite.ctx, msg)
	require.Nil(t, res)
	require.EqualError(t, err, vetypes.ErrAttributesDuplicateKeys.Error())
}

func TestProviderUpdateWithDuplicated(t *testing.T) {
	suite := setupTestSuite(t)

	createMsg := &types.MsgCreateProvider{
		Owner:      testutil.AccAddress(t).String(),
		HostURI:    testutil.ProviderHostname(t),
		Attributes: testutil.Attributes(t),
	}

	updateMsg := &types.MsgUpdateProvider{
		Owner:      createMsg.Owner,
		HostURI:    testutil.ProviderHostname(t),
		Attributes: createMsg.Attributes,
	}

	updateMsg.Attributes = append(updateMsg.Attributes, updateMsg.Attributes[0])

	err := suite.keeper.Create(suite.ctx, types.Provider(*createMsg))
	require.NoError(t, err)

	res, err := suite.handler(suite.ctx, updateMsg)
	require.Nil(t, res)
	require.EqualError(t, err, vetypes.ErrAttributesDuplicateKeys.Error())
}

func TestProviderUpdateExisting(t *testing.T) {
	suite := setupTestSuite(t)

	addr := testutil.AccAddress(t)

	createMsg := &types.MsgCreateProvider{
		Owner:      addr.String(),
		HostURI:    testutil.ProviderHostname(t),
		Attributes: testutil.Attributes(t),
	}

	updateMsg := &types.MsgUpdateProvider{
		Owner:      addr.String(),
		HostURI:    testutil.ProviderHostname(t),
		Attributes: createMsg.Attributes,
	}

	err := suite.keeper.Create(suite.ctx, types.Provider(*createMsg))
	require.NoError(t, err)

	res, err := suite.handler(suite.ctx, updateMsg)

	t.Run("ensure event created", func(t *testing.T) {
		ev, err := sdk.ParseTypedEvent(res.Events[1])
		require.NoError(t, err)

		require.IsType(t, &types.EventProviderUpdated{}, ev)

		dev := ev.(*types.EventProviderUpdated)

		require.Equal(t, updateMsg.Owner, dev.Owner)
	})

	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestProviderUpdateNotExisting(t *testing.T) {
	suite := setupTestSuite(t)
	msg := &types.MsgUpdateProvider{
		Owner:   testutil.AccAddress(t).String(),
		HostURI: testutil.ProviderHostname(t),
	}

	res, err := suite.handler(suite.ctx, msg)
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, errors.Is(err, types.ErrProviderNotFound))
}

func TestProviderUpdateAttributes(t *testing.T) {
	suite := setupTestSuite(t)

	addr := testutil.AccAddress(t)

	createMsg := &types.MsgCreateProvider{
		Owner:      addr.String(),
		HostURI:    testutil.ProviderHostname(t),
		Attributes: testutil.Attributes(t),
	}

	updateMsg := &types.MsgUpdateProvider{
		Owner:      addr.String(),
		HostURI:    testutil.ProviderHostname(t),
		Attributes: createMsg.Attributes,
	}

	err := suite.keeper.Create(suite.ctx, types.Provider(*createMsg))
	require.NoError(t, err)

	updateMsg.Attributes = nil
	res, err := suite.handler(suite.ctx, updateMsg)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestProviderDeleteExisting(t *testing.T) {
	suite := setupTestSuite(t)

	addr := testutil.AccAddress(t)

	createMsg := &types.MsgCreateProvider{
		Owner:   addr.String(),
		HostURI: testutil.ProviderHostname(t),
	}

	deleteMsg := &types.MsgDeleteProvider{
		Owner: addr.String(),
	}

	err := suite.keeper.Create(suite.ctx, types.Provider(*createMsg))
	require.NoError(t, err)

	res, err := suite.handler(suite.ctx, deleteMsg)
	require.NotNil(t, res)
	require.NoError(t, err)

	t.Run("ensure event created", func(t *testing.T) {
		// Find the EventProviderDeleted event
		var found bool
		for _, event := range res.Events {
			ev, parseErr := sdk.ParseTypedEvent(event)
			if parseErr != nil {
				continue
			}
			if dev, ok := ev.(*types.EventProviderDeleted); ok {
				require.Equal(t, deleteMsg.Owner, dev.Owner)
				found = true
				break
			}
		}
		require.True(t, found, "EventProviderDeleted should be emitted")
	})

	// Verify provider no longer exists
	_, found := suite.keeper.Get(suite.ctx, addr)
	require.False(t, found, "provider should not exist after deletion")
}

func TestProviderDeleteNonExisting(t *testing.T) {
	suite := setupTestSuite(t)
	msg := &types.MsgDeleteProvider{
		Owner: testutil.AccAddress(t).String(),
	}

	res, err := suite.handler(suite.ctx, msg)
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, errors.Is(err, types.ErrProviderNotFound))
}

func TestProviderDeleteWithActiveLeases(t *testing.T) {
	ssuite := state.SetupTestSuite(t)

	// Prepare bank mocks for escrow operations
	ssuite.PrepareMocks(func(ts *state.TestSuite) {
		bkeeper := ts.BankKeeper()
		bkeeper.
			On("SendCoinsFromAccountToModule", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		bkeeper.
			On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		bkeeper.
			On("SendCoinsFromModuleToModule", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
	})

	suite := &testSuite{
		t:       t,
		ctx:     ssuite.Context(),
		keeper:  ssuite.ProviderKeeper(),
		mkeeper: ssuite.MarketKeeper(),
	}
	suite.handler = handler.NewHandler(suite.keeper, suite.mkeeper, nil, nil)

	// Create a provider
	providerAddr := testutil.AccAddress(t)
	createMsg := &types.MsgCreateProvider{
		Owner:   providerAddr.String(),
		HostURI: testutil.ProviderHostname(t),
	}
	err := suite.keeper.Create(suite.ctx, types.Provider(*createMsg))
	require.NoError(t, err)

	// Create an active lease for this provider
	deployerAddr := testutil.AccAddress(t)
	group := testutil.DeploymentGroup(t, testutil.DeploymentID(t), 0)

	order, err := suite.mkeeper.CreateOrder(suite.ctx, group.ID, group.GroupSpec)
	require.NoError(t, err)

	bidID := mv1.MakeBidID(order.ID, providerAddr)
	price := testutil.VEDecCoinRandom(t)
	roffer := mtypes.ResourceOfferFromRU(group.GroupSpec.Resources)

	bid, err := suite.mkeeper.CreateBid(suite.ctx, bidID, price, roffer)
	require.NoError(t, err)

	// Set up escrow account for the bid
	bidDepositMsg := &mtypes.MsgCreateBid{
		ID: bidID,
		Deposit: depositv1.Deposit{
			Amount:  mtypes.DefaultBidMinDeposit,
			Sources: depositv1.Sources{depositv1.SourceBalance},
		}}
	deposits, err := ssuite.EscrowKeeper().AuthorizeDeposits(suite.ctx, bidDepositMsg)
	require.NoError(t, err)

	err = ssuite.EscrowKeeper().AccountCreate(
		suite.ctx,
		bid.ID.ToEscrowAccountID(),
		providerAddr,
		deposits,
	)
	require.NoError(t, err)

	// Create lease and match order/bid
	err = suite.mkeeper.CreateLease(suite.ctx, bid)
	require.NoError(t, err)
	suite.mkeeper.OnBidMatched(suite.ctx, bid)
	suite.mkeeper.OnOrderMatched(suite.ctx, order)

	// Set up escrow for the deployment
	defaultDeposit, err := dtypes.DefaultParams().MinDepositFor("uve")
	require.NoError(t, err)

	deploymentDepositMsg := &dtypes.MsgCreateDeployment{
		ID: order.ID.GroupID().DeploymentID(),
		Deposit: depositv1.Deposit{
			Amount:  defaultDeposit,
			Sources: depositv1.Sources{depositv1.SourceBalance},
		}}

	dDeposits, err := ssuite.EscrowKeeper().AuthorizeDeposits(suite.ctx, deploymentDepositMsg)
	require.NoError(t, err)

	err = ssuite.EscrowKeeper().AccountCreate(
		suite.ctx,
		bid.ID.DeploymentID().ToEscrowAccountID(),
		deployerAddr,
		dDeposits,
	)
	require.NoError(t, err)

	err = ssuite.EscrowKeeper().PaymentCreate(
		suite.ctx,
		bid.ID.LeaseID().ToEscrowPaymentID(),
		providerAddr,
		bid.Price,
	)
	require.NoError(t, err)

	// Now try to delete the provider - should fail with ErrProviderHasActiveLeases
	deleteMsg := &types.MsgDeleteProvider{
		Owner: providerAddr.String(),
	}

	res, err := suite.handler(suite.ctx, deleteMsg)
	require.Nil(t, res)
	require.Error(t, err)
	require.True(t, errors.Is(err, types.ErrProviderHasActiveLeases), "expected ErrProviderHasActiveLeases, got: %v", err)

	// Verify provider still exists
	_, found := suite.keeper.Get(suite.ctx, providerAddr)
	require.True(t, found, "provider should still exist after failed deletion")
}

func TestRequestDomainVerification(t *testing.T) {
	suite := setupTestSuite(t)

	providerAddr := testutil.AccAddress(t)
	createMsg := &types.MsgCreateProvider{
		Owner:   providerAddr.String(),
		HostURI: testutil.ProviderHostname(t),
	}

	err := suite.keeper.Create(suite.ctx, types.Provider(*createMsg))
	require.NoError(t, err)

	msg := &types.MsgRequestDomainVerification{
		Owner:  providerAddr.String(),
		Domain: "provider.example.com",
		Method: types.VERIFICATION_METHOD_DNS_TXT,
	}

	res, err := suite.handler(suite.ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	t.Run("ensure event created", func(t *testing.T) {
		var found bool
		for _, event := range res.Events {
			ev, parseErr := sdk.ParseTypedEvent(event)
			if parseErr != nil {
				continue
			}
			if dev, ok := ev.(*types.EventProviderDomainVerificationRequested); ok {
				require.Equal(t, msg.Owner, dev.Owner)
				require.Equal(t, msg.Domain, dev.Domain)
				require.Equal(t, "dns_txt", dev.Method)
				require.NotEmpty(t, dev.Token)
				found = true
				break
			}
		}
		require.True(t, found, "EventProviderDomainVerificationRequested should be emitted")
	})
}

func TestRequestDomainVerification_ProviderNotFound(t *testing.T) {
	suite := setupTestSuite(t)

	msg := &types.MsgRequestDomainVerification{
		Owner:  testutil.AccAddress(t).String(),
		Domain: "provider.example.com",
		Method: types.VERIFICATION_METHOD_DNS_TXT,
	}

	res, err := suite.handler(suite.ctx, msg)
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, errors.Is(err, types.ErrProviderNotFound))
}

func TestConfirmDomainVerification(t *testing.T) {
	suite := setupTestSuite(t)

	providerAddr := testutil.AccAddress(t)
	createMsg := &types.MsgCreateProvider{
		Owner:   providerAddr.String(),
		HostURI: testutil.ProviderHostname(t),
	}

	err := suite.keeper.Create(suite.ctx, types.Provider(*createMsg))
	require.NoError(t, err)

	requestMsg := &types.MsgRequestDomainVerification{
		Owner:  providerAddr.String(),
		Domain: "provider.example.com",
		Method: types.VERIFICATION_METHOD_DNS_TXT,
	}

	_, err = suite.handler(suite.ctx, requestMsg)
	require.NoError(t, err)

	confirmMsg := &types.MsgConfirmDomainVerification{
		Owner: providerAddr.String(),
		Proof: "dns-txt-verified-proof",
	}

	res, err := suite.handler(suite.ctx, confirmMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	t.Run("ensure event created", func(t *testing.T) {
		var found bool
		for _, event := range res.Events {
			ev, parseErr := sdk.ParseTypedEvent(event)
			if parseErr != nil {
				continue
			}
			if dev, ok := ev.(*types.EventProviderDomainVerificationConfirmed); ok {
				require.Equal(t, confirmMsg.Owner, dev.Owner)
				require.Equal(t, "provider.example.com", dev.Domain)
				require.NotEmpty(t, dev.Method)
				found = true
				break
			}
		}
		require.True(t, found, "EventProviderDomainVerificationConfirmed should be emitted")
	})
}

func TestConfirmDomainVerification_NoRequest(t *testing.T) {
	suite := setupTestSuite(t)

	providerAddr := testutil.AccAddress(t)
	createMsg := &types.MsgCreateProvider{
		Owner:   providerAddr.String(),
		HostURI: testutil.ProviderHostname(t),
	}

	err := suite.keeper.Create(suite.ctx, types.Provider(*createMsg))
	require.NoError(t, err)

	confirmMsg := &types.MsgConfirmDomainVerification{
		Owner: providerAddr.String(),
		Proof: "dns-txt-verified-proof",
	}

	res, err := suite.handler(suite.ctx, confirmMsg)
	require.Error(t, err)
	require.NotNil(t, res)
	require.Contains(t, err.Error(), "not found")
}

func TestRevokeDomainVerification(t *testing.T) {
	suite := setupTestSuite(t)

	providerAddr := testutil.AccAddress(t)
	createMsg := &types.MsgCreateProvider{
		Owner:   providerAddr.String(),
		HostURI: testutil.ProviderHostname(t),
	}

	err := suite.keeper.Create(suite.ctx, types.Provider(*createMsg))
	require.NoError(t, err)

	requestMsg := &types.MsgRequestDomainVerification{
		Owner:  providerAddr.String(),
		Domain: "provider.example.com",
		Method: types.VERIFICATION_METHOD_DNS_TXT,
	}

	_, err = suite.handler(suite.ctx, requestMsg)
	require.NoError(t, err)

	confirmMsg := &types.MsgConfirmDomainVerification{
		Owner: providerAddr.String(),
		Proof: "dns-txt-verified-proof",
	}

	_, err = suite.handler(suite.ctx, confirmMsg)
	require.NoError(t, err)

	revokeMsg := &types.MsgRevokeDomainVerification{
		Owner: providerAddr.String(),
	}

	res, err := suite.handler(suite.ctx, revokeMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	t.Run("ensure event created", func(t *testing.T) {
		var found bool
		for _, event := range res.Events {
			ev, parseErr := sdk.ParseTypedEvent(event)
			if parseErr != nil {
				continue
			}
			if dev, ok := ev.(*types.EventProviderDomainVerificationRevoked); ok {
				require.Equal(t, revokeMsg.Owner, dev.Owner)
				require.Equal(t, "provider.example.com", dev.Domain)
				found = true
				break
			}
		}
		require.True(t, found, "EventProviderDomainVerificationRevoked should be emitted")
	})
}

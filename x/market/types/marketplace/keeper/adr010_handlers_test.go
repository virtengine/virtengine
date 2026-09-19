package keeper

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
	marketplace "github.com/virtengine/virtengine/x/market/types/marketplace"
)

func TestMsgCreateOrderDirect(t *testing.T) {
	provider := providerAddr(11)
	customer := providerAddr(12)
	k, ctx := setupResolutionKeeper(t, provider)
	ms := NewMsgServerImpl(k)
	goCtx := sdk.WrapSDKContext(ctx)

	offering := fixedOffering(provider, 1, 200)
	require.NoError(t, k.CreateOffering(ctx, &offering))

	res, err := ms.CreateOrder(goCtx, &marketplace.MsgCreateOrder{
		Customer:          customer,
		OfferingId:        offering.ID.String(),
		AcquisitionMode:   string(marketplace.AcquisitionModeDirect),
		RequestedQuantity: 1,
		MaxBidPrice:       500,
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.OrderId)

	stored, found := k.GetOrder(ctx, marketplace.OrderID{CustomerAddress: customer, Sequence: 1})
	require.True(t, found)
	require.Equal(t, marketplace.OrderStateOpen, stored.State)
	require.Equal(t, marketplace.AcquisitionModeDirect, stored.EffectiveAcquisitionMode())
}

func TestMsgCreateOrderSelectorBid(t *testing.T) {
	customer := providerAddr(13)
	k, ctx := setupResolutionKeeper(t)
	ms := NewMsgServerImpl(k)
	goCtx := sdk.WrapSDKContext(ctx)

	res, err := ms.CreateOrder(goCtx, &marketplace.MsgCreateOrder{
		Customer:          customer,
		AcquisitionMode:   string(marketplace.AcquisitionModeBid),
		Category:          string(marketplace.OfferingCategoryCompute),
		RequestedQuantity: 2,
		MaxBidPrice:       1000,
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.OrderId)

	stored, found := k.GetOrder(ctx, marketplace.OrderID{CustomerAddress: customer, Sequence: 1})
	require.True(t, found)
	require.True(t, stored.IsBidOrder())
	require.NotNil(t, stored.MatchingDeadline)
	require.NotNil(t, stored.Selector)
	require.Equal(t, marketplace.OfferingCategoryCompute, stored.Selector.Category)
}

func TestMsgPlaceAndWithdrawBid(t *testing.T) {
	provider := providerAddr(14)
	customer := providerAddr(15)
	k, ctx := setupResolutionKeeper(t, provider)
	ms := NewMsgServerImpl(k)
	goCtx := sdk.WrapSDKContext(ctx)

	orderRes, err := ms.CreateOrder(goCtx, &marketplace.MsgCreateOrder{
		Customer:          customer,
		AcquisitionMode:   string(marketplace.AcquisitionModeBid),
		RequestedQuantity: 1,
		MaxBidPrice:       1000,
	})
	require.NoError(t, err)

	bidRes, err := ms.PlaceBid(goCtx, &marketplace.MsgPlaceBid{
		Provider: provider,
		OrderId:  orderRes.OrderId,
		Price:    700,
	})
	require.NoError(t, err)
	require.NotEmpty(t, bidRes.BidId)

	_, err = ms.WithdrawBid(goCtx, &marketplace.MsgWithdrawBid{Provider: provider, BidId: bidRes.BidId})
	require.NoError(t, err)

	bidID, err := marketplace.ParseBidID(bidRes.BidId)
	require.NoError(t, err)
	stored, found := k.GetBid(ctx, bidID)
	require.True(t, found)
	require.Equal(t, marketplace.BidStateWithdrawn, stored.State)
}

func TestMsgRegisterWaldurSourceAuthority(t *testing.T) {
	k, ctx := setupResolutionKeeper(t)
	ms := NewMsgServerImpl(k)
	goCtx := sdk.WrapSDKContext(ctx)

	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	_, err = ms.RegisterWaldurSource(goCtx, &marketplace.MsgRegisterWaldurSource{
		Authority:  providerAddr(99),
		InstanceId: "waldur-x",
		PublicKey:  hex.EncodeToString(pub),
	})
	require.Error(t, err, "wrong authority must be rejected")

	_, err = ms.RegisterWaldurSource(goCtx, &marketplace.MsgRegisterWaldurSource{
		Authority:  k.GetAuthority(),
		InstanceId: "waldur-x",
		PublicKey:  hex.EncodeToString(pub),
	})
	require.NoError(t, err)

	_, found := k.GetWaldurSource(ctx, "waldur-x")
	require.True(t, found)
}

func TestMsgIngestWaldurOfferingEndToEnd(t *testing.T) {
	provider := providerAddr(16)
	k, ctx := setupResolutionKeeper(t, provider)
	ms := NewMsgServerImpl(k)
	goCtx := sdk.WrapSDKContext(ctx)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	_, err = ms.RegisterWaldurSource(goCtx, &marketplace.MsgRegisterWaldurSource{
		Authority:  k.GetAuthority(),
		InstanceId: "waldur-eu",
		PublicKey:  hex.EncodeToString(pub),
	})
	require.NoError(t, err)

	params := k.GetParams(ctx)
	params.WaldurIngestCustomerProviders = map[string]string{"cust-9": provider}
	require.NoError(t, k.SetParams(ctx, params))

	snapshot := &marketplacev1.WaldurOfferingSnapshot{
		Uuid:           "snap-1",
		InstanceId:     "waldur-eu",
		Name:           "Ingested",
		Type:           "VirtEngine.Compute",
		State:          "Active",
		CustomerUuid:   "cust-9",
		Shared:         true,
		Billable:       true,
		Created:        1000,
		Modified:       2000,
		SnapshotHeight: 4,
		Components: []marketplacev1.WaldurPricingComponent{
			{Name: "hourly", Type: "usage", BillingType: "usage", MeasuredUnit: "hour", Price: "1.00"},
		},
	}
	imp := snapshotToImport(snapshot)
	att := marketplace.SignWaldurOfferingAttestation("waldur-eu", imp.UUID, imp.IngestChecksum(), 4, priv)

	res, err := ms.IngestWaldurOffering(goCtx, &marketplace.MsgIngestWaldurOffering{
		Relayer:   providerAddr(17),
		Snapshot:  snapshot,
		Signature: att.Signature,
	})
	require.NoError(t, err)
	require.Equal(t, string(marketplace.IngestActionCreate), res.Action)
	require.NotEmpty(t, res.OfferingId)
}

func TestMsgSetOfferingVisibilityAndCatalogQuery(t *testing.T) {
	provider := providerAddr(18)
	k, ctx := setupResolutionKeeper(t, provider)
	ms := NewMsgServerImpl(k)
	qs := NewQueryServerImpl(k)
	goCtx := sdk.WrapSDKContext(ctx)

	offering := fixedOffering(provider, 1, 100)
	require.NoError(t, k.CreateOffering(ctx, &offering))

	_, err := ms.SetOfferingVisibility(goCtx, &marketplace.MsgSetOfferingVisibility{
		Provider:   provider,
		OfferingId: offering.ID.String(),
		Visibility: string(marketplace.OfferingVisibilityUnlisted),
	})
	require.NoError(t, err)

	// Unlisted offerings are hidden from the default catalogue but visible
	// when requested.
	catalog, err := qs.Catalog(goCtx, &marketplacev1.QueryCatalogRequest{Category: "compute"})
	require.NoError(t, err)
	require.Empty(t, catalog.Offerings)

	catalog, err = qs.Catalog(goCtx, &marketplacev1.QueryCatalogRequest{Category: "compute", IncludeUnlisted: true})
	require.NoError(t, err)
	require.Len(t, catalog.Offerings, 1)
	require.Equal(t, offering.ID.ProviderAddress, catalog.Offerings[0].Id.ProviderAddress)
	require.Equal(t, offering.ID.Sequence, catalog.Offerings[0].Id.Sequence)
}

func TestQueryWaldurCommands(t *testing.T) {
	provider := providerAddr(20)
	k, ctx := setupResolutionKeeper(t, provider)

	offering := fixedOffering(provider, 1, 100)
	offering.Source = marketplace.OfferingSourceWaldur
	offering.Waldur = &marketplace.WaldurOfferingRef{InstanceID: "waldur-eu", OfferingUUID: "uuid-1"}
	require.NoError(t, k.CreateOffering(ctx, &offering))

	order := marketplace.Order{
		ID:                marketplace.OrderID{CustomerAddress: providerAddr(21), Sequence: 1},
		OfferingID:        offering.ID,
		State:             marketplace.OrderStateOpen,
		AcquisitionMode:   marketplace.AcquisitionModeDirect,
		RequestedQuantity: 1,
		MaxBidPrice:       1000,
		CreatedAt:         time.Unix(1, 0).UTC(),
		UpdatedAt:         time.Unix(1, 0).UTC(),
	}
	seedOrder(t, k, ctx, order)
	_, err := k.ResolveOpenOrders(ctx)
	require.NoError(t, err)

	qs := NewQueryServerImpl(k)
	goCtx := sdk.WrapSDKContext(ctx)
	res, err := qs.WaldurCommands(goCtx, &marketplacev1.QueryWaldurCommandsRequest{InstanceId: "waldur-eu", PendingOnly: true})
	require.NoError(t, err)
	require.Len(t, res.Commands, 1)
	require.False(t, res.Commands[0].Acked)
}

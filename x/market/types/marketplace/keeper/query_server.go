package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
)

type queryServer struct {
	keeper IKeeper
}

var _ marketplacev1.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the marketplace QueryServer interface.
func NewQueryServerImpl(k IKeeper) marketplacev1.QueryServer {
	return queryServer{keeper: k}
}

// OfferingPrice calculates pricing for a specific offering.
func (qs queryServer) OfferingPrice(
	goCtx context.Context,
	req *marketplacev1.QueryOfferingPriceRequest,
) (*marketplacev1.QueryOfferingPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.OfferingId == "" {
		return nil, status.Error(codes.InvalidArgument, "offering_id is required")
	}

	offeringID, err := marketplacetypes.ParseOfferingID(req.OfferingId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	offering, found := qs.keeper.GetOffering(ctx, offeringID)
	if !found {
		return nil, marketplacetypes.ErrOfferingNotFound
	}

	quantity := req.Quantity
	if quantity == 0 {
		quantity = 1
	}

	quote, err := marketplacetypes.CalculateOfferingPrice(offering, req.ResourceUnits, quantity)
	if err != nil {
		return nil, marketplacetypes.ErrPricingInvalid.Wrap(err.Error())
	}

	return &marketplacev1.QueryOfferingPriceResponse{
		Total: quote.Total,
	}, nil
}

package v1

import (
	"fmt"

	cerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
)

// MakeBidID returns BidID instance with provided order details and provider
func MakeBidID(id OrderID, provider sdk.AccAddress) BidID {
	return BidID{
		Owner:    id.Owner,
		DSeq:     id.DSeq,
		GSeq:     id.GSeq,
		OSeq:     id.OSeq,
		Provider: provider.String(),
	}
}

// Equals method compares specific bid with provided bid
func (id BidID) Equals(other BidID) bool {
	return id.OrderID().Equals(other.OrderID()) &&
		id.Provider == other.Provider
}

// LeaseID method returns lease details of bid
func (id BidID) LeaseID() LeaseID {
	return LeaseID(id)
}

// OrderID method returns OrderID details with specific bid details
func (id BidID) OrderID() OrderID {
	return OrderID{
		Owner: id.Owner,
		DSeq:  id.DSeq,
		GSeq:  id.GSeq,
		OSeq:  id.OSeq,
	}
}

// String method for consistent output.
func (id BidID) String() string {
	return fmt.Sprintf("%s/%s", id.OrderID().String(), id.Provider)
}

// GroupID method returns GroupID details with specific bid details
func (id BidID) GroupID() dtypes.GroupID {
	return id.OrderID().GroupID()
}

// DeploymentID method returns deployment details with specific bid details
func (id BidID) DeploymentID() dtypes.DeploymentID {
	return id.GroupID().DeploymentID()
}

// Validate validates bid instance and returns nil
func (id BidID) Validate() error {
	if err := id.OrderID().Validate(); err != nil {
		return cerrors.Wrap(err, "BidID: Invalid OrderID")
	}
	if _, err := sdk.AccAddressFromBech32(id.Provider); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrap("BidID: Invalid Provider Address")
	}
	if id.Owner == id.Provider {
		return sdkerrors.ErrConflict.Wrap("BidID: self-bid")
	}

	// BSeq is hardcoded to 0 for now. This will be lifted
	// once multiple bids support enable
	if id.BSeq != 0 {
		return sdkerrors.ErrConflict.Wrap("BidID: invalid bseq")
	}

	return nil
}

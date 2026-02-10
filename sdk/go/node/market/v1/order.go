package v1

import (
	"fmt"

	cerrors "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
)

// MakeOrderID returns OrderID instance with provided groupID details and oseq
func MakeOrderID(id dtypes.GroupID, oseq uint32) OrderID {
	return OrderID{
		Owner: id.Owner,
		DSeq:  id.DSeq,
		GSeq:  id.GSeq,
		OSeq:  oseq,
	}
}

// GroupID method returns groupID details for specific order
func (id OrderID) GroupID() dtypes.GroupID {
	return dtypes.GroupID{
		Owner: id.Owner,
		DSeq:  id.DSeq,
		GSeq:  id.GSeq,
	}
}

// Equals method compares specific order with provided order
func (id OrderID) Equals(other OrderID) bool {
	return id.GroupID().Equals(other.GroupID()) && id.OSeq == other.OSeq
}

// Validate method for OrderID and returns nil
func (id OrderID) Validate() error {
	if err := id.GroupID().Validate(); err != nil {
		return cerrors.Wrap(err, "OrderID: Invalid GroupID")
	}
	if id.OSeq == 0 {
		return sdkerrors.ErrInvalidSequence.Wrap("OrderID: Invalid Order Sequence")
	}
	return nil
}

// String provides stringer interface to save reflected formatting.
func (id OrderID) String() string {
	return fmt.Sprintf("%s/%v", id.GroupID(), id.OSeq)
}

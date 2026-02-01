package v1

import (
	"errors"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dpath "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
)

const (
	ordersPath = "orders"
	orderPath  = "order"
	bidsPath   = "bids"
	bidPath    = "bid"
	leasesPath = "leases"
	leasePath  = "lease"
)

var (
	ErrInvalidPath = errors.New("query: invalid path")
	ErrOwnerValue  = errors.New("query: invalid owner value")
	ErrStateValue  = errors.New("query: invalid state value")
)

// LeasePath return lease path of given lease id for queries
func LeasePath(id LeaseID) string {
	return fmt.Sprintf("%s/%s/%s", leasePath, orderParts(id.OrderID()), id.Provider)
}

func orderParts(id OrderID) string {
	return fmt.Sprintf("%s/%v/%v/%v", id.Owner, id.DSeq, id.GSeq, id.OSeq)
}

// parseOrderPath returns orderID details with provided queries, and return
// error if occurred due to wrong query
func parseOrderPath(parts []string) (OrderID, error) {
	if len(parts) < 4 {
		return OrderID{}, ErrInvalidPath
	}

	did, err := dpath.ParseGroupPath(parts[0:3])
	if err != nil {
		return OrderID{}, err
	}

	oseq, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		return OrderID{}, err
	}

	return MakeOrderID(did, uint32(oseq)), nil
}

// parseBidPath returns bidID details with provided queries, and return
// error if occurred due to wrong query
func parseBidPath(parts []string) (BidID, error) {
	if len(parts) < 5 {
		return BidID{}, ErrInvalidPath
	}

	oid, err := parseOrderPath(parts[0:4])
	if err != nil {
		return BidID{}, err
	}

	provider, err := sdk.AccAddressFromBech32(parts[4])
	if err != nil {
		return BidID{}, err
	}

	return MakeBidID(oid, provider), nil
}

// ParseLeasePath returns leaseID details with provided queries, and return
// error if occurred due to wrong query
func ParseLeasePath(parts []string) (LeaseID, error) {
	bid, err := parseBidPath(parts)
	if err != nil {
		return LeaseID{}, err
	}

	return MakeLeaseID(bid), nil
}


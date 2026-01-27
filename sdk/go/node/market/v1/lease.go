package v1

import (
	"strings"

	sdkerrors "cosmossdk.io/errors"
	"gopkg.in/yaml.v3"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
)

// MakeLeaseID returns LeaseID instance with provided bid details
func MakeLeaseID(id BidID) LeaseID {
	return LeaseID(id)
}

// Equals method compares specific lease with provided lease
func (id LeaseID) Equals(other LeaseID) bool {
	return id.BidID().Equals(other.BidID())
}

// Validate calls the BidID's validator and returns any error.
func (id LeaseID) Validate() error {
	if err := id.BidID().Validate(); err != nil {
		return sdkerrors.Wrap(err, "LeaseID: Invalid BidID")
	}
	return nil
}

// BidID method returns BidID details with specific LeaseID
func (id LeaseID) BidID() BidID {
	return BidID(id)
}

// OrderID method returns OrderID details with specific lease details
func (id LeaseID) OrderID() OrderID {
	return id.BidID().OrderID()
}

// GroupID method returns GroupID details with specific lease details
func (id LeaseID) GroupID() dtypes.GroupID {
	return id.OrderID().GroupID()
}

// DeploymentID method returns deployment details with specific lease details
func (id LeaseID) DeploymentID() dtypes.DeploymentID {
	return id.GroupID().DeploymentID()
}

// String method provides human-readable representation of LeaseID.
func (id LeaseID) String() string {
	return id.BidID().String()
}

// String implements the Stringer interface for a Lease object.
func (obj Lease) String() string {
	out, _ := yaml.Marshal(obj)
	return string(out)
}

// Leases is a collection of Lease
type Leases []Lease

// String implements the Stringer interface for a Leases object.
func (l Leases) String() string {
	var out string
	for _, order := range l {
		out += order.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// Filters returns whether lease filters valid or not
func (obj Lease) Filters(filters LeaseFilters, stateVal Lease_State) bool {
	// Checking owner filter
	if filters.Owner != "" && filters.Owner != obj.ID.Owner {
		return false
	}

	// Checking dseq filter
	if filters.DSeq != 0 && filters.DSeq != obj.ID.DSeq {
		return false
	}

	// Checking gseq filter
	if filters.GSeq != 0 && filters.GSeq != obj.ID.GSeq {
		return false
	}

	// Checking oseq filter
	if filters.OSeq != 0 && filters.OSeq != obj.ID.OSeq {
		return false
	}

	// Checking provider filter
	if filters.Provider != "" && filters.Provider != obj.ID.Provider {
		return false
	}

	// Checking state filter
	if stateVal != 0 && stateVal != obj.State {
		return false
	}

	return true
}

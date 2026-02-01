package v1beta3

import (
	types "github.com/virtengine/virtengine/sdk/go/node/types/v1beta3"
)

type attributesMatching map[string]types.Attributes

const (
	// ManifestVersionLength is the length of manifest version
	ManifestVersionLength = 32

	// DefaultOrderBiddingDuration is the default time limit for an Order being active.
	// After the duration, the Order is automatically closed.
	// ( 24(hr) * 3600(seconds per hour) ) / 7s-Block
	DefaultOrderBiddingDuration = int64(12342)

	// MaxBiddingDuration is roughly 30 days of block height
	MaxBiddingDuration = DefaultOrderBiddingDuration * int64(30)
)

// ID method returns DeploymentID details of specific deployment
func (obj Deployment) ID() DeploymentID {
	return obj.DeploymentID
}

// MatchAttributes method compares provided attributes with specific group attributes
func (g *GroupSpec) MatchAttributes(attr types.Attributes) bool {
	return types.AttributesSubsetOf(g.Requirements.Attributes, attr)
}

// ID method returns GroupID details of specific group
func (g Group) ID() GroupID {
	return g.GroupID
}

// ValidateClosable provides error response if group is already closed,
// and thus should not be closed again, else nil.
func (g Group) ValidateClosable() error {
	switch g.State {
	case GroupClosed:
		return ErrGroupClosed
	default:
		return nil
	}
}

// ValidatePausable provides error response if group is not pausable
func (g Group) ValidatePausable() error {
	switch g.State {
	case GroupClosed:
		return ErrGroupClosed
	case GroupPaused:
		return ErrGroupPaused
	default:
		return nil
	}
}

// ValidatePausable provides error response if group is not pausable
func (g Group) ValidateStartable() error {
	switch g.State {
	case GroupClosed:
		return ErrGroupClosed
	case GroupOpen:
		return ErrGroupOpen
	default:
		return nil
	}
}

// GetName method returns group name
func (g Group) GetName() string {
	return g.GroupSpec.Name
}

// GetResourceUnits method returns resources list in group
func (g Group) GetResourceUnits() ResourceUnits {
	return g.GroupSpec.Resources
}

// Accept returns whether deployment filters valid or not
func (filters DeploymentFilters) Accept(obj Deployment, stateVal Deployment_State) bool {
	// Checking owner filter
	if filters.Owner != "" && filters.Owner != obj.DeploymentID.Owner {
		return false
	}

	// Checking dseq filter
	if filters.DSeq != 0 && filters.DSeq != obj.DeploymentID.DSeq {
		return false
	}

	// Checking state filter
	if stateVal != 0 && stateVal != obj.State {
		return false
	}

	return true
}


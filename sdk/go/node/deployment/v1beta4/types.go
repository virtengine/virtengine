package v1beta4

import (
	"bytes"

	v1 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	attr "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
)

type attributesMatching map[string]attr.Attributes

const (
	GatewayVersion = "v1beta4"

	// ManifestHashLength is the length of manifest hash
	ManifestHashLength = 32

	// DefaultOrderBiddingDuration is the default time limit for an Order being active.
	// After the duration, the Order is automatically closed.
	// ( 24(hr) * 3600(seconds per hour) ) / 7s-Block
	DefaultOrderBiddingDuration = int64(12342)

	// MaxBiddingDuration is roughly 30 days of block height
	MaxBiddingDuration = DefaultOrderBiddingDuration * int64(30)
)

// MatchAttributes method compares provided attributes with specific group attributes
func (g *GroupSpec) MatchAttributes(at attr.Attributes) bool {
	return attr.AttributesSubsetOf(g.Requirements.Attributes, at)
}

// ValidateClosable provides error response if group is already closed,
// and thus should not be closed again, else nil.
func (g Group) ValidateClosable() error {
	switch g.State {
	case GroupClosed:
		return v1.ErrGroupClosed
	default:
		return nil
	}
}

// ValidatePausable provides error response if group is not pausable
func (g Group) ValidatePausable() error {
	switch g.State {
	case GroupClosed:
		return v1.ErrGroupClosed
	case GroupPaused:
		return v1.ErrGroupPaused
	default:
		return nil
	}
}

// ValidatePausable provides error response if group is not pausable
func (g Group) ValidateStartable() error {
	switch g.State {
	case GroupClosed:
		return v1.ErrGroupClosed
	case GroupOpen:
		return v1.ErrGroupOpen
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

// DeploymentResponses is a collection of DeploymentResponse
type DeploymentResponses []QueryDeploymentResponse

func (ds DeploymentResponses) String() string {
	var buf bytes.Buffer

	const sep = "\n\n"

	for _, d := range ds {
		buf.WriteString(d.String())
		buf.WriteString(sep)
	}

	if len(ds) > 0 {
		buf.Truncate(buf.Len() - len(sep))
	}

	return buf.String()
}

// Accept returns whether deployment filters valid or not
func (filters DeploymentFilters) Accept(obj v1.Deployment, stateVal v1.Deployment_State) bool {
	// Checking owner filter
	if filters.Owner != "" && filters.Owner != obj.ID.Owner {
		return false
	}

	// Checking dseq filter
	if filters.DSeq != 0 && filters.DSeq != obj.ID.DSeq {
		return false
	}

	// Checking state filter
	if stateVal != 0 && stateVal != obj.State {
		return false
	}

	return true
}

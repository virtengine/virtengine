// do not edit

package v1beta3

import (
	"fmt"

	atypes "github.com/virtengine/virtengine/sdk/go/node/audit/v1beta3"
	sdk "github.com/virtengine/virtengine/sdk/go/node/types/sdk"
	types "github.com/virtengine/virtengine/sdk/go/node/types/v1beta3"
)

type ResourceGroup interface {
	GetName() string
	GetResourceUnits() ResourceUnits
}

var _ ResourceGroup = (*GroupSpec)(nil)

type GroupSpecs []*GroupSpec

func (gspecs GroupSpecs) Dup() GroupSpecs {
	res := make(GroupSpecs, 0, len(gspecs))

	for _, gspec := range gspecs {
		gs := gspec.Dup()
		res = append(res, &gs)
	}
	return res
}

func (g GroupSpec) Dup() GroupSpec {
	res := GroupSpec{
		Name:         g.Name,
		Requirements: g.Requirements.Dup(),
		Resources:    g.Resources.Dup(),
	}

	return res
}

// ValidateBasic asserts non-zero values
func (g GroupSpec) ValidateBasic() error {
	return g.validate()
}

// GetResourceUnits method returns resources list in group
func (g GroupSpec) GetResourceUnits() ResourceUnits {
	resources := make(ResourceUnits, 0, len(g.Resources))

	for _, r := range g.Resources {
		resources = append(resources, r)
	}

	return resources
}

// GetName method returns group name
func (g GroupSpec) GetName() string {
	return g.Name
}

// Price method returns price of group
func (g GroupSpec) Price() sdk.DecCoin {
	var price sdk.DecCoin
	for idx, resource := range g.Resources {
		if idx == 0 {
			price = resource.FullPrice()
			continue
		}
		price = price.Add(resource.FullPrice())
	}
	return price
}

// MatchResourcesRequirements check if resources attributes match provider's capabilities
func (g GroupSpec) MatchResourcesRequirements(pattr types.Attributes) bool {
	for _, rgroup := range g.GetResourceUnits() {
		pgroup := pattr.GetCapabilitiesGroup("storage")
		for _, storage := range rgroup.Storage {
			if len(storage.Attributes) == 0 {
				continue
			}

			if !storage.Attributes.IN(pgroup) {
				return false
			}
		}
		if gpu := rgroup.GPU; gpu.Units.Val.Uint64() > 0 {
			attr := gpu.Attributes
			if len(attr) == 0 {
				continue
			}

			pgroup = pattr.GetCapabilitiesMap("gpu")

			if !gpu.Attributes.AnyIN(pgroup) {
				return false
			}
		}
	}

	return true
}

// MatchRequirements method compares provided attributes with specific group attributes.
// Argument provider is a bit cumbersome. First element is attributes from x/provider store
// in case tenant does not need signed attributes at all
// rest of elements (if any) are attributes signed by various auditors
func (g GroupSpec) MatchRequirements(provider []atypes.Provider) bool {
	if (len(g.Requirements.SignedBy.AnyOf) != 0) || (len(g.Requirements.SignedBy.AllOf) != 0) {
		// we cannot match if there is no signed attributes
		if len(provider) < 2 {
			return false
		}

		existingRequirements := make(attributesMatching)

		for _, existing := range provider[1:] {
			existingRequirements[existing.Auditor] = existing.Attributes
		}

		if len(g.Requirements.SignedBy.AllOf) != 0 {
			for _, validator := range g.Requirements.SignedBy.AllOf {
				// if at least one signature does not exist or no match on attributes - requirements cannot match
				if existingAttr, exists := existingRequirements[validator]; !exists ||
					!types.AttributesSubsetOf(g.Requirements.Attributes, existingAttr) {
					return false
				}
			}
		}

		if len(g.Requirements.SignedBy.AnyOf) != 0 {
			for _, validator := range g.Requirements.SignedBy.AnyOf {
				if existingAttr, exists := existingRequirements[validator]; exists &&
					types.AttributesSubsetOf(g.Requirements.Attributes, existingAttr) {
					return true
				}
			}

			return false
		}

		return true
	}

	return types.AttributesSubsetOf(g.Requirements.Attributes, provider[0].Attributes)
}

// validate does validation for provided deployment group
func (g *GroupSpec) validate() error {
	if g.Name == "" {
		return fmt.Errorf("empty group spec name denomination")
	}

	if err := g.GetResourceUnits().Validate(); err != nil {
		return err
	}

	if err := g.validatePricing(); err != nil {
		return err
	}

	return nil
}

func (g *GroupSpec) validatePricing() error {
	var price sdk.DecCoin

	mem := sdk.NewInt(0)

	for idx, resource := range g.Resources {
		if err := resource.validatePricing(); err != nil {
			return fmt.Errorf("group %v: %w", g.GetName(), err)
		}

		// all must be same denomination
		if idx == 0 {
			price = resource.FullPrice()
		} else {
			rprice := resource.FullPrice()
			if rprice.Denom != price.Denom {
				return fmt.Errorf("multi-denonimation group: (%v == %v fails)", rprice.Denom, price.Denom)
			}
			price = price.Add(rprice)
		}

		memCount := sdk.NewInt(0)
		if u := resource.Memory; u != nil {
			memCount.Add(sdk.NewIntFromUint64(u.Quantity.Value()))
		}

		mem = mem.Add(memCount.Mul(sdk.NewIntFromUint64(uint64(resource.Count))))
	}

	return nil
}

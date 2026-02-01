package v2beta3

import (
	"fmt"
	"sort"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
)

var _ dtypes.ResourceGroup = (*Group)(nil)

// GetName returns the name of group
func (g Group) GetName() string {
	return g.Name
}

func (g Group) GetResourceUnits() dtypes.ResourceUnits {
	groups := make(map[uint32]*dtypes.ResourceUnit)

	for _, svc := range g.Services {
		if _, exists := groups[svc.Resources.ID]; !exists {
			groups[svc.Resources.ID] = &dtypes.ResourceUnit{
				Resources: svc.Resources,
				Count:     svc.Count,
			}
		} else {
			groups[svc.Resources.ID].Count += svc.Count
		}
	}

	units := make(dtypes.ResourceUnits, 0, len(groups))

	for i := range groups {
		units = append(units, *groups[i])
	}

	return units
}

func (g Group) AllHostnames() []string {
	allHostnames := make([]string, 0)
	for _, service := range g.Services {
		for _, expose := range service.Expose {
			allHostnames = append(allHostnames, expose.Hosts...)
		}
	}

	return allHostnames
}

func (g *Group) Validate(helper *validateManifestGroupsHelper) error {
	if len(g.Services) == 0 {
		return fmt.Errorf("%w: group %q contains no services", ErrInvalidManifest, g.GetName())
	}

	if !sort.IsSorted(g.Services) {
		return fmt.Errorf("%w: group %q services is not sorted", ErrInvalidManifest, g.GetName())
	}

	for _, s := range g.Services {
		if err := s.validate(helper); err != nil {
			return err
		}
	}

	return nil
}

// checkAgainstGSpec check if manifest group is within GroupSpec resources
// NOTE: it modifies caller's gspec
func (g *Group) checkAgainstGSpec(gspec *groupSpec) error {
	for _, svc := range g.Services {
		if err := svc.checkAgainstGSpec(gspec); err != nil {
			return fmt.Errorf("%w: group %q: %w", ErrManifestCrossValidation, g.Name, err)
		}
	}

	return nil
}


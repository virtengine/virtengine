package v2beta3

import (
	"fmt"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
)

type Groups []Group

func (groups Groups) Validate() error {
	helper := validateManifestGroupsHelper{
		hostnames: make(map[string]int),
	}

	names := make(map[string]int) // used as a set

	for _, group := range groups {
		if err := group.Validate(&helper); err != nil {
			return err
		}
		if _, exists := names[group.GetName()]; exists {
			return fmt.Errorf("%w: duplicate group %q", ErrInvalidManifest, group.GetName())
		}

		names[group.GetName()] = 0 // Value stored is not used
	}

	if helper.globalServiceCount == 0 {
		return fmt.Errorf("%w: zero global services", ErrInvalidManifest)
	}

	return nil
}

func (groups Groups) CheckAgainstGSpecs(gspecs dtypes.GroupSpecs) error {
	gspecs = gspecs.Dup()

	if err := groups.Validate(); err != nil {
		return err
	}

	if len(groups) != len(gspecs) {
		return fmt.Errorf("invalid manifest: group count mismatch (%v != %v)", len(groups), len(gspecs))
	}

	dgroupByName := newGroupSpecsHelper(gspecs)

	for _, mgroup := range groups {
		dgroup, dgroupExists := dgroupByName[mgroup.GetName()]

		if !dgroupExists {
			return fmt.Errorf("invalid manifest: unknown deployment group ('%v')", mgroup.GetName())
		}

		if err := mgroup.checkAgainstGSpec(dgroup); err != nil {
			return err
		}
	}

	for _, gspec := range dgroupByName {
		for resID, eps := range gspec.endpoints {
			if eps.httpEndpoints > 0 {
				return fmt.Errorf("%w: group %q: resources ID (%d): under-utilized (%d) HTTP endpoints",
					ErrManifestCrossValidation, gspec.gs.Name, resID, eps.httpEndpoints)
			}

			if eps.portEndpoints > 0 {
				return fmt.Errorf("%w: group %q: resources ID (%d): under-utilized (%d) PORT endpoints",
					ErrManifestCrossValidation, gspec.gs.Name, resID, eps.portEndpoints)
			}

			if eps.ipEndpoints > 0 {
				return fmt.Errorf("%w: group %q: resources ID (%d): under-utilized (%d) IP endpoints",
					ErrManifestCrossValidation, gspec.gs.Name, resID, eps.ipEndpoints)
			}
		}

		for _, gRes := range gspec.gs.Resources {
			if gRes.Count > 0 {
				return fmt.Errorf("%w: group %q: resources ID (%d): under-utilized (%d) resources",
					ErrManifestCrossValidation, gspec.gs.GetName(), gRes.ID, gRes.Count)
			}
		}
	}

	return nil
}

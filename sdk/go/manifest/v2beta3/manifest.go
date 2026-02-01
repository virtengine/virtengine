package v2beta3

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
)

var (
	serviceNameValidationRegex = regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)
	hostnameMaxLen             = 255
)

// Manifest store list of groups
type Manifest Groups

// GetGroups returns a manifest with groups list
func (m Manifest) GetGroups() Groups {
	return Groups(m)
}

// Validate does validation for manifest
func (m Manifest) Validate() error {
	if len(m) == 0 {
		return fmt.Errorf("%w: manifest is empty", ErrInvalidManifest)
	}

	return m.GetGroups().Validate()
}

func (m Manifest) CheckAgainstDeployment(dgroups []dtypes.Group) error {
	gspecs := make([]dtypes.GroupSpec, 0, len(dgroups))

	for _, dgroup := range dgroups {
		gspec := dgroup.GroupSpec
		gspecs = append(gspecs, gspec)
	}

	return m.CheckAgainstGSpecs(gspecs)
}

func (m Manifest) CheckAgainstGSpecs(gspecs dtypes.GroupSpecs) error {
	return m.GetGroups().CheckAgainstGSpecs(gspecs)
}

// Version calculates the identifying deterministic hash for an SDL.
// Sha256 returns 32 byte sum of the SDL.
func (m Manifest) Version() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	sortedBytes, err := sdk.SortJSON(data)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(sortedBytes)

	return sum[:], nil
}


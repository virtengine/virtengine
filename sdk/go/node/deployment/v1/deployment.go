package v1

import (
	"fmt"
	"strconv"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dsdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	HashLength = 32
)

// Equals method compares specific deployment with provided deployment
func (id DeploymentID) Equals(other DeploymentID) bool {
	return id.Owner == other.Owner && id.DSeq == other.DSeq
}

// Validate method for DeploymentID and returns nil
func (id DeploymentID) Validate() error {
	_, err := sdk.AccAddressFromBech32(id.Owner)
	switch {
	case err != nil:
		return sdkerrors.Wrap(dsdkerrors.ErrInvalidAddress, "DeploymentID: Invalid Owner Address")
	case id.DSeq == 0:
		return sdkerrors.Wrap(dsdkerrors.ErrInvalidSequence, "DeploymentID: Invalid Deployment Sequence")
	}
	return nil
}

// String method for deployment IDs
func (id DeploymentID) String() string {
	return fmt.Sprintf("%s/%d", id.Owner, id.DSeq)
}

func (id DeploymentID) GetOwnerAddress() (sdk.Address, error) {
	return sdk.AccAddressFromBech32(id.Owner)
}

func ParseDeploymentID(val string) (DeploymentID, error) {
	parts := strings.Split(val, "/")
	return ParseDeploymentPath(parts)
}

// ParseDeploymentPath returns DeploymentID details with provided queries, and return
// error if occurred due to wrong query
func ParseDeploymentPath(parts []string) (DeploymentID, error) {
	if len(parts) != 2 {
		return DeploymentID{}, ErrInvalidIDPath
	}

	owner, err := sdk.AccAddressFromBech32(parts[0])
	if err != nil {
		return DeploymentID{}, err
	}

	dseq, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return DeploymentID{}, err
	}

	return DeploymentID{
		Owner: owner.String(),
		DSeq:  dseq,
	}, nil
}

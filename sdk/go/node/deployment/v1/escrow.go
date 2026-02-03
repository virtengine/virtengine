package v1

import (
	ev1 "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
)

func (id DeploymentID) ToEscrowAccountID() ev1.Account {
	return ev1.Account{
		Scope: ev1.ScopeDeployment,
		XID:   id.String(),
	}
}

func DeploymentIDFromEscrowID(id ev1.Account) (DeploymentID, error) {
	if id.Scope != ev1.ScopeDeployment {
		return DeploymentID{}, ErrInvalidEscrowID
	}

	did, err := ParseDeploymentID(id.XID)
	if err != nil {
		return DeploymentID{}, err
	}

	return did, nil
}

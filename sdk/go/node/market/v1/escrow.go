package v1

import (
	"fmt"
	"strings"

	cerrors "cosmossdk.io/errors"

	ev1 "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
)

func (id BidID) ToEscrowAccountID() ev1.Account {
	return ev1.Account{
		Scope: ev1.ScopeBid,
		XID:   id.String(),
	}
}

func (id LeaseID) ToEscrowPaymentID() ev1.Payment {
	return ev1.Payment{
		AID: id.DeploymentID().ToEscrowAccountID(),
		XID: fmt.Sprintf("%d/%d/%s", id.GSeq, id.OSeq, id.Provider),
	}
}

func LeaseIDFromPaymentID(id ev1.Payment) (LeaseID, error) {
	if id.AID.Scope != ev1.ScopeDeployment {
		return LeaseID{}, ErrInvalidEscrowID
	}

	xid := strings.Join([]string{id.AID.XID, id.XID}, "/")
	parts := strings.Split(xid, "/")
	if len(parts) != 5 {
		return LeaseID{}, cerrors.Wrapf(ErrInvalidEscrowID, "invalid payment id %s", xid)
	}

	return ParseLeasePath(parts)
}


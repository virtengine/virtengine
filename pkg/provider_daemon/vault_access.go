package provider_daemon

import (
	"context"

	"github.com/virtengine/virtengine/pkg/data_vault"
)

// ChainOrgResolver resolves organization membership using the portal chain query.
type ChainOrgResolver struct {
	ChainQuery ChainQuery
}

// IsMember returns true if the address is a member of the org.
func (c ChainOrgResolver) IsMember(ctx context.Context, orgID, address string) (bool, error) {
	if c.ChainQuery == nil || orgID == "" || address == "" {
		return false, nil
	}
	members, err := c.ChainQuery.ListOrganizationMembers(ctx, orgID)
	if err != nil {
		return false, err
	}
	for _, member := range members {
		if member.Address == address {
			return true, nil
		}
	}
	return false, nil
}

var _ data_vault.OrgResolver = ChainOrgResolver{}

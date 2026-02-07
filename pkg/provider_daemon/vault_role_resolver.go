package provider_daemon

import (
	"context"

	"github.com/virtengine/virtengine/pkg/data_vault"
	rolesv1 "github.com/virtengine/virtengine/sdk/go/node/roles/v1"
)

// GRPCRoleResolver resolves roles via the chain gRPC roles query service.
type GRPCRoleResolver struct {
	client rolesv1.QueryClient
}

// NewGRPCRoleResolver returns a role resolver backed by the roles query client.
func NewGRPCRoleResolver(client rolesv1.QueryClient) *GRPCRoleResolver {
	if client == nil {
		return nil
	}
	return &GRPCRoleResolver{client: client}
}

// HasRole checks if the address has the specified role on chain.
func (r *GRPCRoleResolver) HasRole(ctx context.Context, address, role string) (bool, error) {
	if r == nil || r.client == nil || address == "" || role == "" {
		return false, nil
	}
	resp, err := r.client.HasRole(ctx, &rolesv1.QueryHasRoleRequest{
		Address: address,
		Role:    role,
	})
	if err != nil {
		return false, err
	}
	return resp.GetHasRole(), nil
}

var _ data_vault.RoleResolver = (*GRPCRoleResolver)(nil)

package provider_daemon

import (
	"context"
	"time"

	rolesv1 "github.com/virtengine/virtengine/sdk/go/node/roles/v1"
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// GRPCPortalChainQuery resolves role and consent checks via gRPC while falling back to noop for others.
type GRPCPortalChainQuery struct {
	NoopChainQuery
	rolesClient rolesv1.QueryClient
	veidClient  veidv1.QueryClient
}

// NewGRPCPortalChainQuery creates a chain query backed by roles + veid gRPC clients.
func NewGRPCPortalChainQuery(rolesClient rolesv1.QueryClient, veidClient veidv1.QueryClient) *GRPCPortalChainQuery {
	if rolesClient == nil && veidClient == nil {
		return nil
	}
	return &GRPCPortalChainQuery{
		rolesClient: rolesClient,
		veidClient:  veidClient,
	}
}

// HasRole checks if the address has the specified chain role.
func (q *GRPCPortalChainQuery) HasRole(ctx context.Context, address, role string) (bool, error) {
	if q == nil || q.rolesClient == nil || address == "" || role == "" {
		return false, nil
	}
	resp, err := q.rolesClient.HasRole(ctx, &rolesv1.QueryHasRoleRequest{Address: address, Role: role})
	if err != nil {
		return false, err
	}
	return resp.GetHasRole(), nil
}

// HasConsent checks if the address has valid consent for the scope.
func (q *GRPCPortalChainQuery) HasConsent(ctx context.Context, address, scopeID string) (bool, error) {
	if q == nil || q.veidClient == nil || address == "" || scopeID == "" {
		return false, nil
	}
	resp, err := q.veidClient.ConsentSettings(ctx, &veidv1.QueryConsentSettingsRequest{
		AccountAddress: address,
		ScopeId:        scopeID,
	})
	if err != nil {
		return false, err
	}
	for _, info := range resp.ScopeConsents {
		if info.ScopeId != scopeID {
			continue
		}
		if !info.Granted || !info.IsActive {
			return false, nil
		}
		if info.ExpiresAt > 0 {
			expiry := time.Unix(info.ExpiresAt, 0)
			if time.Now().After(expiry) {
				return false, nil
			}
		}
		return true, nil
	}
	return false, nil
}

var _ ChainQuery = (*GRPCPortalChainQuery)(nil)

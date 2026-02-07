package data_vault

import (
	"context"
	"fmt"
)

// AccessAction describes the type of access being requested.
type AccessAction string

const (
	AccessActionUpload   AccessAction = "upload"
	AccessActionRead     AccessAction = "read"
	AccessActionMetadata AccessAction = "metadata"
	AccessActionDelete   AccessAction = "delete"
	AccessActionAudit    AccessAction = "audit"
)

// Role names mirrored from x/roles/types.Role.String().
const (
	RoleAdministrator   = "administrator"
	RoleModerator       = "moderator"
	RoleValidator       = "validator"
	RoleServiceProvider = "service_provider"
	RoleCustomer        = "customer"
	RoleSupportAgent    = "support_agent"
)

// AccessRequest contains authorization inputs.
type AccessRequest struct {
	Action AccessAction

	// Requester is the wallet address requesting access.
	Requester string

	// Owner is the wallet address that owns the blob.
	Owner string

	// Scope identifies the blob scope.
	Scope Scope

	// OrgID is the org ID asserted by the requester.
	OrgID string

	// ResourceOrgID is the org ID stored on the blob.
	ResourceOrgID string
}

// AccessControl authorizes access to vault resources.
type AccessControl interface {
	Authorize(ctx context.Context, req AccessRequest) error
}

// RoleResolver checks if an address has a role.
type RoleResolver interface {
	HasRole(ctx context.Context, address, role string) (bool, error)
}

// OrgResolver checks if an address belongs to an organization.
type OrgResolver interface {
	IsMember(ctx context.Context, orgID, address string) (bool, error)
}

// OrgRoleResolver provides org member role information.
type OrgRoleResolver interface {
	MemberRole(ctx context.Context, orgID, address string) (string, bool, error)
}

// ScopePolicy describes who can access data in a scope.
type ScopePolicy struct {
	AllowOwner      bool
	AllowOrgMembers bool
	AllowedOrgRoles []string
	AllowedRoles    []string
}

// AccessPolicy maps scopes to policies.
type AccessPolicy struct {
	ScopePolicies map[Scope]ScopePolicy
}

// DefaultAccessPolicy returns a default access policy.
func DefaultAccessPolicy() AccessPolicy {
	return AccessPolicy{
		ScopePolicies: map[Scope]ScopePolicy{
			ScopeVEID: {
				AllowOwner:      true,
				AllowOrgMembers: true,
				AllowedOrgRoles: []string{"admin"},
				AllowedRoles:    []string{RoleAdministrator, RoleSupportAgent},
			},
			ScopeSupport: {
				AllowOwner:      true,
				AllowOrgMembers: true,
				AllowedOrgRoles: []string{"admin", "member"},
				AllowedRoles:    []string{RoleAdministrator, RoleSupportAgent},
			},
			ScopeMarket: {
				AllowOwner:      true,
				AllowOrgMembers: true,
				AllowedOrgRoles: []string{"admin", "member"},
				AllowedRoles:    []string{RoleAdministrator, RoleServiceProvider},
			},
			ScopeAudit: {
				AllowOwner:      false,
				AllowOrgMembers: true,
				AllowedOrgRoles: []string{"admin"},
				AllowedRoles:    []string{RoleAdministrator, RoleModerator},
			},
		},
	}
}

// PolicyAccessControl implements AccessControl using org + role checks.
type PolicyAccessControl struct {
	policy      AccessPolicy
	roleChecker RoleResolver
	orgChecker  OrgResolver
}

// NewPolicyAccessControl creates a policy-based access controller.
func NewPolicyAccessControl(policy AccessPolicy, roleChecker RoleResolver, orgChecker OrgResolver) *PolicyAccessControl {
	if len(policy.ScopePolicies) == 0 {
		policy = DefaultAccessPolicy()
	}
	return &PolicyAccessControl{
		policy:      policy,
		roleChecker: roleChecker,
		orgChecker:  orgChecker,
	}
}

// Authorize enforces access policy for a request.
func (p *PolicyAccessControl) Authorize(ctx context.Context, req AccessRequest) error {
	if req.Requester == "" {
		return NewVaultError("Authorize", ErrUnauthorized, "requester required")
	}
	if req.Scope == "" {
		return NewVaultError("Authorize", ErrInvalidScope, "scope required")
	}

	policy, ok := p.policy.ScopePolicies[req.Scope]
	if !ok {
		return NewVaultError("Authorize", ErrInvalidScope, fmt.Sprintf("unknown scope %s", req.Scope))
	}

	if policy.AllowOwner && req.Owner != "" && req.Owner == req.Requester {
		return nil
	}

	resourceOrg := req.ResourceOrgID
	if resourceOrg == "" {
		resourceOrg = req.OrgID
	}

	if resourceOrg == "" {
		return NewVaultError("Authorize", ErrUnauthorized, "org required")
	}
	if req.OrgID != "" && req.OrgID != resourceOrg {
		return NewVaultError("Authorize", ErrUnauthorized, "org mismatch")
	}

	if p.orgChecker == nil {
		return NewVaultError("Authorize", ErrUnauthorized, "org membership resolver unavailable")
	}
	isMember, err := p.orgChecker.IsMember(ctx, resourceOrg, req.Requester)
	if err != nil {
		return NewVaultError("Authorize", ErrUnauthorized, err.Error())
	}
	if !isMember {
		return NewVaultError("Authorize", ErrUnauthorized, "org membership required")
	}

	if policy.AllowOrgMembers && len(policy.AllowedRoles) == 0 && len(policy.AllowedOrgRoles) == 0 {
		return nil
	}

	if len(policy.AllowedOrgRoles) > 0 {
		orgRoleResolver, ok := p.orgChecker.(OrgRoleResolver)
		if !ok {
			return NewVaultError("Authorize", ErrUnauthorized, "org role resolver unavailable")
		}
		role, ok, err := orgRoleResolver.MemberRole(ctx, resourceOrg, req.Requester)
		if err != nil {
			return NewVaultError("Authorize", ErrUnauthorized, err.Error())
		}
		if ok && roleAllowed(role, policy.AllowedOrgRoles) {
			return nil
		}
	}

	if p.roleChecker == nil {
		return NewVaultError("Authorize", ErrUnauthorized, "role resolver unavailable")
	}

	for _, role := range policy.AllowedRoles {
		hasRole, err := p.roleChecker.HasRole(ctx, req.Requester, role)
		if err != nil {
			return NewVaultError("Authorize", ErrUnauthorized, err.Error())
		}
		if hasRole {
			return nil
		}
	}

	return NewVaultError("Authorize", ErrUnauthorized, "insufficient role")
}

func roleAllowed(role string, allowed []string) bool {
	if role == "" {
		return false
	}
	for _, candidate := range allowed {
		if role == candidate {
			return true
		}
	}
	return false
}

// OwnerOnlyAccessControl allows only the owner to access the blob.
type OwnerOnlyAccessControl struct{}

// Authorize enforces owner-only access.
func (OwnerOnlyAccessControl) Authorize(_ context.Context, req AccessRequest) error {
	if req.Requester == "" {
		return NewVaultError("Authorize", ErrUnauthorized, "requester required")
	}
	if req.Owner == "" || req.Requester != req.Owner {
		return NewVaultError("Authorize", ErrUnauthorized, "owner required")
	}
	return nil
}

// StaticRoleResolver is a simple role resolver for tests.
type StaticRoleResolver struct {
	Roles map[string]map[string]bool
}

// HasRole returns true if the role is present in the map.
func (r StaticRoleResolver) HasRole(_ context.Context, address, role string) (bool, error) {
	roles := r.Roles[address]
	if roles == nil {
		return false, nil
	}
	return roles[role], nil
}

// StaticOrgResolver is a simple org membership resolver for tests.
type StaticOrgResolver struct {
	Members map[string]map[string]bool
	Roles   map[string]map[string]string
}

// IsMember returns true if the address is in the org's member set.
func (r StaticOrgResolver) IsMember(_ context.Context, orgID, address string) (bool, error) {
	members := r.Members[orgID]
	if members == nil {
		return false, nil
	}
	return members[address], nil
}

// MemberRole returns the role for a member in an org.
func (r StaticOrgResolver) MemberRole(_ context.Context, orgID, address string) (string, bool, error) {
	roles := r.Roles[orgID]
	if roles == nil {
		return "", false, nil
	}
	role, ok := roles[address]
	return role, ok, nil
}

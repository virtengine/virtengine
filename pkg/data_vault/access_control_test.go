package data_vault

import (
	"context"
	"testing"
)

func TestPolicyAccessControl_OwnerAllowed(t *testing.T) {
	ctrl := NewPolicyAccessControl(DefaultAccessPolicy(), nil, nil)

	err := ctrl.Authorize(context.Background(), AccessRequest{
		Action:    AccessActionRead,
		Requester: "owner1",
		Owner:     "owner1",
		Scope:     ScopeVEID,
	})
	if err != nil {
		t.Fatalf("expected owner access to be allowed, got %v", err)
	}
}

func TestPolicyAccessControl_RoleAllowed(t *testing.T) {
	ctrl := NewPolicyAccessControl(DefaultAccessPolicy(),
		StaticRoleResolver{Roles: map[string]map[string]bool{
			"user1": {RoleAdministrator: true},
		}},
		StaticOrgResolver{Members: map[string]map[string]bool{
			"org-1": {"user1": true},
		}},
	)

	err := ctrl.Authorize(context.Background(), AccessRequest{
		Action:        AccessActionRead,
		Requester:     "user1",
		Owner:         "owner2",
		Scope:         ScopeSupport,
		OrgID:         "org-1",
		ResourceOrgID: "org-1",
	})
	if err != nil {
		t.Fatalf("expected role-based access to be allowed, got %v", err)
	}
}

func TestPolicyAccessControl_RoleDenied(t *testing.T) {
	ctrl := NewPolicyAccessControl(DefaultAccessPolicy(),
		StaticRoleResolver{Roles: map[string]map[string]bool{
			"user1": {RoleCustomer: true},
		}},
		StaticOrgResolver{Members: map[string]map[string]bool{
			"org-1": {"user1": true},
		}},
	)

	err := ctrl.Authorize(context.Background(), AccessRequest{
		Action:        AccessActionRead,
		Requester:     "user1",
		Owner:         "owner2",
		Scope:         ScopeSupport,
		OrgID:         "org-1",
		ResourceOrgID: "org-1",
	})
	if err == nil {
		t.Fatalf("expected role-based access to be denied")
	}
}

func TestPolicyAccessControl_OrgRoleAllowed(t *testing.T) {
	ctrl := NewPolicyAccessControl(DefaultAccessPolicy(),
		nil,
		StaticOrgResolver{
			Members: map[string]map[string]bool{
				"org-1": {"user1": true},
			},
			Roles: map[string]map[string]string{
				"org-1": {"user1": "admin"},
			},
		},
	)

	err := ctrl.Authorize(context.Background(), AccessRequest{
		Action:        AccessActionRead,
		Requester:     "user1",
		Owner:         "owner2",
		Scope:         ScopeSupport,
		OrgID:         "org-1",
		ResourceOrgID: "org-1",
	})
	if err != nil {
		t.Fatalf("expected org role access to be allowed, got %v", err)
	}
}

func TestPolicyAccessControl_OrgMismatchDenied(t *testing.T) {
	ctrl := NewPolicyAccessControl(DefaultAccessPolicy(),
		StaticRoleResolver{Roles: map[string]map[string]bool{
			"user1": {RoleAdministrator: true},
		}},
		StaticOrgResolver{Members: map[string]map[string]bool{
			"org-1": {"user1": true},
		}},
	)

	err := ctrl.Authorize(context.Background(), AccessRequest{
		Action:        AccessActionRead,
		Requester:     "user1",
		Owner:         "owner2",
		Scope:         ScopeSupport,
		OrgID:         "org-2",
		ResourceOrgID: "org-1",
	})
	if err == nil {
		t.Fatalf("expected org mismatch to be denied")
	}
}

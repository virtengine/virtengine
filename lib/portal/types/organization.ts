/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Organization types for x/group module integration.
 * Supports team management, role-based access, and billing aggregation.
 */

// =============================================================================
// Organization Types
// =============================================================================

export interface Organization {
  id: string;
  name: string;
  description?: string;
  admin: string;
  totalWeight: string;
  createdAt: Date;
  metadata: OrganizationMetadata;
}

export interface OrganizationMetadata {
  name: string;
  description?: string;
  website?: string;
  logo?: string;
}

// =============================================================================
// Member Types
// =============================================================================

export type OrganizationRole = "admin" | "member" | "viewer";

export interface OrganizationMember {
  address: string;
  weight: string;
  role: OrganizationRole;
  addedAt: Date;
  metadata?: MemberMetadata;
}

export interface MemberMetadata {
  name?: string;
  email?: string;
}

// =============================================================================
// Invite Types
// =============================================================================

export type InviteStatus = "pending" | "accepted" | "rejected" | "expired";

export interface OrganizationInvite {
  id: string;
  organizationId: string;
  inviterAddress: string;
  inviteeAddress: string;
  role: OrganizationRole;
  createdAt: Date;
  expiresAt: Date;
  status: InviteStatus;
}

// =============================================================================
// Request Types
// =============================================================================

export interface CreateOrganizationRequest {
  name: string;
  description?: string;
  initialMembers?: { address: string; role: OrganizationRole }[];
}

export interface InviteMemberRequest {
  address: string;
  role: OrganizationRole;
}

// =============================================================================
// Billing Types
// =============================================================================

export interface OrganizationBillingPeriod {
  period: string;
  amount: number;
  deployments: number;
}

export interface OrganizationBillingSummary {
  totalSpend: number;
  currentPeriodSpend: number;
  previousPeriodSpend: number;
  changePercent: number;
  byMember: {
    address: string;
    amount: number;
    percentage: number;
  }[];
  history: OrganizationBillingPeriod[];
}

// =============================================================================
// Role Permissions
// =============================================================================

export const ROLE_PERMISSIONS: Record<OrganizationRole, string[]> = {
  admin: [
    "manage_members",
    "manage_settings",
    "create_deployments",
    "view_deployments",
    "manage_billing",
    "view_billing",
  ],
  member: ["create_deployments", "view_deployments", "view_billing"],
  viewer: ["view_deployments", "view_billing"],
};

export function hasPermission(
  role: OrganizationRole,
  permission: string,
): boolean {
  return ROLE_PERMISSIONS[role]?.includes(permission) ?? false;
}

// =============================================================================
// Role Display
// =============================================================================

export const ROLE_LABELS: Record<OrganizationRole, string> = {
  admin: "Admin",
  member: "Member",
  viewer: "Viewer",
};

export const ROLE_DESCRIPTIONS: Record<OrganizationRole, string> = {
  admin: "Full access, can manage members and settings",
  member: "Can create and manage deployments",
  viewer: "Read-only access to deployments",
};

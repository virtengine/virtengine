/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * useOrganization Hook
 * Organization management via Cosmos SDK x/group module.
 *
 * Provides CRUD operations for organizations, member management,
 * and billing aggregation.
 */

import {
  useState,
  useCallback,
  useEffect,
  useContext,
  useMemo,
  createContext,
  createElement,
} from "react";
import type { ReactNode } from "react";
import type { QueryClient } from "../types/chain";
import type {
  Organization,
  OrganizationMember,
  OrganizationRole,
  OrganizationBillingSummary,
  CreateOrganizationRequest,
  InviteMemberRequest,
} from "../types/organization";

// =============================================================================
// State Types
// =============================================================================

export interface OrganizationState {
  isLoading: boolean;
  organizations: Organization[];
  selectedOrgId: string | null;
  error: string | null;
}

export interface OrganizationDetailState {
  isLoading: boolean;
  organization: Organization | null;
  members: OrganizationMember[];
  billing: OrganizationBillingSummary | null;
  error: string | null;
}

export interface OrganizationActions {
  fetchOrganizations: () => Promise<void>;
  selectOrganization: (orgId: string | null) => void;
  createOrganization: (
    request: CreateOrganizationRequest,
  ) => Promise<Organization>;
  fetchOrganizationDetail: (orgId: string) => Promise<void>;
  inviteMember: (orgId: string, request: InviteMemberRequest) => Promise<void>;
  removeMember: (orgId: string, memberAddress: string) => Promise<void>;
  updateMemberRole: (
    orgId: string,
    memberAddress: string,
    role: OrganizationRole,
  ) => Promise<void>;
  leaveOrganization: (orgId: string) => Promise<void>;
  fetchBilling: (orgId: string) => Promise<void>;
}

export interface OrganizationContextValue {
  state: OrganizationState;
  detail: OrganizationDetailState;
  actions: OrganizationActions;
  selectedOrganization: Organization | null;
  currentUserRole: OrganizationRole | null;
}

// =============================================================================
// Context
// =============================================================================

const OrganizationContext = createContext<OrganizationContextValue | null>(
  null,
);

// =============================================================================
// Helpers
// =============================================================================

function parseOrganization(raw: Record<string, unknown>): Organization {
  const metadataStr = (raw.metadata as string) || "{}";
  let metadata: Record<string, unknown>;
  try {
    metadata = JSON.parse(metadataStr) as Record<string, unknown>;
  } catch {
    metadata = {};
  }
  return {
    id: raw.id as string,
    name: (metadata.name as string) || `Organization ${raw.id}`,
    description: metadata.description as string | undefined,
    admin: raw.admin as string,
    totalWeight: raw.total_weight as string,
    createdAt: new Date(raw.created_at as string),
    metadata: {
      name: (metadata.name as string) || `Organization ${raw.id}`,
      description: metadata.description as string | undefined,
      website: metadata.website as string | undefined,
      logo: metadata.logo as string | undefined,
    },
  };
}

function parseMember(raw: Record<string, unknown>): OrganizationMember {
  const member = (raw.member || raw) as Record<string, unknown>;
  const metadataStr = (member.metadata as string) || "{}";
  let metadata: Record<string, unknown>;
  try {
    metadata = JSON.parse(metadataStr) as Record<string, unknown>;
  } catch {
    metadata = {};
  }
  const weight = member.weight as string;
  return {
    address: member.address as string,
    weight,
    role:
      (metadata.role as OrganizationRole) ||
      (weight === "1" ? "member" : "viewer"),
    addedAt: new Date(member.added_at as string),
    metadata: {
      name: metadata.name as string | undefined,
      email: metadata.email as string | undefined,
    },
  };
}

// =============================================================================
// Provider
// =============================================================================

export interface OrganizationProviderProps {
  children: ReactNode;
  queryClient?: QueryClient;
  accountAddress?: string | null;
}

export function OrganizationProvider({
  children,
  queryClient,
  accountAddress,
}: OrganizationProviderProps): JSX.Element {
  const [state, setState] = useState<OrganizationState>({
    isLoading: false,
    organizations: [],
    selectedOrgId: null,
    error: null,
  });

  const [detail, setDetail] = useState<OrganizationDetailState>({
    isLoading: false,
    organization: null,
    members: [],
    billing: null,
    error: null,
  });

  const fetchOrganizations = useCallback(async () => {
    if (!queryClient || !accountAddress) {
      setState((prev) => ({ ...prev, isLoading: false }));
      return;
    }

    setState((prev) => ({ ...prev, isLoading: true, error: null }));
    try {
      const result = await queryClient.query<{
        groups: Record<string, unknown>[];
      }>("/cosmos/group/v1/groups_by_member", { address: accountAddress });
      const organizations = (result.groups ?? []).map(parseOrganization);
      setState((prev) => ({
        ...prev,
        isLoading: false,
        organizations,
        error: null,
      }));
    } catch (error) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to load organizations",
      }));
    }
  }, [queryClient, accountAddress]);

  const selectOrganization = useCallback((orgId: string | null) => {
    setState((prev) => ({ ...prev, selectedOrgId: orgId }));
  }, []);

  const createOrganization = useCallback(
    async (request: CreateOrganizationRequest): Promise<Organization> => {
      if (!queryClient || !accountAddress) {
        throw new Error("Wallet not connected");
      }

      const metadata = JSON.stringify({
        name: request.name,
        description: request.description,
      });

      const members = [
        {
          address: accountAddress,
          weight: "1",
          metadata: JSON.stringify({ role: "admin" }),
        },
        ...(request.initialMembers || []).map((m) => ({
          address: m.address,
          weight: m.role === "viewer" ? "0" : "1",
          metadata: JSON.stringify({ role: m.role }),
        })),
      ];

      const msg = {
        typeUrl: "/cosmos.group.v1.MsgCreateGroup",
        value: { admin: accountAddress, members, metadata },
      };

      // Placeholder: in production, this would sign and broadcast the tx
      await queryClient.query("/cosmos/tx/v1beta1/simulate", {
        tx_bytes: JSON.stringify(msg),
      });

      const org: Organization = {
        id: `pending-${Date.now()}`,
        name: request.name,
        description: request.description,
        admin: accountAddress,
        totalWeight: String(members.length),
        createdAt: new Date(),
        metadata: { name: request.name, description: request.description },
      };

      setState((prev) => ({
        ...prev,
        organizations: [...prev.organizations, org],
      }));

      return org;
    },
    [queryClient, accountAddress],
  );

  const fetchOrganizationDetail = useCallback(
    async (orgId: string) => {
      if (!queryClient) {
        return;
      }

      setDetail((prev) => ({ ...prev, isLoading: true, error: null }));
      try {
        const [infoResult, membersResult] = await Promise.all([
          queryClient.query<{ info: Record<string, unknown> }>(
            `/cosmos/group/v1/group_info/${orgId}`,
          ),
          queryClient.query<{ members: Record<string, unknown>[] }>(
            `/cosmos/group/v1/group_members/${orgId}`,
          ),
        ]);

        setDetail({
          isLoading: false,
          organization: parseOrganization(infoResult.info),
          members: (membersResult.members ?? []).map(parseMember),
          billing: null,
          error: null,
        });
      } catch (error) {
        setDetail((prev) => ({
          ...prev,
          isLoading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load organization",
        }));
      }
    },
    [queryClient],
  );

  const inviteMember = useCallback(
    async (orgId: string, request: InviteMemberRequest) => {
      if (!queryClient || !accountAddress) {
        throw new Error("Wallet not connected");
      }

      const msg = {
        typeUrl: "/cosmos.group.v1.MsgUpdateGroupMembers",
        value: {
          admin: accountAddress,
          groupId: orgId,
          memberUpdates: [
            {
              address: request.address,
              weight: request.role === "viewer" ? "0" : "1",
              metadata: JSON.stringify({ role: request.role }),
            },
          ],
        },
      };

      // Placeholder: in production, this would sign and broadcast the tx
      await queryClient.query("/cosmos/tx/v1beta1/simulate", {
        tx_bytes: JSON.stringify(msg),
      });

      setDetail((prev) => ({
        ...prev,
        members: [
          ...prev.members,
          {
            address: request.address,
            weight: request.role === "viewer" ? "0" : "1",
            role: request.role,
            addedAt: new Date(),
            metadata: {},
          },
        ],
      }));
    },
    [queryClient, accountAddress],
  );

  const removeMember = useCallback(
    async (orgId: string, memberAddress: string) => {
      if (!queryClient || !accountAddress) {
        throw new Error("Wallet not connected");
      }

      const msg = {
        typeUrl: "/cosmos.group.v1.MsgUpdateGroupMembers",
        value: {
          admin: accountAddress,
          groupId: orgId,
          memberUpdates: [
            { address: memberAddress, weight: "0", metadata: "" },
          ],
        },
      };

      // Placeholder: in production, this would sign and broadcast the tx
      await queryClient.query("/cosmos/tx/v1beta1/simulate", {
        tx_bytes: JSON.stringify(msg),
      });

      setDetail((prev) => ({
        ...prev,
        members: prev.members.filter((m) => m.address !== memberAddress),
      }));
    },
    [queryClient, accountAddress],
  );

  const updateMemberRole = useCallback(
    async (orgId: string, memberAddress: string, role: OrganizationRole) => {
      if (!queryClient || !accountAddress) {
        throw new Error("Wallet not connected");
      }

      const msg = {
        typeUrl: "/cosmos.group.v1.MsgUpdateGroupMembers",
        value: {
          admin: accountAddress,
          groupId: orgId,
          memberUpdates: [
            {
              address: memberAddress,
              weight: role === "viewer" ? "0" : "1",
              metadata: JSON.stringify({ role }),
            },
          ],
        },
      };

      // Placeholder: in production, this would sign and broadcast the tx
      await queryClient.query("/cosmos/tx/v1beta1/simulate", {
        tx_bytes: JSON.stringify(msg),
      });

      setDetail((prev) => ({
        ...prev,
        members: prev.members.map((m) =>
          m.address === memberAddress
            ? { ...m, role, weight: role === "viewer" ? "0" : "1" }
            : m,
        ),
      }));
    },
    [queryClient, accountAddress],
  );

  const leaveOrganization = useCallback(
    async (orgId: string) => {
      if (!queryClient || !accountAddress) {
        throw new Error("Wallet not connected");
      }

      const msg = {
        typeUrl: "/cosmos.group.v1.MsgLeaveGroup",
        value: { address: accountAddress, groupId: orgId },
      };

      // Placeholder: in production, this would sign and broadcast the tx
      await queryClient.query("/cosmos/tx/v1beta1/simulate", {
        tx_bytes: JSON.stringify(msg),
      });

      setState((prev) => ({
        ...prev,
        organizations: prev.organizations.filter((o) => o.id !== orgId),
      }));
    },
    [queryClient, accountAddress],
  );

  const fetchBilling = useCallback(
    async (orgId: string) => {
      if (!queryClient) {
        return;
      }

      try {
        const result = await queryClient.query<{
          billing: OrganizationBillingSummary;
        }>(`/organizations/${orgId}/billing`);
        setDetail((prev) => ({
          ...prev,
          billing: result.billing ?? null,
        }));
      } catch {
        // Billing is optional; silently fail
      }
    },
    [queryClient],
  );

  useEffect(() => {
    void fetchOrganizations();
  }, [fetchOrganizations]);

  const selectedOrganization = useMemo(() => {
    return (
      state.organizations.find((o) => o.id === state.selectedOrgId) ?? null
    );
  }, [state.organizations, state.selectedOrgId]);

  const currentUserRole = useMemo(() => {
    if (!accountAddress || !detail.members.length) return null;
    const member = detail.members.find((m) => m.address === accountAddress);
    return member?.role ?? null;
  }, [accountAddress, detail.members]);

  const actions: OrganizationActions = useMemo(
    () => ({
      fetchOrganizations,
      selectOrganization,
      createOrganization,
      fetchOrganizationDetail,
      inviteMember,
      removeMember,
      updateMemberRole,
      leaveOrganization,
      fetchBilling,
    }),
    [
      fetchOrganizations,
      selectOrganization,
      createOrganization,
      fetchOrganizationDetail,
      inviteMember,
      removeMember,
      updateMemberRole,
      leaveOrganization,
      fetchBilling,
    ],
  );

  const contextValue = useMemo<OrganizationContextValue>(
    () => ({
      state,
      detail,
      actions,
      selectedOrganization,
      currentUserRole,
    }),
    [state, detail, actions, selectedOrganization, currentUserRole],
  );

  return createElement(
    OrganizationContext.Provider,
    { value: contextValue },
    children,
  );
}

// =============================================================================
// Hook
// =============================================================================

export function useOrganization(): OrganizationContextValue {
  const context = useContext(OrganizationContext);
  if (!context) {
    throw new Error(
      "useOrganization must be used within an OrganizationProvider",
    );
  }
  return context;
}

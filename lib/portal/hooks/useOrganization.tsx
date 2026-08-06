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
  useRef,
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
import {
  OrganizationMutationError,
  buildOrganizationMutationRequest,
  digestOrganizationMutationRequest,
  requireOrganizationMutationAdapter,
  validateCommittedOrganizationMutation,
  type CommittedOrganizationMutation,
  type OrganizationMutationAction,
  type OrganizationMutationAdapter,
} from "../components/organization/organization-mutation";

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
  chainId: string;
  accountAddress?: string | null;
  mutationAdapter?: OrganizationMutationAdapter;
}

export function OrganizationProvider({
  children,
  queryClient,
  chainId,
  accountAddress,
  mutationAdapter,
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

  const queryGeneration = useRef(0);
  const stateQueryGeneration = useRef(0);
  const queryAuthority = useRef({ queryClient, chainId, accountAddress });
  const organizationsRequest = useRef(0);
  const detailRequest = useRef(0);
  const billingRequest = useRef(0);
  const detailOrganizationId = useRef<string | null>(null);
  const queryResetPending = useRef(false);
  if (
    queryAuthority.current.queryClient !== queryClient ||
    queryAuthority.current.chainId !== chainId ||
    queryAuthority.current.accountAddress !== accountAddress
  ) {
    queryAuthority.current = { queryClient, chainId, accountAddress };
    queryGeneration.current += 1;
    organizationsRequest.current += 1;
    detailRequest.current += 1;
    billingRequest.current += 1;
    detailOrganizationId.current = null;
    queryResetPending.current = true;
  }
  const renderQueryGeneration = queryGeneration.current;
  const effectiveState: OrganizationState =
    stateQueryGeneration.current === renderQueryGeneration
      ? state
      : {
          isLoading: false,
          organizations: [],
          selectedOrgId: null,
          error: null,
        };
  const effectiveDetail: OrganizationDetailState =
    stateQueryGeneration.current === renderQueryGeneration
      ? detail
      : {
          isLoading: false,
          organization: null,
          members: [],
          billing: null,
          error: null,
        };

  useEffect(() => {
    if (!queryResetPending.current) return;
    queryResetPending.current = false;
    stateQueryGeneration.current = renderQueryGeneration;
    setState({
      isLoading: false,
      organizations: [],
      selectedOrgId: null,
      error: null,
    });
    setDetail({
      isLoading: false,
      organization: null,
      members: [],
      billing: null,
      error: null,
    });
  }, [renderQueryGeneration]);

  const mutationGeneration = useRef(0);
  const mutationAuthority = useRef({
    mutationAdapter,
    chainId,
    accountAddress,
  });
  const mutationsInFlight = useRef(new Set<string>());
  const activeMutationControllers = useRef(new Map<string, AbortController>());
  if (
    mutationAuthority.current.mutationAdapter !== mutationAdapter ||
    mutationAuthority.current.chainId !== chainId ||
    mutationAuthority.current.accountAddress !== accountAddress
  ) {
    for (const controller of activeMutationControllers.current.values()) {
      controller.abort();
    }
    activeMutationControllers.current.clear();
    mutationsInFlight.current.clear();
    mutationAuthority.current = { mutationAdapter, chainId, accountAddress };
    mutationGeneration.current += 1;
  }
  const renderMutationGeneration = mutationGeneration.current;

  useEffect(
    () => () => {
      mutationGeneration.current += 1;
      for (const controller of activeMutationControllers.current.values()) {
        controller.abort();
      }
      activeMutationControllers.current.clear();
      mutationsInFlight.current.clear();
    },
    [],
  );

  const fetchOrganizations = useCallback(async () => {
    if (!queryClient || !accountAddress) {
      setState((prev) => ({ ...prev, isLoading: false }));
      return;
    }

    const generation = renderQueryGeneration;
    const requestId = ++organizationsRequest.current;

    setState((prev) => ({ ...prev, isLoading: true, error: null }));
    try {
      const result = await queryClient.query<{
        groups: Record<string, unknown>[];
      }>("/cosmos/group/v1/groups_by_member", { address: accountAddress });
      const organizations = (result.groups ?? []).map(parseOrganization);
      if (
        generation !== queryGeneration.current ||
        requestId !== organizationsRequest.current
      ) {
        return;
      }
      setState((prev) => ({
        ...prev,
        isLoading: false,
        organizations,
        error: null,
      }));
    } catch (error) {
      if (
        generation !== queryGeneration.current ||
        requestId !== organizationsRequest.current
      ) {
        return;
      }
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to load organizations",
      }));
    }
  }, [queryClient, accountAddress, renderQueryGeneration]);

  const selectOrganization = useCallback((orgId: string | null) => {
    const normalized = orgId?.trim() || null;
    detailRequest.current += 1;
    billingRequest.current += 1;
    detailOrganizationId.current = normalized;
    setState((prev) => ({ ...prev, selectedOrgId: normalized }));
    setDetail({
      isLoading: false,
      organization: null,
      members: [],
      billing: null,
      error: null,
    });
  }, []);

  const mutateOrganization = useCallback(
    async (
      action: OrganizationMutationAction,
      input:
        | CreateOrganizationRequest
        | { organizationId: string; invitation: InviteMemberRequest }
        | { organizationId: string; memberAddress: string }
        | {
            organizationId: string;
            memberAddress: string;
            role: OrganizationRole;
          }
        | { organizationId: string },
      lockKey: string,
    ): Promise<CommittedOrganizationMutation> => {
      if (
        !accountAddress ||
        renderMutationGeneration !== mutationGeneration.current
      ) {
        throw new OrganizationMutationError("feature_unavailable");
      }
      const binding = { chainId, accountAddress };
      const adapter = requireOrganizationMutationAdapter(
        mutationAdapter,
        binding,
      );
      const normalizedLockKey = lockKey === "create" ? lockKey : lockKey.trim();
      if (mutationsInFlight.current.has(normalizedLockKey)) {
        throw new OrganizationMutationError("submission_in_progress");
      }
      mutationsInFlight.current.add(normalizedLockKey);
      try {
        const request = buildOrganizationMutationRequest(
          action,
          binding,
          input,
        );
        const requestDigest = await digestOrganizationMutationRequest(request);
        if (
          renderMutationGeneration !== mutationGeneration.current ||
          mutationAuthority.current.mutationAdapter !== adapter
        ) {
          throw new OrganizationMutationError("request_changed");
        }
        const controller = new AbortController();
        activeMutationControllers.current.set(normalizedLockKey, controller);
        const rawResult = await adapter.mutateOrganization(request, {
          requestDigest,
          idempotencyKey: requestDigest,
          signal: controller.signal,
        });
        if (
          renderMutationGeneration !== mutationGeneration.current ||
          mutationAuthority.current.mutationAdapter !== adapter
        ) {
          controller.abort();
          throw new OrganizationMutationError("submission_cancelled");
        }
        return validateCommittedOrganizationMutation(
          rawResult,
          request,
          requestDigest,
        );
      } finally {
        if (renderMutationGeneration === mutationGeneration.current) {
          mutationsInFlight.current.delete(normalizedLockKey);
          activeMutationControllers.current.delete(normalizedLockKey);
        }
      }
    },
    [accountAddress, chainId, mutationAdapter, renderMutationGeneration],
  );

  const createOrganization = useCallback(
    async (request: CreateOrganizationRequest): Promise<Organization> => {
      const result = await mutateOrganization("create", request, "create");
      if (result.action !== "create") {
        throw new OrganizationMutationError("invalid_committed_result");
      }
      setState((prev) => ({
        ...prev,
        organizations: prev.organizations.some(
          (organization) => organization.id === result.organization.id,
        )
          ? prev.organizations.map((organization) =>
              organization.id === result.organization.id
                ? result.organization
                : organization,
            )
          : [...prev.organizations, result.organization],
      }));
      return result.organization;
    },
    [mutateOrganization],
  );

  const fetchOrganizationDetail = useCallback(
    async (orgId: string) => {
      if (!queryClient) {
        return;
      }

      const normalizedOrgId = orgId.trim();
      const generation = renderQueryGeneration;
      const requestId = ++detailRequest.current;
      detailOrganizationId.current = normalizedOrgId;

      setDetail((prev) => ({ ...prev, isLoading: true, error: null }));
      try {
        const [infoResult, membersResult] = await Promise.all([
          queryClient.query<{ info: Record<string, unknown> }>(
            `/cosmos/group/v1/group_info/${normalizedOrgId}`,
          ),
          queryClient.query<{ members: Record<string, unknown>[] }>(
            `/cosmos/group/v1/group_members/${normalizedOrgId}`,
          ),
        ]);

        if (
          generation !== queryGeneration.current ||
          requestId !== detailRequest.current ||
          detailOrganizationId.current !== normalizedOrgId
        ) {
          return;
        }
        const organization = parseOrganization(infoResult.info);
        if (organization.id !== normalizedOrgId) return;
        setDetail({
          isLoading: false,
          organization,
          members: (membersResult.members ?? []).map(parseMember),
          billing: null,
          error: null,
        });
      } catch (error) {
        if (
          generation !== queryGeneration.current ||
          requestId !== detailRequest.current ||
          detailOrganizationId.current !== normalizedOrgId
        ) {
          return;
        }
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
    [queryClient, renderQueryGeneration],
  );

  const inviteMember = useCallback(
    async (orgId: string, request: InviteMemberRequest) => {
      const result = await mutateOrganization(
        "invite",
        { organizationId: orgId, invitation: request },
        orgId,
      );
      if (result.action !== "invite") {
        throw new OrganizationMutationError("invalid_committed_result");
      }
      const normalizedOrgId = orgId.trim();
      if (detailOrganizationId.current === null)
        detailOrganizationId.current = normalizedOrgId;
      if (detailOrganizationId.current === normalizedOrgId) {
        setDetail((prev) => ({ ...prev, members: [...result.members] }));
      }
    },
    [mutateOrganization],
  );

  const removeMember = useCallback(
    async (orgId: string, memberAddress: string) => {
      const result = await mutateOrganization(
        "remove",
        { organizationId: orgId, memberAddress },
        orgId,
      );
      if (result.action !== "remove") {
        throw new OrganizationMutationError("invalid_committed_result");
      }
      const normalizedOrgId = orgId.trim();
      if (detailOrganizationId.current === null)
        detailOrganizationId.current = normalizedOrgId;
      if (detailOrganizationId.current === normalizedOrgId) {
        setDetail((prev) => ({ ...prev, members: [...result.members] }));
      }
    },
    [mutateOrganization],
  );

  const updateMemberRole = useCallback(
    async (orgId: string, memberAddress: string, role: OrganizationRole) => {
      const result = await mutateOrganization(
        "update_role",
        { organizationId: orgId, memberAddress, role },
        orgId,
      );
      if (result.action !== "update_role") {
        throw new OrganizationMutationError("invalid_committed_result");
      }
      const normalizedOrgId = orgId.trim();
      if (detailOrganizationId.current === null)
        detailOrganizationId.current = normalizedOrgId;
      if (detailOrganizationId.current === normalizedOrgId) {
        setDetail((prev) => ({ ...prev, members: [...result.members] }));
      }
    },
    [mutateOrganization],
  );

  const leaveOrganization = useCallback(
    async (orgId: string) => {
      const result = await mutateOrganization(
        "leave",
        { organizationId: orgId },
        orgId,
      );
      if (result.action !== "leave") {
        throw new OrganizationMutationError("invalid_committed_result");
      }
      setState((prev) => ({
        ...prev,
        organizations: prev.organizations.filter(
          (organization) => organization.id !== result.organizationId,
        ),
        selectedOrgId:
          prev.selectedOrgId === result.organizationId
            ? null
            : prev.selectedOrgId,
      }));
      if (detailOrganizationId.current === result.organizationId) {
        detailOrganizationId.current = null;
        detailRequest.current += 1;
        billingRequest.current += 1;
        setDetail({
          isLoading: false,
          organization: null,
          members: [],
          billing: null,
          error: null,
        });
      }
    },
    [mutateOrganization],
  );

  const fetchBilling = useCallback(
    async (orgId: string) => {
      if (!queryClient) {
        return;
      }

      const normalizedOrgId = orgId.trim();
      const generation = renderQueryGeneration;
      const requestId = ++billingRequest.current;

      try {
        const result = await queryClient.query<{
          billing: OrganizationBillingSummary;
        }>(`/organizations/${normalizedOrgId}/billing`);
        if (
          generation !== queryGeneration.current ||
          requestId !== billingRequest.current ||
          detailOrganizationId.current !== normalizedOrgId
        ) {
          return;
        }
        setDetail((prev) => ({
          ...prev,
          billing: result.billing ?? null,
        }));
      } catch {
        // Billing is optional; silently fail
      }
    },
    [queryClient, renderQueryGeneration],
  );

  useEffect(() => {
    void fetchOrganizations();
  }, [fetchOrganizations]);

  const selectedOrganization = useMemo(() => {
    return (
      effectiveState.organizations.find(
        (o) => o.id === effectiveState.selectedOrgId,
      ) ?? null
    );
  }, [effectiveState.organizations, effectiveState.selectedOrgId]);

  const currentUserRole = useMemo(() => {
    if (!accountAddress || !effectiveDetail.members.length) return null;
    const member = effectiveDetail.members.find(
      (m) => m.address === accountAddress,
    );
    return member?.role ?? null;
  }, [accountAddress, effectiveDetail.members]);

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
      state: effectiveState,
      detail: effectiveDetail,
      actions,
      selectedOrganization,
      currentUserRole,
    }),
    [
      effectiveState,
      effectiveDetail,
      actions,
      selectedOrganization,
      currentUserRole,
    ],
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

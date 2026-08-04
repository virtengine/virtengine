import * as React from "react";
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type {
  OrganizationMutationAdapter,
  OrganizationMutationContext,
  OrganizationMutationRequest,
} from "../../components/organization/organization-mutation";
import {
  buildOrganizationMutationRequest,
  digestOrganizationMutationRequest,
} from "../../components/organization/organization-mutation";
import {
  OrganizationProvider,
  useOrganization,
  type OrganizationContextValue,
} from "../../hooks/useOrganization";
import type {
  Organization,
  OrganizationMember,
} from "../../types/organization";
import type { QueryClient } from "../../types/chain";

const binding = {
  chainId: "virtengine-1",
  accountAddress: "virtengine1admin",
} as const;

const organization = (
  id: string,
  overrides: Partial<Organization> = {},
): Organization => ({
  id,
  name: `Organization ${id}`,
  admin: binding.accountAddress,
  totalWeight: "1",
  createdAt: new Date("2026-08-04T00:00:00.000Z"),
  metadata: { name: `Organization ${id}` },
  ...overrides,
});

const member = (
  address: string,
  role: OrganizationMember["role"] = "member",
): OrganizationMember => ({
  address,
  weight: role === "viewer" ? "0" : "1",
  role,
  addedAt: new Date("2026-08-04T00:00:00.000Z"),
  metadata: {},
});

const committed = (
  request: OrganizationMutationRequest,
  context: OrganizationMutationContext,
  projection: Record<string, unknown>,
) => ({
  status: "committed",
  code: 0,
  txHash: `tx-${request.action}`,
  blockHeight: 42,
  operationId: `operation-${request.action}`,
  requestDigest: context.requestDigest,
  idempotencyKey: context.idempotencyKey,
  request,
  action: request.action,
  ...projection,
});

describe("organization mutation authority", () => {
  let container: HTMLDivElement;
  let root: Root;
  let mounted: boolean;
  let value: OrganizationContextValue;

  const Consumer = () => {
    value = useOrganization();
    return null;
  };

  const renderProvider = async (
    mutationAdapter?: OrganizationMutationAdapter,
    chainId = binding.chainId,
    accountAddress: string | null = binding.accountAddress,
    queryClient?: QueryClient,
  ) => {
    await act(async () =>
      root.render(
        <OrganizationProvider
          chainId={chainId}
          accountAddress={accountAddress}
          mutationAdapter={mutationAdapter}
          queryClient={queryClient}
        >
          <Consumer />
        </OrganizationProvider>,
      ),
    );
  };

  beforeEach(() => {
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    container = document.createElement("div");
    document.body.appendChild(container);
    root = createRoot(container);
    mounted = true;
  });

  afterEach(async () => {
    if (mounted) await act(async () => root.unmount());
    container.remove();
  });

  it("fails unavailable without querying, simulating, or changing state", async () => {
    await renderProvider();
    await expect(
      value.actions.createOrganization({ name: "Local fallback" }),
    ).rejects.toMatchObject({ code: "feature_unavailable" });
    expect(value.state.organizations).toEqual([]);
    expect(value.detail.members).toEqual([]);
  });

  it("upserts create from the committed server organization ID", async () => {
    const adapter: OrganizationMutationAdapter = {
      ...binding,
      mutateOrganization: vi.fn((request, context) =>
        Promise.resolve(
          committed(request, context, {
            organization: organization("server-org"),
          }),
        ),
      ),
    };
    await renderProvider(adapter);
    let result!: Organization;
    await act(async () => {
      result = await value.actions.createOrganization({ name: "Requested" });
    });
    expect(result.id).toBe("server-org");
    expect(Object.isFrozen(result)).toBe(true);
    expect(value.state.organizations).toEqual([result]);
  });

  it("rejects malformed committed evidence without changing state", async () => {
    await renderProvider({
      ...binding,
      mutateOrganization: vi.fn().mockResolvedValue({
        status: "committed",
        code: 0,
      }),
    });
    await expect(
      value.actions.createOrganization({ name: "Rejected" }),
    ).rejects.toMatchObject({ code: "invalid_committed_result" });
    expect(value.state.organizations).toEqual([]);
  });

  it("rejects non-date committed timestamps", async () => {
    await renderProvider({
      ...binding,
      mutateOrganization: vi.fn((request, context) =>
        Promise.resolve(
          committed(request, context, {
            organization: { ...organization("bad-date"), createdAt: null },
          }),
        ),
      ),
    });
    await expect(
      value.actions.createOrganization({ name: "Bad date" }),
    ).rejects.toMatchObject({ code: "invalid_committed_result" });
    expect(value.state.organizations).toEqual([]);
  });

  it("rejects calendar-invalid committed timestamp strings", async () => {
    await renderProvider({
      ...binding,
      mutateOrganization: vi.fn((request, context) =>
        Promise.resolve(
          committed(request, context, {
            organization: {
              ...organization("bad-calendar-date"),
              createdAt: "2026-02-30T00:00:00Z",
            },
          }),
        ),
      ),
    });
    await expect(
      value.actions.createOrganization({ name: "Bad calendar date" }),
    ).rejects.toMatchObject({ code: "invalid_committed_result" });
    expect(value.state.organizations).toEqual([]);
  });

  it("rejects malformed optional organization metadata", async () => {
    await renderProvider({
      ...binding,
      mutateOrganization: vi.fn((request, context) =>
        Promise.resolve(
          committed(request, context, {
            organization: {
              ...organization("bad-metadata"),
              description: 42,
              metadata: { name: "Bad", website: { unsafe: true } },
            },
          }),
        ),
      ),
    });
    await expect(
      value.actions.createOrganization({ name: "Bad metadata" }),
    ).rejects.toMatchObject({ code: "invalid_committed_result" });
    expect(value.state.organizations).toEqual([]);
  });

  it("rejects non-object committed member metadata", async () => {
    await renderProvider({
      ...binding,
      mutateOrganization: vi.fn((request, context) =>
        Promise.resolve(
          committed(request, context, {
            members: [{ ...member("bad-member"), metadata: [] }],
          }),
        ),
      ),
    });
    await expect(
      value.actions.inviteMember("org-1", {
        address: "bad-member",
        role: "member",
      }),
    ).rejects.toMatchObject({ code: "invalid_committed_result" });
    expect(value.detail.members).toEqual([]);
  });

  it("replaces members from committed invite authority", async () => {
    const authoritativeMembers = [
      member(binding.accountAddress, "admin"),
      member("virtengine1invited", "viewer"),
    ];
    await renderProvider({
      ...binding,
      mutateOrganization: vi.fn((request, context) =>
        Promise.resolve(
          committed(request, context, { members: authoritativeMembers }),
        ),
      ),
    });
    await act(async () => {
      await value.actions.inviteMember("org-1", {
        address: "virtengine1invited",
        role: "viewer",
      });
    });
    expect(value.detail.members).toEqual(authoritativeMembers);
    expect(Object.isFrozen(value.detail.members[0])).toBe(true);
  });

  it("applies exact member lists for role/remove and exact organization leave", async () => {
    const orgOne = organization("org-1");
    const orgTwo = organization("org-2");
    const roleMembers = [member("member-1", "admin")];
    const removeMembers = [member("member-2", "viewer")];
    const adapter: OrganizationMutationAdapter = {
      ...binding,
      mutateOrganization: vi.fn((request, context) => {
        if (request.action === "create") {
          const id = request.organization.name === "One" ? "org-1" : "org-2";
          return Promise.resolve(
            committed(request, context, {
              organization: id === "org-1" ? orgOne : orgTwo,
            }),
          );
        }
        if (request.action === "update_role") {
          return Promise.resolve(
            committed(request, context, { members: roleMembers }),
          );
        }
        if (request.action === "remove") {
          return Promise.resolve(
            committed(request, context, { members: removeMembers }),
          );
        }
        return Promise.resolve(
          committed(request, context, {
            organizationId: request.organizationId,
          }),
        );
      }),
    };
    await renderProvider(adapter);
    await act(async () => {
      await value.actions.createOrganization({ name: "One" });
      await value.actions.createOrganization({ name: "Two" });
      await value.actions.updateMemberRole("org-1", "member-1", "admin");
    });
    expect(value.detail.members).toEqual(roleMembers);
    await act(async () => {
      await value.actions.removeMember("org-1", "member-1");
    });
    expect(value.detail.members).toEqual(removeMembers);
    await act(async () => {
      await value.actions.leaveOrganization("org-1");
    });
    expect(value.state.organizations.map(({ id }) => id)).toEqual(["org-2"]);
  });

  it("synchronously blocks duplicate and conflicting mutations per organization", async () => {
    let resolveMutation!: (result: unknown) => void;
    let capturedRequest!: OrganizationMutationRequest;
    let capturedContext!: OrganizationMutationContext;
    const mutateOrganization = vi.fn(
      (
        request: OrganizationMutationRequest,
        context: OrganizationMutationContext,
      ) =>
        new Promise((resolve) => {
          capturedRequest = request;
          capturedContext = context;
          resolveMutation = resolve;
        }),
    );
    await renderProvider({ ...binding, mutateOrganization });
    const pending = value.actions.inviteMember("org-1", {
      address: "member-1",
      role: "member",
    });
    await expect(
      value.actions.inviteMember("org-1", {
        address: "member-1",
        role: "member",
      }),
    ).rejects.toMatchObject({ code: "submission_in_progress" });
    await expect(
      value.actions.removeMember("org-1", "member-1"),
    ).rejects.toMatchObject({ code: "submission_in_progress" });
    await vi.waitFor(() => expect(mutateOrganization).toHaveBeenCalledOnce());
    await act(async () => {
      resolveMutation(
        committed(capturedRequest, capturedContext, { members: [] }),
      );
      await pending;
    });
  });

  it("aborts active work and rejects retained callbacks after authority changes", async () => {
    let signal: AbortSignal | undefined;
    const first: OrganizationMutationAdapter = {
      ...binding,
      mutateOrganization: vi.fn(
        (_request, context) =>
          new Promise((_resolve, reject) => {
            signal = context.signal;
            context.signal.addEventListener(
              "abort",
              () => reject(new Error("aborted")),
              { once: true },
            );
          }),
      ),
    };
    await renderProvider(first);
    const retainedCreate = value.actions.createOrganization;
    const pending = retainedCreate({ name: "Pending" }).catch((error) => error);
    await vi.waitFor(() =>
      expect(first.mutateOrganization).toHaveBeenCalledOnce(),
    );
    const second: OrganizationMutationAdapter = {
      ...binding,
      mutateOrganization: vi.fn(),
    };
    await renderProvider(second);
    expect(signal?.aborted).toBe(true);
    await expect(pending).resolves.toBeInstanceOf(Error);
    await expect(retainedCreate({ name: "Stale" })).rejects.toMatchObject({
      code: "feature_unavailable",
    });
    expect(first.mutateOrganization).toHaveBeenCalledOnce();
    expect(value.state.organizations).toEqual([]);
  });

  it("rejects late noncooperative unmount results and clones nested input", async () => {
    let resolveMutation!: (result: unknown) => void;
    let capturedRequest!: OrganizationMutationRequest;
    let capturedContext!: OrganizationMutationContext;
    const mutateOrganization = vi.fn(
      (
        request: OrganizationMutationRequest,
        context: OrganizationMutationContext,
      ) =>
        new Promise((resolve) => {
          capturedRequest = request;
          capturedContext = context;
          resolveMutation = resolve;
        }),
    );
    await renderProvider({ ...binding, mutateOrganization });
    const input = {
      name: "Immutable",
      initialMembers: [{ address: "member-original", role: "viewer" as const }],
    };
    const pending = value.actions
      .createOrganization(input)
      .catch((error) => error);
    input.initialMembers[0].address = "member-mutated";
    await vi.waitFor(() => expect(mutateOrganization).toHaveBeenCalledOnce());
    expect(capturedRequest.action).toBe("create");
    if (capturedRequest.action !== "create")
      throw new Error("unexpected action");
    expect(capturedRequest.organization.initialMembers?.[0].address).toBe(
      "member-original",
    );
    expect(Object.isFrozen(capturedRequest.organization.initialMembers)).toBe(
      true,
    );
    expect(
      Object.isFrozen(capturedRequest.organization.initialMembers?.[0]),
    ).toBe(true);
    await act(async () => root.unmount());
    mounted = false;
    resolveMutation(
      committed(capturedRequest, capturedContext, {
        organization: organization("late-org"),
      }),
    );
    await expect(pending).resolves.toMatchObject({
      code: "submission_cancelled",
    });
  });

  it("uses deterministic canonical digests and normalized identities", async () => {
    const first = buildOrganizationMutationRequest(
      "create",
      { chainId: " virtengine-1 ", accountAddress: " virtengine1admin " },
      {
        name: " Example ",
        description: "description",
        initialMembers: [{ address: " member-1 ", role: "viewer" }],
      },
    );
    const second = buildOrganizationMutationRequest(
      "create",
      { accountAddress: "virtengine1admin", chainId: "virtengine-1" },
      {
        initialMembers: [{ role: "viewer", address: "member-1" }],
        description: "description",
        name: "Example",
      },
    );
    expect(first).toEqual(second);
    await expect(digestOrganizationMutationRequest(first)).resolves.toBe(
      await digestOrganizationMutationRequest(second),
    );
  });

  it("ignores stale organization queries after account authority changes", async () => {
    let resolveGroups!: (value: unknown) => void;
    const firstQuery = {
      query: vi.fn(() => new Promise((resolve) => (resolveGroups = resolve))),
    } as unknown as QueryClient;
    await renderProvider(
      undefined,
      binding.chainId,
      binding.accountAddress,
      firstQuery,
    );
    await vi.waitFor(() => expect(firstQuery.query).toHaveBeenCalled());

    const secondQuery = {
      query: vi.fn().mockResolvedValue({ groups: [] }),
    } as unknown as QueryClient;
    await renderProvider(
      undefined,
      binding.chainId,
      "virtengine1other",
      secondQuery,
    );
    await act(async () =>
      resolveGroups({
        groups: [
          {
            id: "stale-org",
            admin: binding.accountAddress,
            total_weight: "1",
            created_at: "2026-08-04T00:00:00.000Z",
            metadata: JSON.stringify({ name: "Stale" }),
          },
        ],
      }),
    );
    expect(value.state.organizations).toEqual([]);
  });

  it("ignores stale organization queries after chain authority changes", async () => {
    let resolveGroups!: (value: unknown) => void;
    const firstQuery = {
      query: vi.fn(() => new Promise((resolve) => (resolveGroups = resolve))),
    } as unknown as QueryClient;
    await renderProvider(
      undefined,
      binding.chainId,
      binding.accountAddress,
      firstQuery,
    );
    await vi.waitFor(() => expect(firstQuery.query).toHaveBeenCalled());

    const secondQuery = {
      query: vi.fn().mockResolvedValue({ groups: [] }),
    } as unknown as QueryClient;
    await renderProvider(
      undefined,
      "virtengine-testnet-1",
      binding.accountAddress,
      secondQuery,
    );
    await act(async () =>
      resolveGroups({
        groups: [
          {
            id: "stale-chain-org",
            admin: binding.accountAddress,
            total_weight: "1",
            created_at: "2026-08-04T00:00:00.000Z",
            metadata: JSON.stringify({ name: "Stale chain" }),
          },
        ],
      }),
    );
    expect(value.state.organizations).toEqual([]);
  });

  it("does not project committed members into another organization detail", async () => {
    await renderProvider({
      ...binding,
      mutateOrganization: vi.fn((request, context) =>
        Promise.resolve(
          committed(request, context, { members: [member("member-org-1")] }),
        ),
      ),
    });
    act(() => value.actions.selectOrganization("org-2"));
    await act(async () => {
      await value.actions.inviteMember("org-1", {
        address: "member-org-1",
        role: "member",
      });
    });
    expect(value.detail.members).toEqual([]);
  });

  it("clears selected organization and detail after committed leave", async () => {
    const adapter: OrganizationMutationAdapter = {
      ...binding,
      mutateOrganization: vi.fn((request, context) =>
        request.action === "create"
          ? Promise.resolve(
              committed(request, context, {
                organization: organization("org-1"),
              }),
            )
          : Promise.resolve(
              committed(request, context, { organizationId: "org-1" }),
            ),
      ),
    };
    await renderProvider(adapter);
    await act(async () => {
      await value.actions.createOrganization({ name: "One" });
    });
    act(() => value.actions.selectOrganization("org-1"));
    await act(async () => {
      await value.actions.leaveOrganization("org-1");
    });
    expect(value.state.selectedOrgId).toBeNull();
    expect(value.detail.organization).toBeNull();
    expect(value.detail.members).toEqual([]);
  });
});

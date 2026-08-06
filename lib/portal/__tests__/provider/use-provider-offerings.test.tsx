import * as React from "react";
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  ProviderOfferingMutationError,
  buildProviderOfferingMutationRequest,
  digestProviderOfferingRequest,
  validateCommittedProviderOfferingMutation,
  type ProviderOfferingMutationAdapter,
} from "../../components/provider/offering-mutation";
import { ProviderProvider, useProvider } from "../../hooks/useProvider";
import type { QueryClient } from "../../types/chain";
import type { OfferingDraft, ProviderOffering } from "../../types/provider";

const binding = {
  chainId: "virtengine-1",
  accountAddress: "virtengine1provider",
} as const;
const draft: OfferingDraft = {
  title: "Authoritative compute",
  description: "Compute capacity",
  type: "compute",
  region: "us-east",
  resources: { cpuCores: 8, memoryGB: 32, storageGB: 100 },
  pricing: {
    basePrice: "10",
    unit: "per_hour",
    denom: "uve",
    depositMultiplier: 1,
    minDurationSeconds: 3600,
  },
  capacity: { totalUnits: 10, availableUnits: 10, maxConcurrentOrders: 5 },
  identityRequirements: { minScore: 0, requiredScopes: [], mfaRequired: false },
  autoPublish: false,
};

const offering = (
  overrides: Partial<ProviderOffering> = {},
): ProviderOffering => ({
  id: "offering-authoritative-1",
  title: draft.title,
  type: draft.type,
  status: "draft",
  activeOrders: 0,
  totalOrders: 0,
  capacityUtilization: 0,
  totalRevenue: "0",
  createdAt: 1_775_000_000_000,
  updatedAt: 1_775_000_000_000,
  ...overrides,
});

const committedAdapter = (): ProviderOfferingMutationAdapter => ({
  ...binding,
  mutateOffering: vi.fn((request, context) =>
    Promise.resolve({
      status: "committed",
      code: 0,
      txHash: "ABC123",
      blockHeight: 42,
      operationId: `operation-${request.action}`,
      requestDigest: context.requestDigest,
      idempotencyKey: context.idempotencyKey,
      request,
      offering: offering({
        id: request.offeringId ?? "offering-authoritative-1",
        status:
          request.action === "publish"
            ? "active"
            : request.action === "pause"
              ? "paused"
              : "draft",
      }),
    }),
  ),
});

describe("provider offering mutation authority", () => {
  let container: HTMLDivElement;
  let root: Root;
  let mounted: boolean;
  let provider: ReturnType<typeof useProvider>;

  const Consumer = () => {
    provider = useProvider();
    return null;
  };

  const renderProvider = async (adapter?: ProviderOfferingMutationAdapter) => {
    await act(async () =>
      root.render(
        <ProviderProvider
          queryClient={
            {
              queryProvider: vi.fn().mockResolvedValue(null),
            } as unknown as QueryClient
          }
          chainId={binding.chainId}
          accountAddress={binding.accountAddress}
          offeringMutationAdapter={adapter}
        >
          <Consumer />
        </ProviderProvider>,
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

  it("fails closed without exact mutation authority", async () => {
    await renderProvider();
    await expect(provider.actions.createOffering(draft)).rejects.toMatchObject({
      code: "feature_unavailable",
    });
    expect(provider.state.offerings).toEqual([]);
  });

  it("creates only from exact committed evidence and server identity", async () => {
    const adapter = committedAdapter();
    await renderProvider(adapter);
    let result!: ProviderOffering;
    await act(async () => {
      result = await provider.actions.createOffering(draft);
    });
    expect(result.id).toBe("offering-authoritative-1");
    expect(Object.isFrozen(result)).toBe(true);
    expect(provider.state.offerings).toEqual([result]);
    expect(adapter.mutateOffering).toHaveBeenCalledTimes(1);
  });

  it("rejects malformed evidence without local state change", async () => {
    await renderProvider({
      ...binding,
      mutateOffering: vi
        .fn()
        .mockResolvedValue({ status: "committed", code: 0 }),
    });
    await expect(provider.actions.createOffering(draft)).rejects.toBeInstanceOf(
      ProviderOfferingMutationError,
    );
    expect(provider.state.offerings).toEqual([]);
  });

  it("publishes and pauses only from exact committed status", async () => {
    const adapter = committedAdapter();
    await renderProvider(adapter);
    await act(async () => {
      await provider.actions.createOffering(draft);
      await provider.actions.publishOffering("offering-authoritative-1");
    });
    expect(provider.state.offerings[0].status).toBe("active");
    await act(async () => {
      await provider.actions.pauseOffering("offering-authoritative-1");
    });
    expect(provider.state.offerings[0].status).toBe("paused");
  });

  it("blocks duplicate submissions for the same operation", async () => {
    let resolveMutation!: (value: unknown) => void;
    let requestValue:
      | Parameters<ProviderOfferingMutationAdapter["mutateOffering"]>[0]
      | undefined;
    let contextValue:
      | Parameters<ProviderOfferingMutationAdapter["mutateOffering"]>[1]
      | undefined;
    const mutateOffering = vi.fn(
      (request, context) =>
        new Promise((resolve) => {
          requestValue = request;
          contextValue = context;
          resolveMutation = resolve;
        }),
    );
    await renderProvider({ ...binding, mutateOffering });
    const pending = provider.actions.createOffering(draft);
    await expect(provider.actions.createOffering(draft)).rejects.toMatchObject({
      code: "submission_in_progress",
    });
    await vi.waitFor(() => expect(mutateOffering).toHaveBeenCalledTimes(1));
    await act(async () => {
      resolveMutation({
        status: "committed",
        code: 0,
        txHash: "ABC123",
        blockHeight: 42,
        operationId: "operation-create",
        requestDigest: contextValue!.requestDigest,
        idempotencyKey: contextValue!.idempotencyKey,
        request: requestValue,
        offering: offering(),
      });
      await pending;
    });
    expect(mutateOffering).toHaveBeenCalledTimes(1);
  });

  it("rejects retained callbacks after adapter authority changes", async () => {
    const first = committedAdapter();
    await renderProvider(first);
    const staleCreate = provider.actions.createOffering;
    await renderProvider(committedAdapter());

    await expect(staleCreate(draft)).rejects.toBeInstanceOf(
      ProviderOfferingMutationError,
    );
    expect(first.mutateOffering).not.toHaveBeenCalled();
  });

  it("deep-freezes nested request data before adapter dispatch", async () => {
    const mutateOffering = vi.fn((request, context) => {
      expect(Object.isFrozen(request)).toBe(true);
      expect(Object.isFrozen(request.draft)).toBe(true);
      expect(
        Object.isFrozen(
          request.draft && "resources" in request.draft
            ? request.draft.resources
            : null,
        ),
      ).toBe(true);
      return Promise.resolve({
        status: "committed",
        code: 0,
        txHash: "ABC123",
        blockHeight: 42,
        operationId: "operation-create",
        requestDigest: context.requestDigest,
        idempotencyKey: context.idempotencyKey,
        request,
        offering: offering(),
      });
    });
    await renderProvider({ ...binding, mutateOffering });
    await act(async () => {
      await provider.actions.createOffering(draft);
    });
    expect(mutateOffering).toHaveBeenCalledOnce();
  });

  it("aborts an active adapter when authority changes", async () => {
    let signal: AbortSignal | undefined;
    const mutateOffering = vi.fn(
      (_request, context) =>
        new Promise((_resolve, reject) => {
          signal = context.signal;
          context.signal.addEventListener(
            "abort",
            () => reject(new Error("aborted")),
            {
              once: true,
            },
          );
        }),
    );
    await renderProvider({ ...binding, mutateOffering });
    const pending = provider.actions
      .createOffering(draft)
      .catch((error) => error);
    await vi.waitFor(() => expect(mutateOffering).toHaveBeenCalledOnce());
    await renderProvider(committedAdapter());

    expect(signal?.aborted).toBe(true);
    await expect(pending).resolves.toBeInstanceOf(Error);
    expect(provider.state.offerings).toEqual([]);
  });

  it("rejects retained and late mutation results after unmount", async () => {
    let resolveMutation!: (value: unknown) => void;
    let requestValue:
      | Parameters<ProviderOfferingMutationAdapter["mutateOffering"]>[0]
      | undefined;
    let contextValue:
      | Parameters<ProviderOfferingMutationAdapter["mutateOffering"]>[1]
      | undefined;
    const mutateOffering = vi.fn(
      (request, context) =>
        new Promise((resolve) => {
          requestValue = request;
          contextValue = context;
          resolveMutation = resolve;
        }),
    );
    await renderProvider({ ...binding, mutateOffering });
    const staleCreate = provider.actions.createOffering;
    const pending = staleCreate(draft).catch((error) => error);
    await vi.waitFor(() => expect(mutateOffering).toHaveBeenCalledOnce());
    await act(async () => root.unmount());
    mounted = false;

    await expect(staleCreate(draft)).rejects.toBeInstanceOf(
      ProviderOfferingMutationError,
    );
    resolveMutation({
      status: "committed",
      code: 0,
      txHash: "ABC123",
      blockHeight: 42,
      operationId: "operation-create",
      requestDigest: contextValue!.requestDigest,
      idempotencyKey: contextValue!.idempotencyKey,
      request: requestValue,
      offering: offering(),
    });
    await expect(pending).resolves.toBeInstanceOf(
      ProviderOfferingMutationError,
    );
  });

  it("serializes conflicting operations for one offering", async () => {
    const adapter = committedAdapter();
    await renderProvider(adapter);
    await act(async () => {
      await provider.actions.createOffering(draft);
    });
    let resolveUpdate!: (value: unknown) => void;
    vi.mocked(adapter.mutateOffering).mockClear();
    vi.mocked(adapter.mutateOffering).mockImplementationOnce(
      () => new Promise((resolve) => (resolveUpdate = resolve)),
    );
    const pending = provider.actions
      .updateOffering("offering-authoritative-1", { title: "Updated" })
      .catch((error) => error);
    await expect(
      provider.actions.publishOffering(" offering-authoritative-1 "),
    ).rejects.toMatchObject({ code: "submission_in_progress" });
    await vi.waitFor(() =>
      expect(adapter.mutateOffering).toHaveBeenCalledOnce(),
    );
    await act(async () => resolveUpdate({}));
    await expect(pending).resolves.toBeInstanceOf(
      ProviderOfferingMutationError,
    );
  });

  it("upserts repeated committed create evidence by authoritative ID", async () => {
    await renderProvider(committedAdapter());
    await act(async () => {
      await provider.actions.createOffering(draft);
      await provider.actions.createOffering({
        ...draft,
        description: "Repeated",
      });
    });
    expect(provider.state.offerings).toHaveLength(1);
    expect(provider.state.offerings[0].id).toBe("offering-authoritative-1");
  });

  it("rejects invalid action shapes and create status mismatches", async () => {
    expect(() =>
      buildProviderOfferingMutationRequest("update", binding, "offering-1"),
    ).toThrow(ProviderOfferingMutationError);
    expect(() =>
      buildProviderOfferingMutationRequest("publish", binding, "offering-1", {
        title: "unexpected",
      }),
    ).toThrow(ProviderOfferingMutationError);

    const request = buildProviderOfferingMutationRequest(
      "create",
      binding,
      undefined,
      draft,
    );
    const digest = await digestProviderOfferingRequest(request);
    expect(() =>
      validateCommittedProviderOfferingMutation(
        {
          status: "committed",
          code: 0,
          txHash: "ABC123",
          blockHeight: 42,
          operationId: "operation-create",
          requestDigest: digest,
          idempotencyKey: digest,
          request,
          offering: offering({ status: "active" }),
        },
        request,
        digest,
      ),
    ).toThrow(ProviderOfferingMutationError);
  });
});

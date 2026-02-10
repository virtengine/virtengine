import { describe, it, expect } from "vitest";
import { buildChatContext } from "../../src/chat/context";
import type { QueryClient } from "../../types/chain";

const mockQueryClient: QueryClient = {
  queryAccount: async () => ({
    address: "virtengine1abc",
    publicKey: null,
    accountNumber: 1,
    sequence: 1,
  }),
  queryBalance: async (_address: string, denom: string) => ({
    denom,
    amount: "2500000",
  }),
  queryIdentity: async () => ({
    address: "virtengine1abc",
    status: "verified",
    score: 92,
    modelVersion: "v1",
    updatedAt: Date.now(),
    blockHeight: 123,
  }),
  queryOffering: async () => ({
    id: "offering",
    providerAddress: "provider",
    status: "active",
    metadata: {},
    createdAt: Date.now(),
  }),
  queryOrder: async () => ({
    id: "order",
    offeringId: "offering",
    customerAddress: "virtengine1abc",
    providerAddress: "provider",
    state: "open",
    createdAt: Date.now(),
  }),
  queryJob: async () => ({
    id: "job",
    customerAddress: "virtengine1abc",
    providerAddress: "provider",
    status: "queued",
    createdAt: Date.now(),
  }),
  queryProvider: async () => ({
    address: "provider",
    status: "active",
    reliabilityScore: 95,
    registeredAt: Date.now(),
  }),
  query: async (path: string) => {
    if (path.includes("leases")) {
      return { leases: [{ lease_id: "lease-1", state: "active" }] };
    }
    if (path.includes("orders")) {
      return { orders: [{ order_id: "order-1", state: "open" }] };
    }
    return {};
  },
};

describe("buildChatContext", () => {
  it("should build context with balance, identity, and leases", async () => {
    const context = await buildChatContext({
      walletAddress: "virtengine1abc",
      chainId: "virtengine-1",
      queryClient: mockQueryClient,
      tokenDenom: "uve",
      roles: ["customer"],
      permissions: ["marketplace:order:create"],
    });

    expect(context.walletAddress).toBe("virtengine1abc");
    expect(context.balance?.denom).toBe("uve");
    expect(context.veid?.status).toBe("verified");
    expect(context.activeLeases?.[0].id).toBe("lease-1");
    expect(context.activeOrders?.[0].id).toBe("order-1");
  });
});

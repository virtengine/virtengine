import { describe, it, expect, vi } from "vitest";
import type { ChainQueryClient } from "../../src/chain/client";
import {
  fetchMarketOfferings,
  fetchMarketOffering,
  fetchMarketOrders,
  fetchMarketOrder,
  fetchMarketBids,
  fetchMarketBid,
  fetchMarketLeases,
  fetchMarketLease,
} from "../../src/chain/queries/market";
import {
  fetchEscrowAccounts,
  fetchEscrowPayments,
  fetchEscrowSettlements,
} from "../../src/chain/queries/escrow";
import {
  fetchVeidRecord,
  fetchVeidScore,
  fetchVeidScopes,
} from "../../src/chain/queries/veid";
import {
  fetchGovProposals,
  fetchGovProposal,
  fetchGovVotes,
  fetchGovParams,
} from "../../src/chain/queries/governance";
import {
  fetchStakingValidators,
  fetchStakingDelegations,
} from "../../src/chain/queries/staking";
import {
  fetchProviders,
  fetchProviderRegistration,
  fetchProviderStatus,
} from "../../src/chain/queries/provider";

describe("chain query helpers", () => {
  it("fetchMarketOfferings passes pagination params", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { offerings: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchMarketOfferings(client, {
      pagination: { limit: 10, key: "next" },
    });

    const [paths, params] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(paths).toContain("/virtengine/market/v1/offerings");
    expect(params["pagination.limit"]).toBe("10");
    expect(params["pagination.key"]).toBe("next");
  });

  it("fetchMarketOffering requests detail paths", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { offering: { id: "offering" } },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchMarketOffering(client, "provider/1");
    const [paths] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(paths).toContain("/virtengine/market/v1/offerings/provider/1");
  });

  it("fetchMarketOrders builds filter params", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { orders: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchMarketOrders(
      client,
      { owner: "ve1", state: "open" },
      { pagination: { limit: 5 } },
    );
    const [, params] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(params["filters.owner"]).toBe("ve1");
    expect(params["filters.state"]).toBe("open");
    expect(params["pagination.limit"]).toBe("5");
  });

  it("fetchMarketOrder builds id params", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { order: { id: {} } },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchMarketOrder(client, {
      owner: "ve1",
      dseq: 1,
      gseq: 2,
      oseq: 3,
    });
    const [, params] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(params["id.owner"]).toBe("ve1");
    expect(params["id.dseq"]).toBe("1");
  });

  it("fetchMarketBids builds filter params", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { bids: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchMarketBids(client, { provider: "ve1", bseq: 9 });
    const [, params] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(params["filters.provider"]).toBe("ve1");
    expect(params["filters.bseq"]).toBe("9");
  });

  it("fetchMarketBid builds id params", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { bid: {} },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchMarketBid(client, {
      owner: "ve1",
      dseq: 1,
      gseq: 2,
      oseq: 3,
      provider: "ve2",
      bseq: 4,
    });
    const [, params] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(params["id.provider"]).toBe("ve2");
    expect(params["id.bseq"]).toBe("4");
  });

  it("fetchMarketLeases builds filter params", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { leases: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchMarketLeases(client, { state: "active" });
    const [, params] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(params["filters.state"]).toBe("active");
  });

  it("fetchMarketLease builds id params", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { lease: {} },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchMarketLease(client, {
      owner: "ve1",
      dseq: 1,
      gseq: 2,
      oseq: 3,
      provider: "ve2",
      bseq: 4,
    });
    const [, params] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(params["id.oseq"]).toBe("3");
  });

  it("fetchEscrowAccounts builds filter params", async () => {
    const client = {
      getJson: vi.fn().mockResolvedValue({
        data: { accounts: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchEscrowAccounts(
      client,
      { state: "open" },
      { pagination: { limit: 3 } },
    );
    const [, params] = (client.getJson as unknown as ReturnType<typeof vi.fn>)
      .mock.calls[0];
    expect(params.state).toBe("open");
    expect(params["pagination.limit"]).toBe("3");
  });

  it("fetchEscrowPayments builds filter params", async () => {
    const client = {
      getJson: vi.fn().mockResolvedValue({
        data: { payments: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchEscrowPayments(client, { xid: "scope-1" });
    const [, params] = (client.getJson as unknown as ReturnType<typeof vi.fn>)
      .mock.calls[0];
    expect(params.xid).toBe("scope-1");
  });

  it("fetchEscrowSettlements targets settlement endpoint", async () => {
    const client = {
      getJson: vi.fn().mockResolvedValue({
        data: { settlements: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchEscrowSettlements(client, "order-1");
    const [path] = (client.getJson as unknown as ReturnType<typeof vi.fn>).mock
      .calls[0];
    expect(path).toContain(
      "/virtengine/settlement/v1/settlements/by-order/order-1",
    );
  });

  it("fetchVeidRecord uses identity record fallback paths", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { record: { status: "verified" } },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchVeidRecord(client, "ve1");
    const [paths] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(paths).toContain("/virtengine/veid/v1/identity_record/ve1");
  });

  it("fetchVeidScore uses score endpoint", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { score: { score: "10" } },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchVeidScore(client, "ve1");
    const [paths] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(paths[0]).toContain("/virtengine/veid/v1/score/ve1");
  });

  it("fetchVeidScopes selects scopes by type when provided", async () => {
    const client = {
      getJson: vi.fn().mockResolvedValue({
        data: { scopes: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchVeidScopes(client, "ve1", "KYC");
    const [path] = (client.getJson as unknown as ReturnType<typeof vi.fn>).mock
      .calls[0];
    expect(path).toContain("/virtengine/veid/v1/scopes/ve1/type/KYC");
  });

  it("fetchGovProposals adds status filter", async () => {
    const client = {
      getJson: vi.fn().mockResolvedValue({
        data: { proposals: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchGovProposals(client, "PROPOSAL_STATUS_VOTING_PERIOD");
    const [, params] = (client.getJson as unknown as ReturnType<typeof vi.fn>)
      .mock.calls[0];
    expect(params.proposal_status).toBe("PROPOSAL_STATUS_VOTING_PERIOD");
  });

  it("fetchGovProposal uses proposal path", async () => {
    const client = {
      getJson: vi.fn().mockResolvedValue({
        data: { proposal: { id: "1" } },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchGovProposal(client, "12");
    const [path] = (client.getJson as unknown as ReturnType<typeof vi.fn>).mock
      .calls[0];
    expect(path).toContain("/cosmos/gov/v1/proposals/12");
  });

  it("fetchGovVotes uses votes path", async () => {
    const client = {
      getJson: vi.fn().mockResolvedValue({
        data: { votes: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchGovVotes(client, "5");
    const [path] = (client.getJson as unknown as ReturnType<typeof vi.fn>).mock
      .calls[0];
    expect(path).toContain("/cosmos/gov/v1/proposals/5/votes");
  });

  it("fetchGovParams targets params path", async () => {
    const client = {
      getJson: vi.fn().mockResolvedValue({
        data: { params: {} },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchGovParams(client, "deposit");
    const [path] = (client.getJson as unknown as ReturnType<typeof vi.fn>).mock
      .calls[0];
    expect(path).toContain("/cosmos/gov/v1/params/deposit");
  });

  it("fetchStakingValidators adds status filter", async () => {
    const client = {
      getJson: vi.fn().mockResolvedValue({
        data: { validators: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchStakingValidators(client, "BOND_STATUS_BONDED");
    const [, params] = (client.getJson as unknown as ReturnType<typeof vi.fn>)
      .mock.calls[0];
    expect(params.status).toBe("BOND_STATUS_BONDED");
  });

  it("fetchStakingDelegations uses delegations path", async () => {
    const client = {
      getJson: vi.fn().mockResolvedValue({
        data: { delegation_responses: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchStakingDelegations(client, "ve1");
    const [path] = (client.getJson as unknown as ReturnType<typeof vi.fn>).mock
      .calls[0];
    expect(path).toContain("/cosmos/staking/v1beta1/delegations/ve1");
  });

  it("fetchProviders hits providers list", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { providers: [] },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchProviders(client, { pagination: { limit: 2 } });
    const [paths] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(paths).toContain("/virtengine/provider/v1beta4/providers");
  });

  it("fetchProviderRegistration hits provider detail", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { provider: { owner: "ve1" } },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchProviderRegistration(client, "ve1");
    const [paths] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(paths).toContain("/virtengine/provider/v1beta4/providers/ve1");
  });

  it("fetchProviderStatus includes status endpoints", async () => {
    const client = {
      getJsonWithPathFallback: vi.fn().mockResolvedValue({
        data: { status: "active" },
        endpoint: "http://rest",
        status: 200,
      }),
    } as unknown as ChainQueryClient;

    await fetchProviderStatus(client, "ve1");
    const [paths] = (
      client.getJsonWithPathFallback as unknown as ReturnType<typeof vi.fn>
    ).mock.calls[0];
    expect(paths.some((path: string) => path.includes("/status"))).toBe(true);
  });
});

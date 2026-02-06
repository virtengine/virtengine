import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type { BidID } from "../generated/protos/virtengine/market/v1/bid.ts";
import type { LeaseID } from "../generated/protos/virtengine/market/v1/lease.ts";
import type { OrderID } from "../generated/protos/virtengine/market/v1/order.ts";
import type { MsgCloseBid, MsgCreateBid } from "../generated/protos/virtengine/market/v1beta5/bidmsg.ts";
import type { MsgCloseLease } from "../generated/protos/virtengine/market/v1beta5/leasemsg.ts";
import type { MarketClientDeps } from "./MarketClient.ts";
import { MarketClient } from "./MarketClient.ts";

type MockFn = (...args: unknown[]) => Promise<unknown>;

const txResponse = () => ({
  height: 1,
  transactionHash: "TXHASH",
  code: 0,
  rawLog: "",
  gasWanted: 100,
  gasUsed: 90,
  data: new Uint8Array(),
  events: [],
  eventsRaw: [],
  msgResponses: [],
});

describe("MarketClient", () => {
  let client: MarketClient;
  let deps: MarketClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          market: {
            v1beta5: {
              getOrder: jest.fn<MockFn>().mockResolvedValue({ order: { id: "order-1" } }),
              getOrders: jest.fn<MockFn>().mockResolvedValue({ orders: [{ id: "order-1" }] }),
              getBid: jest.fn<MockFn>().mockResolvedValue({ bid: { id: "bid-1" } }),
              getBids: jest.fn<MockFn>().mockResolvedValue({ bids: [{ bid: { id: "bid-1" } }] }),
              getLease: jest.fn<MockFn>().mockResolvedValue({ lease: { id: "lease-1" } }),
              getLeases: jest.fn<MockFn>().mockResolvedValue({ leases: [{ lease: { id: "lease-1" } }] }),
              createBid: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              closeBid: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              closeLease: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as MarketClientDeps["sdk"],
    };
    client = new MarketClient(deps);
  });

  it("fetches an order", async () => {
    const order = await client.getOrder({} as OrderID);
    expect(order).toBeTruthy();
    expect(deps.sdk.virtengine.market.v1beta5.getOrder).toHaveBeenCalled();
  });

  it("lists orders", async () => {
    const orders = await client.listOrders({ owner: "virt1" });
    expect(orders).toHaveLength(1);
  });

  it("fetches a bid", async () => {
    const bid = await client.getBid({} as BidID);
    expect(bid).toBeTruthy();
  });

  it("lists bids", async () => {
    const bids = await client.listBids({ provider: "virt1provider" });
    expect(bids).toHaveLength(1);
  });

  it("fetches a lease", async () => {
    const lease = await client.getLease({} as LeaseID);
    expect(lease).toBeTruthy();
  });

  it("lists leases", async () => {
    const leases = await client.listLeases({ owner: "virt1" });
    expect(leases).toHaveLength(1);
  });

  it("creates a bid and returns tx metadata", async () => {
    const result = await client.createBid({} as MsgCreateBid);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("closes a bid and returns tx metadata", async () => {
    const result = await client.closeBid({} as MsgCloseBid);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("closes a lease and returns tx metadata", async () => {
    const result = await client.closeLease({} as MsgCloseLease);
    expect(result.transactionHash).toBe("TXHASH");
  });
});

import { beforeEach, describe, expect, it } from "@jest/globals";

import { MarketClient, type MarketClientDeps } from "./MarketClient.ts";

describe("MarketClient", () => {
  let client: MarketClient;

  beforeEach(() => {
    const deps: MarketClientDeps = {
      sdk: {},
    };
    client = new MarketClient(deps);
  });

  describe("constructor", () => {
    it("should create client instance", () => {
      expect(client).toBeInstanceOf(MarketClient);
    });
  });

  describe("Order queries", () => {
    describe("getOrder", () => {
      it("should throw implementation pending error", async () => {
        await expect(client.getOrder("order-1")).rejects.toThrow(
          "Implementation pending",
        );
      });
    });

    describe("listOrders", () => {
      it("should throw implementation pending error", async () => {
        await expect(client.listOrders()).rejects.toThrow(
          "Implementation pending",
        );
      });

      it("should accept filter options", async () => {
        await expect(
          client.listOrders({ owner: "virt1abc", state: "ORDER_STATE_OPEN" }),
        ).rejects.toThrow("Implementation pending");
      });
    });
  });

  describe("Bid queries", () => {
    describe("getBid", () => {
      it("should throw implementation pending error", async () => {
        await expect(client.getBid("bid-1")).rejects.toThrow(
          "Implementation pending",
        );
      });
    });

    describe("listBids", () => {
      it("should throw implementation pending error", async () => {
        await expect(client.listBids()).rejects.toThrow("Implementation pending");
      });

      it("should accept filter options", async () => {
        await expect(
          client.listBids({ orderId: "order-1", provider: "virt1provider" }),
        ).rejects.toThrow("Implementation pending");
      });
    });
  });

  describe("Lease queries", () => {
    describe("getLease", () => {
      it("should throw implementation pending error", async () => {
        await expect(client.getLease("lease-1")).rejects.toThrow(
          "Implementation pending",
        );
      });
    });

    describe("listLeases", () => {
      it("should throw implementation pending error", async () => {
        await expect(client.listLeases()).rejects.toThrow(
          "Implementation pending",
        );
      });
    });
  });

  describe("Transaction methods", () => {
    describe("createBid", () => {
      it("should throw implementation pending error", async () => {
        await expect(
          client.createBid({
            orderId: "order-1",
            price: { denom: "uvirt", amount: "1000000" },
          }),
        ).rejects.toThrow("Implementation pending");
      });
    });

    describe("closeBid", () => {
      it("should throw implementation pending error", async () => {
        await expect(client.closeBid("bid-1")).rejects.toThrow(
          "Implementation pending",
        );
      });
    });

    describe("closeLease", () => {
      it("should throw implementation pending error", async () => {
        await expect(client.closeLease("lease-1")).rejects.toThrow(
          "Implementation pending",
        );
      });
    });
  });
});

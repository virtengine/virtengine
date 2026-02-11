import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type { MsgActivateEscrow, MsgCreateEscrow, MsgDisputeEscrow, MsgRecordUsage, MsgRefundEscrow, MsgReleaseEscrow, MsgSettleOrder } from "../generated/protos/virtengine/settlement/v1/tx.ts";
import type { SettlementClientDeps } from "./SettlementClient.ts";
import { SettlementClient } from "./SettlementClient.ts";

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

describe("SettlementClient", () => {
  let client: SettlementClient;
  let deps: SettlementClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          settlement: {
            v1: {
              getEscrow: jest.fn<MockFn>().mockResolvedValue({ escrow: { escrowId: "escrow-1" } }),
              getEscrowsByOrder: jest.fn<MockFn>().mockResolvedValue({ escrows: [{ escrowId: "escrow-1" }] }),
              getEscrowsByState: jest.fn<MockFn>().mockResolvedValue({ escrows: [{ escrowId: "escrow-2" }] }),
              getSettlement: jest.fn<MockFn>().mockResolvedValue({ settlement: { settlementId: "settle-1" } }),
              getUsageRecord: jest.fn<MockFn>().mockResolvedValue({ usageRecord: { usageId: "usage-1" } }),
              getUsageRecordsByOrder: jest.fn<MockFn>().mockResolvedValue({ usageRecords: [{ usageId: "usage-1" }] }),
              getPayoutsByProvider: jest.fn<MockFn>().mockResolvedValue({ payouts: [{ payoutId: "payout-1" }] }),
              getClaimableRewards: jest.fn<MockFn>().mockResolvedValue({ rewards: { address: "virt1" } }),
              createEscrow: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              activateEscrow: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              releaseEscrow: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              refundEscrow: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              recordUsage: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              settleOrder: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              disputeEscrow: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
          hpc: {
            v1: {
              getDispute: jest.fn<MockFn>().mockResolvedValue({ dispute: { disputeId: "dispute-1" } }),
            },
          },
        },
      } as unknown as SettlementClientDeps["sdk"],
    };

    client = new SettlementClient(deps);
  });

  it("fetches an escrow", async () => {
    const escrow = await client.getEscrow("escrow-1");
    expect(escrow).toBeTruthy();
  });

  it("lists escrows by order", async () => {
    const escrows = await client.listEscrows({ orderId: "order-1" });
    expect(escrows).toHaveLength(1);
  });

  it("fetches a settlement", async () => {
    const settlement = await client.getSettlement("settle-1");
    expect(settlement).toBeTruthy();
  });

  it("lists payouts", async () => {
    const payouts = await client.listPayouts({ provider: "virt1" });
    expect(payouts).toHaveLength(1);
  });

  it("fetches a usage record", async () => {
    const usage = await client.getUsageRecord("usage-1");
    expect(usage).toBeTruthy();
  });

  it("lists usage records", async () => {
    const usages = await client.listUsageRecords({ orderId: "order-1" });
    expect(usages).toHaveLength(1);
  });

  it("fetches a dispute", async () => {
    const dispute = await client.getDispute("dispute-1");
    expect(dispute).toBeTruthy();
  });

  it("estimates rewards", async () => {
    const rewards = await client.estimateRewards("virt1");
    expect(rewards).toBeTruthy();
  });

  it("creates escrow and returns tx metadata", async () => {
    const result = await client.createEscrow({} as MsgCreateEscrow);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("activates escrow and returns tx metadata", async () => {
    const result = await client.activateEscrow({} as MsgActivateEscrow);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("releases escrow and returns tx metadata", async () => {
    const result = await client.releaseEscrow({} as MsgReleaseEscrow);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("refunds escrow and returns tx metadata", async () => {
    const result = await client.refundEscrow({} as MsgRefundEscrow);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("records usage and returns tx metadata", async () => {
    const result = await client.recordUsage({} as MsgRecordUsage);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("settles order and returns tx metadata", async () => {
    const result = await client.settleOrder({} as MsgSettleOrder);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("opens dispute and returns tx metadata", async () => {
    const result = await client.openDispute({} as MsgDisputeEscrow);
    expect(result.transactionHash).toBe("TXHASH");
  });
});

import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type { EscrowClientDeps } from "./EscrowClient.ts";
import { EscrowClient } from "./EscrowClient.ts";

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

describe("EscrowClient", () => {
  let client: EscrowClient;
  let deps: EscrowClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          escrow: {
            v1: {
              getAccounts: jest.fn<MockFn>().mockResolvedValue({
                accounts: [{ id: { xid: "escrow-1" }, state: "open", balance: { denom: "uvirt", amount: "1000" } }],
              }),
              getPayments: jest.fn<MockFn>().mockResolvedValue({
                payments: [{ paymentId: "pay-1", state: "open", rate: { denom: "uvirt", amount: "10" } }],
              }),
              accountDeposit: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as EscrowClientDeps["sdk"],
    };
    client = new EscrowClient(deps);
  });

  it("should create client instance", () => {
    expect(client).toBeInstanceOf(EscrowClient);
  });

  it("fetches an escrow account by xid", async () => {
    const account = await client.getAccount("escrow-1");
    expect(account).toBeTruthy();
    expect(deps.sdk.virtengine.escrow.v1.getAccounts).toHaveBeenCalledWith({
      xid: "escrow-1",
      state: "",
      pagination: expect.anything(),
    });
  });

  it("returns null when no account found", async () => {
    (deps.sdk.virtengine.escrow.v1.getAccounts as jest.Mock<MockFn>)
      .mockResolvedValueOnce({ accounts: [] });
    const account = await client.getAccount("missing");
    expect(account).toBeNull();
  });

  it("lists escrow accounts", async () => {
    const accounts = await client.listAccounts({ state: "open" });
    expect(accounts).toHaveLength(1);
    expect(deps.sdk.virtengine.escrow.v1.getAccounts).toHaveBeenCalledWith(
      expect.objectContaining({ state: "open" }),
    );
  });

  it("lists escrow accounts with default filters", async () => {
    const accounts = await client.listAccounts();
    expect(accounts).toHaveLength(1);
    expect(deps.sdk.virtengine.escrow.v1.getAccounts).toHaveBeenCalledWith(
      expect.objectContaining({ state: "", xid: "" }),
    );
  });

  it("fetches payments for an escrow account", async () => {
    const payments = await client.getPayments("escrow-1");
    expect(payments).toHaveLength(1);
    expect(deps.sdk.virtengine.escrow.v1.getPayments).toHaveBeenCalledWith(
      expect.objectContaining({ xid: "escrow-1" }),
    );
  });

  it("deposits funds and returns tx metadata", async () => {
    const result = await client.deposit({
      signer: "virt1depositor",
      id: undefined,
      deposit: { amount: { denom: "uvirt", amount: "500" }, sources: [] },
    });
    expect(result.transactionHash).toBe("TXHASH");
    expect(result.code).toBe(0);
  });
});

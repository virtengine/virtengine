import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type { MsgClaimRewards, MsgDelegate, MsgRedelegate, MsgUndelegate } from "../generated/protos/virtengine/delegation/v1/tx.ts";
import type { DelegationClientDeps } from "./DelegationClient.ts";
import { DelegationClient } from "./DelegationClient.ts";

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

describe("DelegationClient", () => {
  let client: DelegationClient;
  let deps: DelegationClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          delegation: {
            v1: {
              getDelegation: jest.fn<MockFn>().mockResolvedValue({ delegation: { delegatorAddress: "virt1" } }),
              getDelegatorDelegations: jest.fn<MockFn>().mockResolvedValue({ delegations: [{ delegatorAddress: "virt1" }] }),
              getValidatorDelegations: jest.fn<MockFn>().mockResolvedValue({ delegations: [{ delegatorAddress: "virt1" }] }),
              getDelegatorRewards: jest.fn<MockFn>().mockResolvedValue({ rewards: [{ validatorAddress: "virtval1" }] }),
              getDelegatorAllRewards: jest.fn<MockFn>().mockResolvedValue({ rewards: [{ validatorAddress: "virtval1" }] }),
              delegate: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              undelegate: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              redelegate: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              claimRewards: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as DelegationClientDeps["sdk"],
    };

    client = new DelegationClient(deps);
  });

  it("fetches a delegation", async () => {
    const delegation = await client.getDelegation("virt1", "virtval1");
    expect(delegation).toBeTruthy();
  });

  it("lists delegations", async () => {
    const delegations = await client.listDelegations({ delegatorAddress: "virt1" });
    expect(delegations).toHaveLength(1);
  });

  it("fetches delegator rewards", async () => {
    const rewards = await client.getDelegatorRewards("virt1", "virtval1");
    expect(rewards).toHaveLength(1);
  });

  it("fetches validator delegators", async () => {
    const delegations = await client.getValidatorDelegators("virtval1");
    expect(delegations).toHaveLength(1);
  });

  it("delegates and returns tx metadata", async () => {
    const result = await client.delegate({} as MsgDelegate);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("undelegates and returns tx metadata", async () => {
    const result = await client.undelegate({} as MsgUndelegate);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("redelegates and returns tx metadata", async () => {
    const result = await client.redelegate({} as MsgRedelegate);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("claims rewards and returns tx metadata", async () => {
    const result = await client.claimRewards({} as MsgClaimRewards);
    expect(result.transactionHash).toBe("TXHASH");
  });
});

import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type { MsgRecordPerformance, MsgSlashValidator, MsgUnjailValidator } from "../generated/protos/virtengine/staking/v1/tx.ts";
import type { VirtStakingClientDeps } from "./VirtStakingClient.ts";
import { VirtStakingClient } from "./VirtStakingClient.ts";

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

describe("VirtStakingClient", () => {
  let client: VirtStakingClient;
  let deps: VirtStakingClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          staking: {
            v1: {
              recordPerformance: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              slashValidator: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              unjailValidator: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as VirtStakingClientDeps["sdk"],
    };

    client = new VirtStakingClient(deps);
  });

  it("records performance", async () => {
    const result = await client.recordPerformance({} as MsgRecordPerformance);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("slashes validator", async () => {
    const result = await client.slashValidator({} as MsgSlashValidator);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("unjails validator", async () => {
    const result = await client.unjailValidator({} as MsgUnjailValidator);
    expect(result.transactionHash).toBe("TXHASH");
  });
});

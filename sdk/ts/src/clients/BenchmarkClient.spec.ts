import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type {
  MsgFlagProvider,
  MsgRequestChallenge,
  MsgResolveAnomalyFlag,
  MsgRespondChallenge,
  MsgSubmitBenchmarks,
  MsgUnflagProvider,
} from "../generated/protos/virtengine/benchmark/v1/tx.ts";
import type { BenchmarkClientDeps } from "./BenchmarkClient.ts";
import { BenchmarkClient } from "./BenchmarkClient.ts";

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

describe("BenchmarkClient", () => {
  let client: BenchmarkClient;
  let deps: BenchmarkClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          benchmark: {
            v1: {
              submitBenchmarks: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              requestChallenge: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              respondChallenge: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              flagProvider: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              unflagProvider: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              resolveAnomalyFlag: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as BenchmarkClientDeps["sdk"],
    };

    client = new BenchmarkClient(deps);
  });

  it("submits benchmarks", async () => {
    const result = await client.submitBenchmarks({} as MsgSubmitBenchmarks);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("requests challenge", async () => {
    const result = await client.requestChallenge({} as MsgRequestChallenge);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("responds to challenge", async () => {
    const result = await client.respondChallenge({} as MsgRespondChallenge);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("flags provider", async () => {
    const result = await client.flagProvider({} as MsgFlagProvider);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("unflags provider", async () => {
    const result = await client.unflagProvider({} as MsgUnflagProvider);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("resolves anomaly flag", async () => {
    const result = await client.resolveAnomalyFlag({} as MsgResolveAnomalyFlag);
    expect(result.transactionHash).toBe("TXHASH");
  });
});

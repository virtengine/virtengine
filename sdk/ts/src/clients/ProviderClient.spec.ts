import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type { MsgCreateProvider, MsgDeleteProvider, MsgUpdateProvider } from "../generated/protos/virtengine/provider/v1beta4/msg.ts";
import type { ProviderClientDeps } from "./ProviderClient.ts";
import { ProviderClient } from "./ProviderClient.ts";

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

describe("ProviderClient", () => {
  let client: ProviderClient;
  let deps: ProviderClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          provider: {
            v1beta4: {
              getProvider: jest.fn<MockFn>().mockResolvedValue({ provider: { owner: "virt1" } }),
              getProviders: jest.fn<MockFn>().mockResolvedValue({ providers: [{ owner: "virt1" }] }),
              createProvider: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              updateProvider: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              deleteProvider: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
          hpc: {
            v1: {
              getClustersByProvider: jest.fn<MockFn>().mockResolvedValue({ clusters: [{ clusterId: "cluster-1" }] }),
              getJobsByProvider: jest.fn<MockFn>().mockResolvedValue({ jobs: [{ jobId: "job-1" }] }),
            },
          },
        },
      } as unknown as ProviderClientDeps["sdk"],
    };

    client = new ProviderClient(deps);
  });

  it("fetches a provider", async () => {
    const provider = await client.getProvider("virt1");
    expect(provider).toBeTruthy();
  });

  it("lists providers", async () => {
    const providers = await client.listProviders();
    expect(providers).toHaveLength(1);
  });

  it("fetches provider capacity", async () => {
    const clusters = await client.getProviderCapacity({ providerAddress: "virt1" });
    expect(clusters).toHaveLength(1);
  });

  it("fetches provider orders", async () => {
    const jobs = await client.getProviderOrders("virt1");
    expect(jobs).toHaveLength(1);
  });

  it("registers provider and returns tx metadata", async () => {
    const result = await client.registerProvider({} as MsgCreateProvider);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("updates provider and returns tx metadata", async () => {
    const result = await client.updateProvider({} as MsgUpdateProvider);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("deactivates provider and returns tx metadata", async () => {
    const result = await client.deactivateProvider({} as MsgDeleteProvider);
    expect(result.transactionHash).toBe("TXHASH");
  });
});

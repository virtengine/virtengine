import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type {
  MsgReactivateApprovedClient,
  MsgRegisterApprovedClient,
  MsgRevokeApprovedClient,
  MsgSuspendApprovedClient,
  MsgUpdateApprovedClient,
} from "../generated/protos/virtengine/config/v1/tx.ts";
import type { ConfigClientDeps } from "./ConfigClient.ts";
import { ConfigClient } from "./ConfigClient.ts";

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

describe("ConfigClient", () => {
  let client: ConfigClient;
  let deps: ConfigClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          config: {
            v1: {
              registerApprovedClient: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              updateApprovedClient: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              suspendApprovedClient: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              revokeApprovedClient: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              reactivateApprovedClient: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as ConfigClientDeps["sdk"],
    };

    client = new ConfigClient(deps);
  });

  it("registers approved client", async () => {
    const result = await client.registerApprovedClient({} as MsgRegisterApprovedClient);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("updates approved client", async () => {
    const result = await client.updateApprovedClient({} as MsgUpdateApprovedClient);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("suspends approved client", async () => {
    const result = await client.suspendApprovedClient({} as MsgSuspendApprovedClient);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("revokes approved client", async () => {
    const result = await client.revokeApprovedClient({} as MsgRevokeApprovedClient);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("reactivates approved client", async () => {
    const result = await client.reactivateApprovedClient({} as MsgReactivateApprovedClient);
    expect(result.transactionHash).toBe("TXHASH");
  });
});

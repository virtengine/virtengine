import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type {
  MsgProposeMeasurement,
  MsgRegisterEnclaveIdentity,
  MsgRevokeMeasurement,
  MsgRotateEnclaveIdentity,
} from "../generated/protos/virtengine/enclave/v1/tx.ts";
import type { EnclaveClientDeps } from "./EnclaveClient.ts";
import { EnclaveClient } from "./EnclaveClient.ts";

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

describe("EnclaveClient", () => {
  let client: EnclaveClient;
  let deps: EnclaveClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          enclave: {
            v1: {
              getEnclaveIdentity: jest.fn<MockFn>().mockResolvedValue({ identity: { validatorAddress: "virtval1" } }),
              getActiveValidatorEnclaveKeys: jest.fn<MockFn>().mockResolvedValue({ identities: [{ validatorAddress: "virtval1" }] }),
              getCommitteeEnclaveKeys: jest.fn<MockFn>().mockResolvedValue({ identities: [{ validatorAddress: "virtval1" }] }),
              getMeasurementAllowlist: jest.fn<MockFn>().mockResolvedValue({ measurements: [{ measurementHash: "hash" }] }),
              getMeasurement: jest.fn<MockFn>().mockResolvedValue({ measurement: { measurementHash: "hash" } }),
              getKeyRotation: jest.fn<MockFn>().mockResolvedValue({ rotation: { validatorAddress: "virtval1" } }),
              getValidKeySet: jest.fn<MockFn>().mockResolvedValue({ validatorKeys: [{ validatorAddress: "virtval1" }] }),
              getAttestedResult: jest.fn<MockFn>().mockResolvedValue({ result: { scopeId: "scope" } }),
              registerEnclaveIdentity: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              rotateEnclaveIdentity: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              proposeMeasurement: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              revokeMeasurement: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as EnclaveClientDeps["sdk"],
    };

    client = new EnclaveClient(deps);
  });

  it("fetches enclave identity", async () => {
    const identity = await client.getEnclaveIdentity("virtval1");
    expect(identity).toBeTruthy();
  });

  it("lists active enclave keys", async () => {
    const keys = await client.listActiveValidatorEnclaveKeys();
    expect(keys).toHaveLength(1);
  });

  it("fetches committee keys", async () => {
    const keys = await client.getCommitteeEnclaveKeys(1);
    expect(keys).toHaveLength(1);
  });

  it("lists measurements", async () => {
    const measurements = await client.listMeasurementAllowlist();
    expect(measurements).toHaveLength(1);
  });

  it("fetches measurement", async () => {
    const measurement = await client.getMeasurement("hash");
    expect(measurement).toBeTruthy();
  });

  it("fetches key rotation", async () => {
    const record = await client.getKeyRotation("virtval1");
    expect(record).toBeTruthy();
  });

  it("fetches valid key set", async () => {
    const keys = await client.getValidKeySet();
    expect(keys).toHaveLength(1);
  });

  it("fetches attested result", async () => {
    const result = await client.getAttestedResult(1, "scope");
    expect(result).toBeTruthy();
  });

  it("registers enclave identity and returns tx metadata", async () => {
    const result = await client.registerEnclaveIdentity({} as MsgRegisterEnclaveIdentity);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("rotates enclave identity and returns tx metadata", async () => {
    const result = await client.rotateEnclaveIdentity({} as MsgRotateEnclaveIdentity);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("proposes measurement and returns tx metadata", async () => {
    const result = await client.proposeMeasurement({} as MsgProposeMeasurement);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("revokes measurement and returns tx metadata", async () => {
    const result = await client.revokeMeasurement({} as MsgRevokeMeasurement);
    expect(result.transactionHash).toBe("TXHASH");
  });
});

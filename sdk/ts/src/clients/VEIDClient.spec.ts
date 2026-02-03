import { beforeEach, describe, expect, it, jest } from "@jest/globals";
import Long from "long";

import { VEIDClient } from "./VEIDClient.ts";
import type { VEIDClientDeps } from "./VEIDClient.ts";
import { IdentityTier, VerificationStatus } from "../generated/protos/virtengine/veid/v1/types.ts";
import type {
  MsgCreateIdentityWallet,
  MsgRequestVerification,
  MsgUploadScope,
} from "../generated/protos/virtengine/veid/v1/tx.ts";

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

describe("VEIDClient", () => {
  let client: VEIDClient;
  let deps: VEIDClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          veid: {
            v1: {
              getIdentity: jest.fn().mockResolvedValue({
                found: true,
                identity: { accountAddress: "virt1", tier: IdentityTier.IDENTITY_TIER_STANDARD },
              }),
              getIdentityScore: jest.fn().mockResolvedValue({
                found: true,
                score: { score: 75, tier: IdentityTier.IDENTITY_TIER_STANDARD },
              }),
              getScopes: jest.fn().mockResolvedValue({
                scopes: [{ scopeId: "scope-1" }],
              }),
              uploadScope: jest.fn().mockImplementation((_input, options) => {
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({
                  scopeId: "scope-1",
                  status: VerificationStatus.VERIFICATION_STATUS_PENDING,
                  uploadedAt: Long.ZERO,
                });
              }),
              requestVerification: jest.fn().mockImplementation((_input, options) => {
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({
                  scopeId: "scope-1",
                  status: VerificationStatus.VERIFICATION_STATUS_IN_PROGRESS,
                  requestedAt: Long.ZERO,
                });
              }),
              createIdentityWallet: jest.fn().mockImplementation((_input, options) => {
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({ walletId: "wallet-1" });
              }),
            },
          },
          hpc: {
            v1: {
              getOffering: jest.fn().mockResolvedValue({
                offering: { offeringId: "offering-1", requiredIdentityThreshold: 60 },
              }),
            },
          },
        },
      } as VEIDClientDeps["sdk"],
    };

    client = new VEIDClient(deps);
  });

  it("fetches identity record", async () => {
    const identity = await client.getIdentity("virt1");
    expect(identity?.accountAddress).toBe("virt1");
  });

  it("fetches identity score", async () => {
    const score = await client.getScore("virt1");
    expect(score?.score).toBe(75);
  });

  it("verifies eligibility using offering threshold", async () => {
    const result = await client.verifyEligibility("virt1", "offering-1");
    expect(result.eligible).toBe(true);
    expect(result.requiredScore).toBe(60);
  });

  it("lists scopes", async () => {
    const scopes = await client.listScopes("virt1");
    expect(scopes).toHaveLength(1);
  });

  it("uploads scope and returns tx metadata", async () => {
    const result = await client.uploadScope({} as MsgUploadScope);
    expect(result.scopeId).toBe("scope-1");
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("requests verification and returns tx metadata", async () => {
    const result = await client.requestVerification({} as MsgRequestVerification);
    expect(result.status).toBe(VerificationStatus.VERIFICATION_STATUS_IN_PROGRESS);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("creates identity wallet and returns tx metadata", async () => {
    const result = await client.createIdentityWallet({} as MsgCreateIdentityWallet);
    expect(result.walletId).toBe("wallet-1");
    expect(result.transactionHash).toBe("TXHASH");
  });
});

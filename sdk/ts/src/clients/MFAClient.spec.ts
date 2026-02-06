import { beforeEach, describe, expect, it, jest } from "@jest/globals";
import Long from "long";

import { FactorEnrollmentStatus, FactorType, SensitiveTransactionType } from "../generated/protos/virtengine/mfa/v1/types.ts";
import { MemoryCache } from "../utils/cache.ts";
import type { MFAClientDeps } from "./MFAClient.ts";
import { MFAClient } from "./MFAClient.ts";

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

describe("MFAClient", () => {
  let client: MFAClient;
  let deps: MFAClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          mfa: {
            v1: {
              getMFAPolicy: jest.fn<MockFn>().mockResolvedValue({
                found: true,
                policy: { requiredFactors: 2, enabledTransactions: [] },
              }),
              getFactorEnrollments: jest.fn<MockFn>().mockResolvedValue({
                enrollments: [
                  { factorId: "factor-1", factorType: FactorType.FACTOR_TYPE_TOTP, status: FactorEnrollmentStatus.ENROLLMENT_STATUS_ACTIVE },
                ],
              }),
              enrollFactor: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({
                  factorId: "factor-new",
                  status: FactorEnrollmentStatus.ENROLLMENT_STATUS_PENDING,
                });
              }),
              createChallenge: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({
                  challengeId: "challenge-1",
                  challengeData: new Uint8Array([1, 2, 3]),
                  expiresAt: Long.fromNumber(9999),
                });
              }),
              verifyChallenge: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({
                  verified: true,
                  sessionId: "session-1",
                  sessionExpiresAt: Long.fromNumber(9999),
                  remainingFactors: [],
                });
              }),
              getChallenge: jest.fn<MockFn>().mockResolvedValue({
                found: true,
                challenge: { challengeId: "challenge-1" },
              }),
              setMFAPolicy: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({ success: true });
              }),
            },
          },
        },
      } as unknown as MFAClientDeps["sdk"],
    };
    client = new MFAClient(deps);
  });

  it("should create client instance", () => {
    expect(client).toBeInstanceOf(MFAClient);
  });

  it("fetches MFA policy for an address", async () => {
    const policy = await client.getPolicy("virt1abc");
    expect(policy).toBeTruthy();
    expect(policy?.requiredFactors).toBe(2);
  });

  it("returns null when no policy found", async () => {
    (deps.sdk.virtengine.mfa.v1.getMFAPolicy as jest.Mock<MockFn>)
      .mockResolvedValueOnce({ found: false });
    const policy = await client.getPolicy("virt1none");
    expect(policy).toBeNull();
  });

  it("caches policy on subsequent calls", async () => {
    const cache = new MemoryCache({ ttlMs: 30000 });
    const client2 = new MFAClient(deps, { enableCaching: true, cache });
    await client2.getPolicy("virt1abc");
    await client2.getPolicy("virt1abc");
    expect(deps.sdk.virtengine.mfa.v1.getMFAPolicy).toHaveBeenCalledTimes(1);
  });

  it("lists factor enrollments", async () => {
    const enrollments = await client.listEnrollments("virt1abc");
    expect(enrollments).toHaveLength(1);
    expect(enrollments[0].factorId).toBe("factor-1");
  });

  it("lists enrollments with filters", async () => {
    await client.listEnrollments("virt1abc", {
      factorTypeFilter: FactorType.FACTOR_TYPE_TOTP,
      statusFilter: FactorEnrollmentStatus.ENROLLMENT_STATUS_ACTIVE,
    });
    expect(deps.sdk.virtengine.mfa.v1.getFactorEnrollments).toHaveBeenCalledWith(
      expect.objectContaining({
        factorTypeFilter: FactorType.FACTOR_TYPE_TOTP,
        statusFilter: FactorEnrollmentStatus.ENROLLMENT_STATUS_ACTIVE,
      }),
    );
  });

  it("enrolls a factor and returns tx metadata", async () => {
    const result = await client.enrollFactor({
      sender: "virt1abc",
      factorType: FactorType.FACTOR_TYPE_TOTP,
      label: "My TOTP",
    });
    expect(result.factorId).toBe("factor-new");
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("creates a challenge and returns tx metadata", async () => {
    const result = await client.createChallenge({
      sender: "virt1abc",
      factorType: FactorType.FACTOR_TYPE_TOTP,
      transactionType: SensitiveTransactionType.SENSITIVE_TX_LARGE_WITHDRAWAL,
    });
    expect(result.challengeId).toBe("challenge-1");
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("verifies a challenge and returns session info", async () => {
    const result = await client.verifyChallenge({
      sender: "virt1abc",
      challengeId: "challenge-1",
      response: {
        challengeId: "challenge-1",
        factorType: FactorType.FACTOR_TYPE_TOTP,
        responseData: new Uint8Array(),
        clientInfo: undefined,
        timestamp: Long.ZERO,
      },
    });
    expect(result.verified).toBe(true);
    expect(result.sessionId).toBe("session-1");
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("fetches a challenge by ID", async () => {
    const challenge = await client.getChallenge("challenge-1");
    expect(challenge?.challengeId).toBe("challenge-1");
  });

  it("returns null for unfound challenge", async () => {
    (deps.sdk.virtengine.mfa.v1.getChallenge as jest.Mock<MockFn>)
      .mockResolvedValueOnce({ found: false });
    const challenge = await client.getChallenge("missing");
    expect(challenge).toBeNull();
  });

  it("sets MFA policy and returns tx metadata", async () => {
    const result = await client.setPolicy({
      sender: "virt1abc",
      policy: {
        accountAddress: "virt1abc",
        requiredFactors: [],
        trustedDeviceRule: undefined,
        recoveryFactors: [],
        keyRotationFactors: [],
        sessionDuration: Long.ZERO,
        veidThreshold: 0,
        enabled: true,
        createdAt: Long.ZERO,
        updatedAt: Long.ZERO,
      },
    });
    expect(result.success).toBe(true);
    expect(result.transactionHash).toBe("TXHASH");
  });
});

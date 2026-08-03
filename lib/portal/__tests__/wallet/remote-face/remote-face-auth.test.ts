import { describe, expect, it, vi } from "vitest";
import {
  RemoteFaceAuthenticationGate,
  type RemoteFaceAuthenticationAuthorities,
  type RemoteFaceAuthenticationRequest,
  type RemoteFaceAuthorizationBinding,
  type RemoteFaceEvidenceProjection,
  type RemoteFaceProfileBinding,
} from "../../../src/wallet/remote-face";

const NOW = 1_800_000_000_000;
const profile: RemoteFaceProfileBinding = {
  t1: { profile: "t1-profile", version: "1", digest: "sha256:t1" },
  t5: { profile: "t5-profile", version: "5", digest: "sha256:t5" },
};
const binding: RemoteFaceAuthorizationBinding = {
  chainId: "virt-1",
  accountId: "account-1",
  sessionId: "session-1",
  deviceId: "device-1",
  actionKind: "transaction",
  actionDigest: "sha256:transaction",
};
const request: RemoteFaceAuthenticationRequest = {
  profile,
  binding,
  nonce: "nonce-1",
  issuedAt: NOW - 1_000,
  expiresAt: NOW + 60_000,
  activeLiveness: { opaque: "active" },
  passiveLiveness: { opaque: "passive" },
  deviceSessionAttestation: { opaque: "attestation" },
  possessionProof: { opaque: "passkey-assertion" },
  recoveryFactors: [{ opaque: "recovery-1" }, { opaque: "recovery-2" }],
  recoveryPolicy: { opaque: "delayed-threshold-policy" },
  remoteFaceResult: { opaque: "supplemental-result" },
};

const evidence = (
  evidenceRef: string,
  overrides: Partial<RemoteFaceEvidenceProjection> = {},
): RemoteFaceEvidenceProjection => ({
  valid: true,
  evidenceRef,
  profile,
  binding,
  validatedAt: NOW - 500,
  expiresAt: NOW + 30_000,
  ...overrides,
});

function harness(overrides: Partial<RemoteFaceAuthenticationAuthorities> = {}) {
  const consumed = new Set<string>();
  const consume = vi.fn(async (nonce: string) => {
    if (consumed.has(nonce)) return false;
    consumed.add(nonce);
    return true;
  });
  const authorizeAndProject = vi.fn(async (input: { nonce: string }) => ({
    authorized: true,
    authorizationRef: "authorization-1",
    nonce: input.nonce,
    profile,
    binding,
    expiresAt: NOW + 20_000,
  }));
  const authorities: RemoteFaceAuthenticationAuthorities = {
    profileAuthority: {
      validateAndProject: vi.fn(async () => ({ enabled: true, profile })),
    },
    activeLivenessValidator: {
      validateAndProject: vi.fn(async () => ({
        ...evidence("active"),
        observedAt: NOW - 1_000,
      })),
    },
    passiveLivenessValidator: {
      validateAndProject: vi.fn(async () => ({
        ...evidence("passive"),
        observedAt: NOW - 1_000,
      })),
    },
    deviceSessionAttestationValidator: {
      validateAndProject: vi.fn(async () => evidence("attestation")),
    },
    possessionValidator: {
      validateAndProject: vi.fn(async () => ({
        ...evidence("passkey"),
        proofType: "passkey-assertion",
      })),
    },
    recoveryFactorValidator: {
      validateAndProject: vi.fn(async (input: unknown) => ({
        valid: true,
        factorClass: "recovery" as const,
        factorRef: (input as { opaque: string }).opaque,
        accountId: binding.accountId,
        validatedAt: NOW - 500,
        expiresAt: NOW + 30_000,
      })),
    },
    recoveryPolicyValidator: {
      validateAndProject: vi.fn(async () => ({
        valid: true,
        accountId: binding.accountId,
        threshold: 2,
        factorRefs: ["recovery-1", "recovery-2"],
        delayMs: 24 * 60 * 60 * 1000,
        expiresAt: NOW + 30_000,
      })),
    },
    nonceGuard: { consume },
    finalAuthorizationAuthority: { authorizeAndProject },
    ...overrides,
  };
  return {
    gate: new RemoteFaceAuthenticationGate(authorities, { now: () => NOW }),
    consume,
    authorizeAndProject,
  };
}

async function expectCode(promise: Promise<unknown>, code: string) {
  await expect(promise).rejects.toMatchObject({
    name: "RemoteFaceAuthenticationError",
    code,
  });
}

describe("RemoteFaceAuthenticationGate", () => {
  it("is unavailable by default and always denies face-only input", async () => {
    await expectCode(
      new RemoteFaceAuthenticationGate().authenticate(request),
      "unavailable",
    );
    const test = harness();
    await expectCode(
      test.gate.authenticate({
        ...request,
        activeLiveness: undefined,
        passiveLiveness: undefined,
        deviceSessionAttestation: undefined,
        possessionProof: undefined,
        recoveryFactors: [undefined, undefined],
        recoveryPolicy: undefined,
      }),
      "invalid_evidence",
    );
    expect(test.authorizeAndProject).not.toHaveBeenCalled();
  });

  it("authorizes a valid combined flow without returning opaque evidence", async () => {
    const test = harness();
    await expect(test.gate.authenticate(request)).resolves.toEqual({
      authorized: true,
      authorizationRef: "authorization-1",
      expiresAt: NOW + 20_000,
    });
    expect(test.consume).toHaveBeenCalledBefore(test.authorizeAndProject);
    expect(
      JSON.stringify(
        await test.gate.authenticate({ ...request, nonce: "nonce-2" }),
      ),
    ).not.toContain("supplemental-result");
  });

  it("rejects disabled and mismatched exact T1/T5 profiles", async () => {
    await expectCode(
      harness({
        profileAuthority: {
          validateAndProject: async () => ({ enabled: false, profile }),
        },
      }).gate.authenticate(request),
      "disabled_profile",
    );
    await expectCode(
      harness({
        profileAuthority: {
          validateAndProject: async () => ({
            enabled: true,
            profile: {
              ...profile,
              t5: { ...profile.t5, digest: "sha256:other" },
            },
          }),
        },
      }).gate.authenticate(request),
      "profile_mismatch",
    );
  });

  it.each([
    ["chainId", "other-chain"],
    ["accountId", "other-account"],
    ["sessionId", "other-session"],
    ["deviceId", "other-device"],
    ["actionKind", "action"],
    ["actionDigest", "sha256:other-action"],
  ] as const)("rejects projected %s drift", async (field, value) => {
    const test = harness({
      deviceSessionAttestationValidator: {
        validateAndProject: async () =>
          evidence("attestation", { binding: { ...binding, [field]: value } }),
      },
    });
    await expectCode(test.gate.authenticate(request), "binding_mismatch");
    expect(test.consume).not.toHaveBeenCalled();
  });

  it("rejects expiry and stale active or passive liveness", async () => {
    await expectCode(
      harness().gate.authenticate({ ...request, expiresAt: NOW }),
      "expired",
    );
    for (const validator of [
      "activeLivenessValidator",
      "passiveLivenessValidator",
    ] as const) {
      await expectCode(
        harness({
          [validator]: {
            validateAndProject: async () => ({
              ...evidence(validator),
              observedAt: NOW - 30_001,
            }),
          },
        }).gate.authenticate(request),
        "stale_liveness",
      );
    }
  });

  it("rejects projected profile mismatch and reused liveness evidence", async () => {
    await expectCode(
      harness({
        activeLivenessValidator: {
          validateAndProject: async () => ({
            ...evidence("active", {
              profile: {
                ...profile,
                t1: { ...profile.t1, version: "other" },
              },
            }),
            observedAt: NOW,
          }),
        },
      }).gate.authenticate(request),
      "profile_mismatch",
    );
    await expectCode(
      harness({
        passiveLivenessValidator: {
          validateAndProject: async () => ({
            ...evidence("active"),
            observedAt: NOW,
          }),
        },
      }).gate.authenticate(request),
      "invalid_evidence",
    );
  });

  it("requires possession and two independent recovery factors with delay", async () => {
    await expectCode(
      harness({
        possessionValidator: {
          validateAndProject: async () => ({
            ...evidence("possession"),
            valid: false,
            proofType: "other-possession-factor",
          }),
        },
      }).gate.authenticate(request),
      "missing_possession",
    );
    await expectCode(
      harness({
        recoveryFactorValidator: {
          validateAndProject: async () => ({
            valid: true,
            factorClass: "recovery",
            factorRef: "same-factor",
            accountId: binding.accountId,
            validatedAt: NOW,
            expiresAt: NOW + 30_000,
          }),
        },
      }).gate.authenticate(request),
      "missing_recovery",
    );
    await expectCode(
      harness({
        recoveryPolicyValidator: {
          validateAndProject: async () => ({
            valid: true,
            accountId: binding.accountId,
            threshold: 2,
            factorRefs: ["recovery-1", "recovery-2"],
            delayMs: 0,
            expiresAt: NOW + 30_000,
          }),
        },
      }).gate.authenticate(request),
      "missing_recovery",
    );
  });

  it("atomically rejects replay and concurrent authorization", async () => {
    const test = harness();
    const [first, second] = await Promise.allSettled([
      test.gate.authenticate(request),
      test.gate.authenticate(request),
    ]);
    expect([first.status, second.status].sort()).toEqual([
      "fulfilled",
      "rejected",
    ]);
    const rejected =
      first.status === "rejected"
        ? first.reason
        : second.status === "rejected"
          ? second.reason
          : undefined;
    expect(rejected).toMatchObject({ code: "replayed_nonce" });
    expect(test.authorizeAndProject).toHaveBeenCalledTimes(1);
    await expectCode(test.gate.authenticate(request), "replayed_nonce");
  });

  it.each([
    "image",
    "face_template",
    "embedding",
    "livenessScore",
    "biometricIdentifier",
  ])("rejects recursive prohibited biometric field %s", async (field) => {
    const test = harness();
    await expectCode(
      test.gate.authenticate({
        ...request,
        remoteFaceResult: { nested: { [field]: "forbidden" } },
      }),
      "prohibited_biometric_data",
    );
    expect(test.consume).not.toHaveBeenCalled();
  });

  it("rejects prohibited fields emitted by validators or final authority", async () => {
    await expectCode(
      harness({
        activeLivenessValidator: {
          validateAndProject: async () =>
            ({
              ...evidence("active"),
              observedAt: NOW,
              nested: { embedding: "forbidden" },
            }) as never,
        },
      }).gate.authenticate(request),
      "prohibited_biometric_data",
    );
    await expectCode(
      harness({
        finalAuthorizationAuthority: {
          authorizeAndProject: async () =>
            ({
              authorized: true,
              authorizationRef: "authorization-1",
              nonce: request.nonce,
              profile,
              binding,
              expiresAt: NOW + 20_000,
              metrics: { confidence: 1 },
            }) as never,
        },
      }).gate.authenticate(request),
      "prohibited_biometric_data",
    );
  });

  it("validates the injected final authorization result after nonce consumption", async () => {
    const test = harness({
      finalAuthorizationAuthority: {
        authorizeAndProject: async () => ({
          authorized: true,
          authorizationRef: "authorization-1",
          nonce: "other-nonce",
          profile,
          binding,
          expiresAt: NOW + 20_000,
        }),
      },
    });
    await expectCode(test.gate.authenticate(request), "authorization_mismatch");
    expect(test.consume).toHaveBeenCalledOnce();
  });

  it("rejects final authorization action, account, session, and device drift", async () => {
    for (const [field, value] of [
      ["actionDigest", "sha256:drift"],
      ["accountId", "other-account"],
      ["sessionId", "other-session"],
      ["deviceId", "other-device"],
    ] as const) {
      const test = harness({
        finalAuthorizationAuthority: {
          authorizeAndProject: async () => ({
            authorized: true,
            authorizationRef: "authorization-1",
            nonce: request.nonce,
            profile,
            binding: { ...binding, [field]: value },
            expiresAt: NOW + 20_000,
          }),
        },
      });
      await expectCode(
        test.gate.authenticate(request),
        "authorization_mismatch",
      );
      expect(test.consume).toHaveBeenCalledOnce();
    }
  });
});

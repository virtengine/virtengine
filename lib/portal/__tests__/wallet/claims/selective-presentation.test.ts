import { describe, expect, it, vi } from "vitest";
import {
  ClaimStoreLockedError,
  SelectivePresentationAdapter,
  SelectivePresentationError,
  type CanonicalChallengeProjection,
  type PresentationBindingProjection,
  type PresentationContext,
  type SelectivePresentationDependencies,
} from "../../../src/wallet/claims";

const challenge = { opaque: "provider-policy-challenge" };
const projection: CanonicalChallengeProjection = {
  audience: "provider.example",
  orderOrCaseId: "order-42",
  purpose: "eligibility-check",
  nonce: "nonce-unique-1",
  expiresAt: "2026-08-02T13:00:00.000Z",
  holder: "holder:ve1",
  issuerAllowlist: ["issuer:one", "issuer:two"],
  statusEpoch: "status-9",
  policyDigest: "sha256:policy",
  policyVersion: "policy-v3",
  requestedClaims: [
    { id: "age-over-18", label: "Age over 18" },
    { id: "residency", label: "Country of residence" },
  ],
  consent: {
    text: "I consent to presenting these claims for this order only.",
    digest: "sha256:consent",
  },
};
const context: PresentationContext = {
  audience: projection.audience,
  orderOrCaseId: projection.orderOrCaseId,
  purpose: projection.purpose,
  nonce: projection.nonce,
  holder: projection.holder,
  policyDigest: projection.policyDigest,
  policyVersion: projection.policyVersion,
};

function presentationBinding(
  source: CanonicalChallengeProjection = projection,
): PresentationBindingProjection {
  const { requestedClaims, ...binding } = structuredClone(source);
  return {
    ...binding,
    disclosedClaimIds: requestedClaims.map(({ id }) => id),
  };
}

function harness(overrides: Partial<SelectivePresentationDependencies> = {}) {
  const validateAndProject = vi.fn(async () => structuredClone(projection));
  const readRequested = vi.fn(async (ids: readonly string[]) =>
    ids.map((id, index) => ({
      id,
      value: new Uint8Array([index + 1]),
      issuer: "issuer:one",
      statusEpoch: projection.statusEpoch,
    })),
  );
  const createPresentation = vi.fn(async () => ({ opaque: "presentation" }));
  const validatePresentation = vi.fn(async () => presentationBinding());
  const checkStatus = vi.fn(async () => ({
    revoked: false,
    statusEpoch: projection.statusEpoch,
  }));
  const consume = vi.fn(async () => true);
  const dependencies: SelectivePresentationDependencies = {
    challengeValidator: { validateAndProject },
    claimReader: { readRequested },
    presenter: {
      createPresentation,
      validateAndProject: validatePresentation,
    },
    statusAuthority: { checkStatus },
    replayGuard: { consume },
    now: () => new Date("2026-08-02T12:00:00.000Z"),
    ...overrides,
  };
  return {
    adapter: new SelectivePresentationAdapter(dependencies),
    dependencies,
    validateAndProject,
    readRequested,
    createPresentation,
    validatePresentation,
    checkStatus,
    consume,
  };
}

async function expectCode(promise: Promise<unknown>, code: string) {
  await expect(promise).rejects.toMatchObject({
    name: "SelectivePresentationError",
    code,
  });
}

describe("SelectivePresentationAdapter", () => {
  it("reviews only requested labels and exact consent without presenting", async () => {
    const test = harness();
    const review = await test.adapter.review(challenge, context);

    expect(review.requestedClaims).toEqual(projection.requestedClaims);
    expect(review.consent).toEqual(projection.consent);
    expect(JSON.stringify(review)).not.toContain("value");
    expect(test.consume).not.toHaveBeenCalled();
    expect(test.readRequested).not.toHaveBeenCalled();
    expect(test.createPresentation).not.toHaveBeenCalled();
  });

  it.each([
    "audience",
    "orderOrCaseId",
    "purpose",
    "nonce",
    "holder",
    "policyDigest",
    "policyVersion",
  ] as const)("rejects a challenge %s mismatch", async (field) => {
    const test = harness();
    await expectCode(
      test.adapter.present(challenge, { ...context, [field]: "mismatch" }),
      "binding_mismatch",
    );
    expect(test.consume).not.toHaveBeenCalled();
  });

  it("rejects expired, missing, duplicate, and unknown requested claims", async () => {
    const expired = harness({
      now: () => new Date("2026-08-02T13:00:00.000Z"),
    });
    await expectCode(expired.adapter.present(challenge, context), "expired");

    for (const requestedClaims of [
      [],
      [projection.requestedClaims[0], projection.requestedClaims[0]],
      [{ id: "", label: "Missing ID" }],
    ]) {
      const test = harness({
        challengeValidator: {
          validateAndProject: async () => ({ ...projection, requestedClaims }),
        },
      });
      await expectCode(
        test.adapter.present(challenge, context),
        "invalid_challenge",
      );
    }

    const unknown = harness({
      claimReader: {
        readRequested: async (ids) => [
          {
            id: ids[0],
            value: true,
            issuer: "issuer:one",
            statusEpoch: projection.statusEpoch,
          },
        ],
      },
    });
    await expectCode(
      unknown.adapter.present(challenge, context),
      "unknown_claim",
    );
  });

  it("rejects issuer mismatch, stale status, and revocation", async () => {
    const badIssuer = harness({
      claimReader: {
        readRequested: async (ids) =>
          ids.map((id) => ({
            id,
            value: true,
            issuer: "issuer:unapproved",
            statusEpoch: projection.statusEpoch,
          })),
      },
    });
    await expectCode(
      badIssuer.adapter.present(challenge, context),
      "issuer_mismatch",
    );

    const stale = harness({
      statusAuthority: {
        checkStatus: async () => ({ revoked: false, statusEpoch: "status-8" }),
      },
    });
    await expectCode(stale.adapter.present(challenge, context), "stale_status");

    const revoked = harness({
      statusAuthority: {
        checkStatus: async () => ({
          revoked: true,
          statusEpoch: projection.statusEpoch,
        }),
      },
    });
    await expectCode(revoked.adapter.present(challenge, context), "revoked");
  });

  it("atomically rejects replay before reading any claims", async () => {
    const test = harness({
      replayGuard: { consume: async () => false },
    });
    await expectCode(
      test.adapter.present(challenge, context),
      "replayed_nonce",
    );
    expect(test.readRequested).not.toHaveBeenCalled();
    expect(test.createPresentation).not.toHaveBeenCalled();
  });

  it("reads exactly requested IDs and rejects extra reader disclosure", async () => {
    const test = harness();
    await test.adapter.present(challenge, context);
    expect(test.readRequested).toHaveBeenCalledOnce();
    expect(test.readRequested).toHaveBeenCalledWith([
      "age-over-18",
      "residency",
    ]);

    const extra = harness({
      claimReader: {
        readRequested: async (ids) => [
          ...ids.map((id) => ({
            id,
            value: true,
            issuer: "issuer:one",
            statusEpoch: projection.statusEpoch,
          })),
          {
            id: "unrequested-secret",
            value: "secret",
            issuer: "issuer:one",
            statusEpoch: projection.statusEpoch,
          },
        ],
      },
    });
    await expectCode(extra.adapter.present(challenge, context), "extra_disclosure");
    expect(extra.createPresentation).not.toHaveBeenCalled();
  });

  it("propagates a locked reader failure and never invokes the presenter", async () => {
    const test = harness({
      claimReader: {
        readRequested: async () => {
          throw new ClaimStoreLockedError();
        },
      },
    });
    await expect(test.adapter.present(challenge, context)).rejects.toBeInstanceOf(
      ClaimStoreLockedError,
    );
    expect(test.createPresentation).not.toHaveBeenCalled();
  });

  it.each([
    "audience",
    "orderOrCaseId",
    "purpose",
    "nonce",
    "expiresAt",
    "holder",
    "statusEpoch",
    "policyDigest",
    "policyVersion",
  ] as const)("rejects a presentation %s mismatch", async (field) => {
    const test = harness();
    test.validatePresentation.mockResolvedValueOnce({
      ...presentationBinding(),
      [field]: "mismatch",
    });
    await expectCode(
      test.adapter.present(challenge, context),
      "presentation_mismatch",
    );
  });

  it("rejects presentation issuer, consent, and exact disclosure mismatches", async () => {
    for (const binding of [
      { ...presentationBinding(), issuerAllowlist: ["issuer:other"] },
      {
        ...presentationBinding(),
        consent: { ...projection.consent, text: "Changed consent" },
      },
      {
        ...presentationBinding(),
        consent: { ...projection.consent, digest: "sha256:changed" },
      },
      { ...presentationBinding(), disclosedClaimIds: ["age-over-18"] },
    ]) {
      const test = harness({
        presenter: {
          createPresentation: async () => ({ opaque: "presentation" }),
          validateAndProject: async () => binding,
        },
      });
      await expectCode(
        test.adapter.present(challenge, context),
        "presentation_mismatch",
      );
    }

    const extra = harness({
      presenter: {
        createPresentation: async () => ({ opaque: "presentation" }),
        validateAndProject: async () => ({
          ...presentationBinding(),
          disclosedClaimIds: [
            "age-over-18",
            "residency",
            "unrequested-secret",
          ],
        }),
      },
    });
    await expectCode(extra.adapter.present(challenge, context), "extra_disclosure");
  });

  it("fails closed on a malformed presentation projection", async () => {
    const test = harness({
      presenter: {
        createPresentation: async () => ({ opaque: "presentation" }),
        validateAndProject: async () =>
          ({
            ...presentationBinding(),
            disclosedClaimIds: undefined,
          }) as unknown as PresentationBindingProjection,
      },
    });
    await expectCode(
      test.adapter.present(challenge, context),
      "presentation_mismatch",
    );
  });

  it("returns only a validated opaque presentation", async () => {
    const test = harness();
    const result = await test.adapter.present(challenge, context);
    expect(result).toEqual({ opaque: "presentation" });
    expect(test.consume).toHaveBeenCalledWith(
      projection.nonce,
      projection.expiresAt,
    );
    expect(test.checkStatus).toHaveBeenCalledTimes(2);
    expect(test.validatePresentation).toHaveBeenCalledWith(result, challenge);
  });

  it.each([
    "challengeValidator",
    "claimReader",
    "presenter",
    "statusAuthority",
    "replayGuard",
  ] as const)("is unavailable without %s", async (dependency) => {
    const test = harness({ [dependency]: undefined });
    expect(test.adapter.isAvailable()).toBe(false);
    await expectCode(test.adapter.review(challenge, context), "unavailable");
    await expectCode(test.adapter.present(challenge, context), "unavailable");
  });

  it("has no usable defaults", async () => {
    const adapter = new SelectivePresentationAdapter();
    expect(adapter.isAvailable()).toBe(false);
    await expect(adapter.present(challenge, context)).rejects.toBeInstanceOf(
      SelectivePresentationError,
    );
  });
});
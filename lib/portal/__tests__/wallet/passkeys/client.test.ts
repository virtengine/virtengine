import { describe, expect, it, vi } from "vitest";
import {
  PasskeyError,
  PlatformPasskeyClient,
  type PasskeyAction,
  type PasskeyAuthenticationRequest,
  type PasskeyChallengeProjector,
  type PasskeyCounterStore,
  type PasskeyServerVerifier,
  type PlatformAssertionCredential,
  type PlatformPasskeyAuthority,
  type PlatformRegistrationCredential,
} from "../../../src/wallet/passkeys";

const now = Date.parse("2026-08-02T12:00:00Z");
const encoder = new TextEncoder();

function clientData(type: string, challenge: string, origin = "https://portal.virtengine.com") {
  return encoder.encode(JSON.stringify({ type, challenge, origin }));
}

function projectionFor(input: {
  binding: { chainId: string; account: string; sessionId: string };
  action: PasskeyAction;
}) {
  const action =
    input.action.kind === "login"
      ? `login:${input.action.purpose}`
      : `${input.action.payloadKind}:${input.action.digest}`;
  const value = `${input.binding.chainId}|${input.binding.account}|${input.binding.sessionId}|${action}`;
  return { challenge: encoder.encode(value), clientDataChallenge: value };
}

function harness() {
  let nextChallenge = "virtengine-1|ve1holder|session-1|login:wallet-login";
  let counter = 0;
  const registration: PlatformRegistrationCredential = {
    id: "credential-1",
    rawId: new Uint8Array([1, 2]),
    type: "public-key",
    authenticatorAttachment: "platform",
    response: {
      clientDataJSON: clientData("webauthn.create", nextChallenge),
      attestationObject: new Uint8Array([3, 4]),
      transports: ["internal"],
    },
  };
  const assertion: PlatformAssertionCredential = {
    id: "credential-1",
    rawId: new Uint8Array([1, 2]),
    type: "public-key",
    authenticatorAttachment: "platform",
    response: {
      clientDataJSON: clientData("webauthn.get", nextChallenge),
      authenticatorData: new Uint8Array([5, 6]),
      signature: new Uint8Array([7, 8]),
      userHandle: new Uint8Array([9]),
    },
  };
  const authority: PlatformPasskeyAuthority = {
    isPlatformAvailable: vi.fn(() => true),
    create: vi.fn(async () => registration),
    get: vi.fn(async () => assertion),
  };
  const facts = {
    verified: true,
    rpIdHashValid: true,
    userPresent: true,
    userVerified: true,
    credentialId: "credential-1",
    counter: 1,
  };
  const verifier: PasskeyServerVerifier = {
    verifyRegistration: vi.fn(async () => facts),
    verifyAssertion: vi.fn(async () => ({ ...facts, counter: ++counter })),
  };
  const projector: PasskeyChallengeProjector = {
    project: vi.fn(async input => {
      const projection = projectionFor(input);
      nextChallenge = projection.clientDataChallenge;
      const type = input.ceremony === "registration" ? "webauthn.create" : "webauthn.get";
      registration.response.clientDataJSON = clientData(type, nextChallenge);
      assertion.response.clientDataJSON = clientData(type, nextChallenge);
      return projection;
    }),
  };
  let storedCounter: number | undefined;
  const counters: PasskeyCounterStore = {
    initialize: vi.fn(async (_id, value) => {
      if (storedCounter !== undefined) return false;
      storedCounter = value;
      return true;
    }),
    advance: vi.fn(async (_id, value) => {
      if (storedCounter !== undefined && value <= storedCounter) return false;
      storedCounter = value;
      return true;
    }),
  };
  const client = new PlatformPasskeyClient(
    "portal.virtengine.com",
    "https://portal.virtengine.com",
    { authority, verifier, challengeProjector: projector, counterStore: counters, now: () => now },
  );
  const registerRequest = {
    chainId: "virtengine-1",
    account: "ve1holder",
    sessionId: "session-1",
    rp: { id: "portal.virtengine.com", name: "VirtEngine" },
    user: { id: new Uint8Array([11]), name: "holder", displayName: "Holder" },
    sessionPurpose: "wallet-login",
    expiresAt: now + 60_000,
  };
  const authRequest: PasskeyAuthenticationRequest = {
    chainId: "virtengine-1",
    account: "ve1holder",
    sessionId: "session-1",
    action: { kind: "login", purpose: "wallet-login" },
    expiresAt: now + 60_000,
  };
  return {
    client,
    authority,
    verifier,
    projector,
    counters,
    facts,
    registration,
    assertion,
    registerRequest,
    authRequest,
    setStoredCounter: (value: number | undefined) => (storedCounter = value),
  };
}

describe("PlatformPasskeyClient", () => {
  it("requires platform resident registration with UV and defaults attestation to none", async () => {
    const test = harness();
    const result = await test.client.register(test.registerRequest);
    expect(test.authority.create).toHaveBeenCalledWith(
      expect.objectContaining({
        rp: test.registerRequest.rp,
        user: test.registerRequest.user,
        authenticatorSelection: {
          authenticatorAttachment: "platform",
          residentKey: "required",
          requireResidentKey: true,
          userVerification: "required",
        },
        attestation: "none",
      }),
    );
    expect(test.projector.project).toHaveBeenCalledWith({
      ceremony: "registration",
      rpId: "portal.virtengine.com",
      binding: test.registerRequest,
      action: { kind: "login", purpose: "wallet-login" },
      expiresAt: now + 60_000,
    });
    expect(result).toMatchObject({
      credentialId: "credential-1",
      transports: ["internal"],
      counter: 1,
    });
    expect(Array.from(result.rawCredentialId)).toEqual([1, 2]);
    expect(Array.from(result.clientDataJSON)).toEqual(
      Array.from(test.registration.response.clientDataJSON),
    );
    expect(Array.from(result.attestationObject)).toEqual([3, 4]);
  });

  it("allows only an injected policy to change attestation", async () => {
    const test = harness();
    const client = new PlatformPasskeyClient(
      "portal.virtengine.com",
      "https://portal.virtengine.com",
      {
        authority: test.authority,
        verifier: test.verifier,
        challengeProjector: test.projector,
        counterStore: test.counters,
        attestationPolicy: { registrationPreference: () => "enterprise" },
        now: () => now,
      },
    );
    await client.register(test.registerRequest);
    expect(test.authority.create).toHaveBeenCalledWith(
      expect.objectContaining({ attestation: "enterprise" }),
    );
  });

  it("binds login to chain, account, session, purpose, RP and projected challenge", async () => {
    const test = harness();
    await test.client.authenticate(test.authRequest);
    expect(test.projector.project).toHaveBeenCalledWith({
      ceremony: "authentication",
      rpId: "portal.virtengine.com",
      binding: test.authRequest,
      action: { kind: "login", purpose: "wallet-login" },
      expiresAt: now + 60_000,
    });
    expect(test.verifier.verifyAssertion).toHaveBeenCalledWith(
      expect.objectContaining({
        rpId: "portal.virtengine.com",
        origin: "https://portal.virtengine.com",
        binding: test.authRequest,
      }),
    );
    const authenticationOptions = vi.mocked(test.authority.get).mock.calls[0][0];
    expect(authenticationOptions).toMatchObject({
      rpId: "portal.virtengine.com",
      userVerification: "required",
    });
    expect(new TextDecoder().decode(authenticationOptions.challenge)).toBe(
      "virtengine-1|ve1holder|session-1|login:wallet-login",
    );
    expect(authenticationOptions.allowCredentials).toBeUndefined();
  });

  it.each([
    ["transaction", "tx-digest-123"],
    ["message", "message-digest-456"],
  ] as const)("binds exact %s digest authorization", async (payloadKind, digest) => {
    const test = harness();
    const request = {
      ...test.authRequest,
      action: { kind: "authorization" as const, payloadKind, digest },
    };
    await test.client.authenticate(request);
    expect(test.projector.project).toHaveBeenCalledWith(
      expect.objectContaining({ action: request.action }),
    );
  });

  it.each([
    ["chain", { chainId: "wrong-chain" }],
    ["account", { account: "ve1wrong" }],
    ["session", { sessionId: "wrong-session" }],
    ["digest", { action: { kind: "authorization", payloadKind: "transaction", digest: "wrong" } }],
  ])("rejects a response bound to the wrong %s", async (_name, override) => {
    const test = harness();
    const original = { ...test.assertion.response };
    vi.mocked(test.authority.get).mockImplementationOnce(async () => ({
      ...test.assertion,
      response: {
        ...original,
        clientDataJSON: clientData(
          "webauthn.get",
          "virtengine-1|ve1holder|session-1|transaction:tx-digest-123",
        ),
      },
    }));
    await expect(
      test.client.authenticate({
        ...test.authRequest,
        action: { kind: "authorization", payloadKind: "transaction", digest: "tx-digest-123" },
        ...override,
      } as PasskeyAuthenticationRequest),
    ).rejects.toMatchObject({ code: "invalid_client_data" });
  });

  it.each([
    ["challenge", { challenge: "wrong" }],
    ["origin", { origin: "https://evil.example" }],
    ["type", { type: "webauthn.create" }],
  ])("rejects wrong client data %s", async (_name, override) => {
    const test = harness();
    vi.mocked(test.authority.get).mockImplementationOnce(async options => ({
      ...test.assertion,
      response: {
        ...test.assertion.response,
        clientDataJSON: clientData(
          (override.type as string | undefined) ?? "webauthn.get",
          (override.challenge as string | undefined) ?? new TextDecoder().decode(options.challenge),
          (override.origin as string | undefined) ?? "https://portal.virtengine.com",
        ),
      },
    }));
    await expect(test.client.authenticate(test.authRequest)).rejects.toMatchObject({
      code: "invalid_client_data",
    });
  });

  it.each([
    ["RP hash", { rpIdHashValid: false }],
    ["UP", { userPresent: false }],
    ["UV", { userVerified: false }],
    ["credential ID", { credentialId: "other" }],
    ["server verification", { verified: false }],
  ])("rejects invalid %s verification facts", async (_name, override) => {
    const test = harness();
    vi.mocked(test.verifier.verifyAssertion).mockResolvedValueOnce({
      ...test.facts,
      ...override,
    });
    await expect(test.client.authenticate(test.authRequest)).rejects.toMatchObject({
      code: "verification_failed",
    });
  });

  it("rejects replay and counter regression atomically", async () => {
    const test = harness();
    test.setStoredCounter(4);
    vi.mocked(test.verifier.verifyAssertion).mockResolvedValue({ ...test.facts, counter: 4 });
    await expect(test.client.authenticate(test.authRequest)).rejects.toMatchObject({
      code: "counter_replay",
    });
    vi.mocked(test.verifier.verifyAssertion).mockResolvedValue({ ...test.facts, counter: 3 });
    await expect(test.client.authenticate(test.authRequest)).rejects.toMatchObject({
      code: "counter_replay",
    });
  });

  it("rejects expired challenges before and after platform interaction", async () => {
    const test = harness();
    await expect(
      test.client.authenticate({ ...test.authRequest, expiresAt: now }),
    ).rejects.toMatchObject({ code: "expired" });
    expect(test.authority.get).not.toHaveBeenCalled();

    const expiring = new PlatformPasskeyClient(
      "portal.virtengine.com",
      "https://portal.virtengine.com",
      {
        authority: test.authority,
        verifier: test.verifier,
        challengeProjector: test.projector,
        counterStore: test.counters,
        now: vi.fn().mockReturnValueOnce(now).mockReturnValue(now + 60_001),
      },
    );
    await expect(expiring.authenticate(test.authRequest)).rejects.toMatchObject({ code: "expired" });
    expect(test.verifier.verifyAssertion).not.toHaveBeenCalled();
  });

  it("rejects ambiguous and unbound actions", async () => {
    const test = harness();
    for (const action of [
      { kind: "login", purpose: "" },
      { kind: "login", purpose: "login", digest: "also-bound" },
      { kind: "authorization", payloadKind: "transaction", digest: "" },
      { kind: "authorization", payloadKind: "transaction", digest: "x", purpose: "also-bound" },
    ]) {
      await expect(
        test.client.authenticate({ ...test.authRequest, action } as PasskeyAuthenticationRequest),
      ).rejects.toMatchObject({ code: "invalid_request" });
    }
    expect(test.authority.get).not.toHaveBeenCalled();
  });

  it("is unavailable by default and rejects unsupported platforms", async () => {
    const unavailable = new PlatformPasskeyClient(
      "portal.virtengine.com",
      "https://portal.virtengine.com",
    );
    expect(unavailable.isAvailable()).toBe(false);
    await expect(unavailable.authenticate(harness().authRequest)).rejects.toMatchObject({
      code: "unavailable",
    });

    const test = harness();
    vi.mocked(test.authority.isPlatformAvailable).mockReturnValue(false);
    expect(test.client.isAvailable()).toBe(false);
    await expect(test.client.authenticate(test.authRequest)).rejects.toMatchObject({
      code: "unsupported_platform",
    });
  });

  it.each(["authority", "verifier", "challengeProjector", "counterStore"] as const)(
    "fails closed without %s",
    async missing => {
      const test = harness();
      const dependencies = {
        authority: test.authority,
        verifier: test.verifier,
        challengeProjector: test.projector,
        counterStore: test.counters,
        now: () => now,
      };
      delete (dependencies as Partial<typeof dependencies>)[missing];
      const client = new PlatformPasskeyClient(
        "portal.virtengine.com",
        "https://portal.virtengine.com",
        dependencies,
      );
      expect(client.isAvailable()).toBe(false);
      await expect(client.authenticate(test.authRequest)).rejects.toMatchObject({
        code: "unavailable",
      });
    },
  );

  it("returns only credential bytes and allowlisted metadata, never facial data", async () => {
    const test = harness();
    Object.assign(test.assertion, {
      facialMetrics: [0.1],
      faceTemplate: "template",
      image: "data:image/png;base64,...",
    });
    const result = await test.client.authenticate(test.authRequest);
    expect(Object.keys(result).sort()).toEqual([
      "authenticatorData",
      "clientDataJSON",
      "counter",
      "credentialId",
      "rawCredentialId",
      "signature",
      "userHandle",
    ]);
    expect(JSON.stringify(result)).not.toMatch(/face|facial|template|image|metric/i);
  });

  it("surfaces typed passkey errors", () => {
    expect(new PasskeyError("expired", "expired")).toMatchObject({
      name: "PasskeyError",
      code: "expired",
    });
  });
});
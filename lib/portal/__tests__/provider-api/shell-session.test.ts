import { describe, expect, it } from "vitest";
import {
  buildProviderShellWebSocketUrl,
  validateProviderShellSessionReceipt,
} from "../../src/provider-api/shell-session";
import type {
  ProviderShellSessionCapability,
  ShellEligibilityProjection,
} from "../../src/provider-api/shell-session";

const now = new Date("2026-08-02T12:00:00Z");
const capability: ProviderShellSessionCapability = {
  receiptVersion: "v1",
  transport: "one_time_reference",
  maxTtlSeconds: 120,
};
const eligibility: ShellEligibilityProjection = {
  authorized: true,
  chainId: "virtengine-1",
  account: "ve1account",
  deploymentId: "deployment-1",
  container: "web",
  providerId: "ve1provider",
  sessionId: "eligibility-session-1",
  policyEpoch: "policy-17",
  statusEpoch: "status-31",
  policyExpiresAt: "2026-08-02T12:05:00Z",
  statusExpiresAt: "2026-08-02T12:01:00Z",
  capabilityDigest: "sha256:capability",
  policyDigest: "sha256:policy",
};
const response = (overrides: Record<string, unknown> = {}) => ({
  version: "v1",
  deployment_id: eligibility.deploymentId,
  container: eligibility.container,
  account: eligibility.account,
  provider_id: eligibility.providerId,
  chain_id: eligibility.chainId,
  eligibility_session_id: eligibility.sessionId,
  session_id: "one-time-reference-1",
  issued_at: "2026-08-02T11:59:55Z",
  expires_at: "2026-08-02T12:01:00Z",
  capability_digest: eligibility.capabilityDigest,
  policy_digest: eligibility.policyDigest,
  ...overrides,
});
const context = (overrides: Record<string, unknown> = {}) => ({
  eligibility,
  capability,
  providerEndpoint: "https://provider.example.com",
  now,
  ...overrides,
});

describe("provider shell session authority", () => {
  it("defaults unavailable without injected authoritative eligibility", () => {
    expect(() =>
      validateProviderShellSessionReceipt(
        response(),
        context({ eligibility: undefined }),
      ),
    ).toThrowError(
      expect.objectContaining({ code: "eligibility_unavailable" }),
    );
  });

  it("rejects expired policy or status projections", () => {
    expect(() =>
      validateProviderShellSessionReceipt(
        response(),
        context({
          eligibility: {
            ...eligibility,
            statusExpiresAt: "2026-08-02T11:59:59Z",
          },
        }),
      ),
    ).toThrowError(expect.objectContaining({ code: "eligibility_expired" }));
  });

  it.each([
    ["deployment_id", "other-deployment"],
    ["container", "worker"],
    ["account", "ve1other"],
    ["provider_id", "ve1other-provider"],
    ["chain_id", "other-chain"],
    ["eligibility_session_id", "other-session"],
    ["capability_digest", "sha256:other-capability"],
    ["policy_digest", "sha256:other-policy"],
  ])("rejects a mismatched %s binding", (field, value) => {
    expect(() =>
      validateProviderShellSessionReceipt(
        response({ [field]: value }),
        context(),
      ),
    ).toThrowError(expect.objectContaining({ code: "receipt_mismatch" }));
  });

  it("rejects expired and overlong provider receipts", () => {
    expect(() =>
      validateProviderShellSessionReceipt(
        response({ expires_at: "2026-08-02T12:03:00Z" }),
        context(),
      ),
    ).toThrowError(expect.objectContaining({ code: "receipt_invalid" }));
    expect(() =>
      validateProviderShellSessionReceipt(
        response({ expires_at: "2026-08-02T11:59:59Z" }),
        context(),
      ),
    ).toThrowError(expect.objectContaining({ code: "receipt_expired" }));
  });

  it("keeps legacy token responses typed unavailable and rejects reusable bearers", () => {
    expect(() =>
      validateProviderShellSessionReceipt(
        { token: "reusable", deployment: "deployment-1" },
        context(),
      ),
    ).toThrowError(expect.objectContaining({ code: "unsafe_transport" }));
    expect(() =>
      validateProviderShellSessionReceipt({ expires_at: "soon" }, context()),
    ).toThrowError(expect.objectContaining({ code: "feature_unavailable" }));
  });

  it("rejects a server URL with token parameters or a foreign provider", () => {
    expect(() =>
      validateProviderShellSessionReceipt(
        response({
          websocket_url: "wss://provider.example.com/shell?token=secret",
        }),
        context(),
      ),
    ).toThrowError(expect.objectContaining({ code: "unsafe_transport" }));
    expect(() =>
      validateProviderShellSessionReceipt(
        response({
          websocket_url: "wss://other.example.com/shell?session_id=opaque",
        }),
        context(),
      ),
    ).toThrowError(expect.objectContaining({ code: "unsafe_transport" }));
    expect(() =>
      validateProviderShellSessionReceipt(
        response({
          websocket_url: "wss://provider.example.com/shell?session_id=other",
        }),
        context(),
      ),
    ).toThrowError(expect.objectContaining({ code: "receipt_mismatch" }));
  });

  it("binds an opaque one-time reference into the only allowed query parameter", () => {
    const receipt = validateProviderShellSessionReceipt(response(), context());
    const url = new URL(
      buildProviderShellWebSocketUrl(
        "https://provider.example.com",
        eligibility.deploymentId,
        receipt,
      ),
    );

    expect(receipt.eligibilitySessionId).toBe(eligibility.sessionId);
    expect(url.protocol).toBe("wss:");
    expect([...url.searchParams.keys()]).toEqual(["session_id"]);
    expect(url.searchParams.get("session_id")).toBe("one-time-reference-1");
    expect(url.toString()).not.toContain("token");
  });

  it("accepts a strictly bound server-issued URL", () => {
    const receipt = validateProviderShellSessionReceipt(
      response({
        websocket_url:
          "wss://provider.example.com/api/v1/deployments/deployment-1/shell?session_id=one-time-reference-1",
      }),
      context({ capability: { ...capability, transport: "server_url" } }),
    );

    expect(receipt.websocketUrl).toContain("session_id=one-time-reference-1");
  });
});

import { afterEach, describe, expect, it, vi } from "vitest";
import { ProviderAPIClient } from "../../src/provider-api/client";
import type { ShellEligibilityProjection } from "../../src/provider-api/shell-session";

const eligibility = (): ShellEligibilityProjection => {
  const now = Date.now();
  return {
    authorized: true,
    chainId: "virtengine-1",
    account: "ve1account",
    deploymentId: "deployment-1",
    container: "web",
    providerId: "ve1provider",
    sessionId: "eligibility-session-1",
    policyEpoch: "policy-17",
    statusEpoch: "status-31",
    policyExpiresAt: new Date(now + 300_000).toISOString(),
    statusExpiresAt: new Date(now + 60_000).toISOString(),
    capabilityDigest: "sha256:capability",
    policyDigest: "sha256:policy",
  };
};

const receipt = (projection: ShellEligibilityProjection) => ({
  version: "v1",
  deployment_id: projection.deploymentId,
  container: projection.container,
  account: projection.account,
  provider_id: projection.providerId,
  chain_id: projection.chainId,
  eligibility_session_id: projection.sessionId,
  session_id: "one-time-reference-1",
  issued_at: new Date(Date.now() - 1_000).toISOString(),
  expires_at: new Date(Date.now() + 60_000).toISOString(),
  capability_digest: projection.capabilityDigest,
  policy_digest: projection.policyDigest,
});

const capability = {
  receiptVersion: "v1" as const,
  transport: "one_time_reference" as const,
  maxTtlSeconds: 120,
};

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("ProviderAPIClient shell authority", () => {
  it("does not call the provider without eligibility or declared capability", async () => {
    const fetcher = vi.fn();
    const withoutEligibility = new ProviderAPIClient({
      endpoint: "https://provider.example.com",
      providerId: "ve1provider",
      shellSessionCapability: capability,
      fetcher: fetcher as typeof fetch,
    });
    await expect(withoutEligibility.createShellSession()).rejects.toMatchObject(
      {
        code: "eligibility_unavailable",
      },
    );

    const withoutCapability = new ProviderAPIClient({
      endpoint: "https://provider.example.com",
      providerId: "ve1provider",
      fetcher: fetcher as typeof fetch,
    });
    await expect(
      withoutCapability.createShellSession(eligibility()),
    ).rejects.toMatchObject({
      code: "feature_unavailable",
    });
    expect(fetcher).not.toHaveBeenCalled();
  });

  it("ignores localStorage bearer tokens and connects with only a one-time reference", async () => {
    const projection = eligibility();
    const fetcher = vi.fn(
      async (_input: RequestInfo | URL, init?: RequestInit) => {
        const body = JSON.parse(String(init?.body)) as Record<string, unknown>;
        expect(body).toMatchObject({
          container: projection.container,
          account: projection.account,
          provider_id: projection.providerId,
          chain_id: projection.chainId,
          eligibility_session_id: projection.sessionId,
          policy_epoch: projection.policyEpoch,
          status_epoch: projection.statusEpoch,
          capability_digest: projection.capabilityDigest,
          policy_digest: projection.policyDigest,
        });
        return new Response(JSON.stringify(receipt(projection)), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      },
    );
    const localStorage = {
      getItem: vi.fn(() => {
        throw new Error("localStorage must not be read");
      }),
    };
    const webSocketUrls: string[] = [];
    class FakeWebSocket {
      onopen: (() => void) | null = null;
      onmessage: ((event: MessageEvent) => void) | null = null;
      onerror: ((event: Event) => void) | null = null;
      onclose: ((event: CloseEvent) => void) | null = null;

      constructor(url: string) {
        webSocketUrls.push(url);
      }

      send() {}
      close() {}
    }
    vi.stubGlobal("localStorage", localStorage);
    vi.stubGlobal("WebSocket", FakeWebSocket);
    const client = new ProviderAPIClient({
      endpoint: "https://provider.example.com",
      providerId: projection.providerId,
      shellSessionCapability: capability,
      fetcher: fetcher as typeof fetch,
    });

    const validated = await client.createShellSession(projection);
    const connection = client.connectShell(validated);

    expect(connection).toBeDefined();
    expect(localStorage.getItem).not.toHaveBeenCalled();
    expect(webSocketUrls).toHaveLength(1);
    const url = new URL(webSocketUrls[0]);
    expect([...url.searchParams.keys()]).toEqual(["session_id"]);
    expect(url.searchParams.get("session_id")).toBe("one-time-reference-1");
    expect(url.toString()).not.toContain("token");
    expect(() => client.connectShell(validated)).toThrowError(
      expect.objectContaining({ code: "receipt_invalid" }),
    );
    expect(webSocketUrls).toHaveLength(1);
  });
});

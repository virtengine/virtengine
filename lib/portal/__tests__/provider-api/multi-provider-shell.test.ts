import { afterEach, describe, expect, it, vi } from "vitest";
import { MultiProviderClient } from "../../src/multi-provider/client";
import type { ShellEligibilityProjection } from "../../src/provider-api/shell-session";

const jsonResponse = (body: unknown, status = 200): Response =>
  new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });

const eligibility = (): ShellEligibilityProjection => ({
  authorized: true,
  chainId: "virtengine-1",
  account: "ve1account",
  deploymentId: "deployment-1",
  container: "web",
  providerId: "ve1provider",
  sessionId: "eligibility-session-1",
  policyEpoch: "policy-17",
  statusEpoch: "status-31",
  policyExpiresAt: new Date(Date.now() + 300_000).toISOString(),
  statusExpiresAt: new Date(Date.now() + 60_000).toISOString(),
  capabilityDigest: "sha256:capability",
  policyDigest: "sha256:policy",
});

const providerResponse = (withShellCapability: boolean) => ({
  providers: [
    {
      owner: "ve1provider",
      host_uri: "https://provider.example.com",
      attributes: withShellCapability
        ? [
            { key: "shell_session_receipt_version", value: "v1" },
            { key: "shell_session_transport", value: "one_time_reference" },
            { key: "shell_session_max_ttl_seconds", value: "120" },
          ]
        : [],
    },
  ],
  pagination: {},
});

afterEach(() => vi.unstubAllGlobals());

describe("MultiProviderClient shell authority", () => {
  it("resolves capability and provider before creating a valid shell connection", async () => {
    const projection = eligibility();
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
    vi.stubGlobal("WebSocket", FakeWebSocket);
    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes("/virtengine/provider/"))
        return jsonResponse(providerResponse(true));
      if (url.endsWith("/api/v1/deployments/deployment-1")) {
        return jsonResponse({ id: "deployment-1", state: "running" });
      }
      if (url.endsWith("/api/v1/deployments/deployment-1/shell/session")) {
        return jsonResponse({
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
      }
      return jsonResponse({ message: "not found" }, 404);
    });
    const client = new MultiProviderClient({
      chainEndpoint: "https://chain.example.com",
      fetcher: fetcher as typeof fetch,
    });

    const session = await client.connectShell(projection);

    expect(session.expiresAt).toBeTruthy();
    expect(webSocketUrls).toHaveLength(1);
    expect(new URL(webSocketUrls[0]).searchParams.get("session_id")).toBe(
      "one-time-reference-1",
    );
    session.connection.close();
  });

  it("does not call the session endpoint for a legacy provider", async () => {
    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes("/virtengine/provider/"))
        return jsonResponse(providerResponse(false));
      if (url.endsWith("/api/v1/deployments/deployment-1")) {
        return jsonResponse({ id: "deployment-1", state: "running" });
      }
      throw new Error("Shell session endpoint must not be called");
    });
    const client = new MultiProviderClient({
      chainEndpoint: "https://chain.example.com",
      fetcher: fetcher as typeof fetch,
    });

    await expect(client.connectShell(eligibility())).rejects.toMatchObject({
      code: "feature_unavailable",
    });
    expect(
      fetcher.mock.calls.some(([input]) =>
        String(input).includes("/shell/session"),
      ),
    ).toBe(false);
  });
});

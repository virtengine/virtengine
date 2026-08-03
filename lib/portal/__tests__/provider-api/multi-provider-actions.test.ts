import { describe, expect, it, vi } from "vitest";
import { MultiProviderClient } from "../../src/multi-provider/client";

const jsonResponse = (body: unknown, status = 200): Response =>
  new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });

const receipt = (overrides: Record<string, unknown> = {}) => ({
  operation_id: "operation-1",
  action: "restart",
  deployment_id: "deployment-1",
  provider_id: "provider-1",
  status: "committed",
  issued_at: "2026-08-01T12:00:00Z",
  completed_at: "2026-08-01T12:00:01Z",
  state: "running",
  version: "2",
  revision: "7",
  ...overrides,
});

const createFetcher = (actionResponse: unknown) =>
  vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.includes("/virtengine/provider/")) {
      return jsonResponse({
        providers: [
          {
            owner: "provider-1",
            host_uri: "https://provider.example.com",
            attributes: [
              { key: "deployment_action_receipt_version", value: "v1" },
              {
                key: "deployment_action_requires_chain_signing",
                value: "false",
              },
            ],
          },
        ],
        pagination: {},
      });
    }
    if (url.endsWith("/api/v1/deployments/deployment-1/actions")) {
      expect(init?.method).toBe("POST");
      return jsonResponse(actionResponse);
    }
    if (url.endsWith("/api/v1/deployments/deployment-1")) {
      return jsonResponse({
        id: "deployment-1",
        state: "running",
        version: "1",
        revision: "6",
      });
    }
    return jsonResponse({ message: "not found" }, 404);
  });

describe("MultiProviderClient deployment actions", () => {
  it("returns the injected projection of a validated provider receipt", async () => {
    const fetcher = createFetcher(receipt());
    const projector = vi.fn((value) => ({ ...value, state: "running" }));
    const client = new MultiProviderClient({
      chainEndpoint: "https://chain.example.com",
      fetcher: fetcher as typeof fetch,
      deploymentActionReceiptProjector: projector,
    });

    const result = await client.performAction("deployment-1", "restart");

    expect(result.operationId).toBe("operation-1");
    expect(result.providerId).toBe("provider-1");
    expect(projector).toHaveBeenCalledOnce();
  });

  it("keeps a legacy success envelope typed unavailable", async () => {
    const client = new MultiProviderClient({
      chainEndpoint: "https://chain.example.com",
      fetcher: createFetcher({
        success: true,
        message: "restarted",
      }) as typeof fetch,
    });

    await expect(
      client.performAction("deployment-1", "restart"),
    ).rejects.toMatchObject({
      code: "feature_unavailable",
    });
  });

  it("rejects a provider mismatch before projection", async () => {
    const projector = vi.fn((value) => value);
    const client = new MultiProviderClient({
      chainEndpoint: "https://chain.example.com",
      fetcher: createFetcher(
        receipt({ provider_id: "provider-2" }),
      ) as typeof fetch,
      deploymentActionReceiptProjector: projector,
    });

    await expect(
      client.performAction("deployment-1", "restart"),
    ).rejects.toMatchObject({
      code: "receipt_mismatch",
    });
    expect(projector).not.toHaveBeenCalled();
  });

  it("fails before POST when the provider omits receipt capability", async () => {
    const fetcher = createFetcher(receipt());
    fetcher.mockImplementation(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes("/virtengine/provider/")) {
        return jsonResponse({
          providers: [
            {
              owner: "provider-1",
              host_uri: "https://provider.example.com",
            },
          ],
          pagination: {},
        });
      }
      if (url.endsWith("/api/v1/deployments/deployment-1")) {
        return jsonResponse({ id: "deployment-1", state: "running" });
      }
      throw new Error("Action endpoint must not be called");
    });
    const client = new MultiProviderClient({
      chainEndpoint: "https://chain.example.com",
      fetcher: fetcher as typeof fetch,
    });

    await expect(
      client.performAction("deployment-1", "restart"),
    ).rejects.toMatchObject({
      code: "feature_unavailable",
    });
  });
});

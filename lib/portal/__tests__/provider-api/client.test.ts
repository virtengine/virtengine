/**
 * ProviderAPIClient unit tests
 * AC-1, AC-2, AC-4, AC-6 — REST client, HMAC auth, error handling
 */
import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  ProviderAPIClient,
  ProviderAPIError,
} from "../../src/provider-api/client";
import type { ProviderAPIClientOptions } from "../../src/provider-api/client";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function errorResponse(
  status: number,
  body: { message?: string; code?: string } = {},
): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function createClient(
  overrides: Partial<ProviderAPIClientOptions> = {},
  fetchMock?: ReturnType<typeof vi.fn>,
) {
  const fetcher = fetchMock ?? vi.fn(() => Promise.resolve(jsonResponse({})));
  return {
    client: new ProviderAPIClient({
      endpoint: "https://provider.example.com",
      retries: 0,
      fetcher: fetcher as unknown as typeof fetch,
      ...overrides,
    }),
    fetcher,
  };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("ProviderAPIClient", () => {
  // -----------------------------------------------------------------------
  // AC-1: Constructor & Configuration
  // -----------------------------------------------------------------------
  describe("constructor", () => {
    it("strips trailing slash from endpoint", () => {
      const { fetcher, client } = createClient({
        endpoint: "https://provider.example.com/",
      });
      fetcher.mockResolvedValueOnce(
        jsonResponse({ status: "ok", version: "1.0" }),
      );
      void client.health();
      expect(fetcher).toHaveBeenCalledWith(
        expect.stringContaining("https://provider.example.com/api/v1/health"),
        expect.anything(),
      );
    });
  });

  // -----------------------------------------------------------------------
  // AC-2: REST Endpoint Methods
  // -----------------------------------------------------------------------
  describe("health()", () => {
    it("calls GET /api/v1/health without auth", async () => {
      const { client, fetcher } = createClient();
      fetcher.mockResolvedValueOnce(
        jsonResponse({ status: "ok", version: "1.0.0" }),
      );

      const result = await client.health();

      expect(result).toEqual({ status: "ok", version: "1.0.0" });
      const [url, init] = fetcher.mock.calls[0] as [string, RequestInit];
      expect(url).toContain("/api/v1/health");
      expect(init.method).toBe("GET");
      // Health does not require auth headers
      expect(init.headers).not.toHaveProperty("X-VE-Signature");
    });
  });

  describe("getDeploymentStatus()", () => {
    it("returns normalised DeploymentStatus", async () => {
      const { client, fetcher } = createClient();
      fetcher.mockResolvedValueOnce(
        jsonResponse({
          lease_id: "lease-1",
          state: "active",
          replicas: { ready: 2, total: 3 },
          services: [{ name: "web", state: "running", replicas: 2, ports: [] }],
          last_updated: "2025-01-01T00:00:00Z",
        }),
      );

      const status = await client.getDeploymentStatus("lease-1");

      expect(status.leaseId).toBe("lease-1");
      expect(status.state).toBe("active");
      expect(status.replicas).toEqual({ ready: 2, total: 3 });
      expect(status.services).toHaveLength(1);
    });
  });

  describe("getDeploymentLogs()", () => {
    it("returns array of log lines from array response", async () => {
      const lines = ["line-1", "line-2", "line-3"];
      const { client, fetcher } = createClient();
      fetcher.mockResolvedValueOnce(jsonResponse(lines));

      const logs = await client.getDeploymentLogs("lease-1");

      expect(logs).toEqual(lines);
    });

    it("splits string response into lines", async () => {
      const { client, fetcher } = createClient();
      fetcher.mockResolvedValueOnce(jsonResponse("line-A\nline-B\nline-C"));

      const logs = await client.getDeploymentLogs("lease-1");

      expect(logs).toEqual(["line-A", "line-B", "line-C"]);
    });

    it("passes query params for tail and timestamps", async () => {
      const { client, fetcher } = createClient();
      fetcher.mockResolvedValueOnce(jsonResponse([]));

      await client.getDeploymentLogs("lease-1", {
        tail: 100,
        timestamps: true,
      });

      const url = (fetcher.mock.calls[0] as [string, RequestInit])[0];
      expect(url).toContain("tail=100");
      expect(url).toContain("timestamps=true");
    });
  });

  describe("getDeploymentMetrics()", () => {
    it("normalises timestamp to Date", async () => {
      const { client, fetcher } = createClient();
      fetcher.mockResolvedValueOnce(
        jsonResponse({
          cpu: { usage: 50, limit: 100 },
          memory: { usage: 256, limit: 512 },
          storage: { usage: 10, limit: 100 },
          timestamp: "2025-06-01T12:00:00Z",
        }),
      );

      const metrics = await client.getDeploymentMetrics("lease-1");

      expect(metrics.cpu).toEqual({ usage: 50, limit: 100 });
      expect(metrics.timestamp).toBeInstanceOf(Date);
    });
  });

  describe("performAction()", () => {
    it("sends POST with action body", async () => {
      const { client, fetcher } = createClient();
      fetcher.mockResolvedValueOnce(
        jsonResponse({ success: true, message: "restarted" }),
      );

      const result = await client.performAction("lease-1", "restart");

      expect(result).toEqual({ success: true, message: "restarted" });
      const [, init] = fetcher.mock.calls[0] as [string, RequestInit];
      expect(init.method).toBe("POST");
      expect(JSON.parse(init.body as string)).toEqual({ action: "restart" });
    });
  });

  describe("listDeployments()", () => {
    it("returns deployments and cursor", async () => {
      const { client, fetcher } = createClient();
      fetcher.mockResolvedValueOnce(
        jsonResponse({
          deployments: [
            { id: "d1", state: "active", created_at: "2025-01-01T00:00:00Z" },
          ],
          next_cursor: "abc",
        }),
      );

      const result = await client.listDeployments({ limit: 5 });

      expect(result.deployments).toHaveLength(1);
      expect(result.deployments[0].id).toBe("d1");
      expect(result.nextCursor).toBe("abc");
    });
  });

  // -----------------------------------------------------------------------
  // AC-4: HMAC Authentication
  // -----------------------------------------------------------------------
  describe("HMAC authentication", () => {
    it("includes signature headers when hmac config is provided", async () => {
      const { client, fetcher } = createClient({
        hmac: { secret: "test-secret", principal: "admin" },
      });
      fetcher.mockResolvedValueOnce(
        jsonResponse({
          leaseId: "x",
          state: "active",
          replicas: { ready: 1, total: 1 },
          services: [],
        }),
      );

      await client.getDeploymentStatus("lease-1");

      const [, init] = fetcher.mock.calls[0] as [string, RequestInit];
      const headers = init.headers as Record<string, string>;
      expect(headers["X-VE-Signature"]).toBeDefined();
      expect(headers["X-VE-Timestamp"]).toBeDefined();
      expect(headers["X-VE-Principal"]).toBe("admin");
    });
  });

  // -----------------------------------------------------------------------
  // AC-6: Error Handling
  // -----------------------------------------------------------------------
  describe("error handling", () => {
    it("throws ProviderAPIError on non-ok response", async () => {
      const { client, fetcher } = createClient();
      fetcher.mockResolvedValueOnce(
        errorResponse(404, {
          message: "deployment not found",
          code: "NOT_FOUND",
        }),
      );

      await expect(client.getDeploymentStatus("bad")).rejects.toThrow(
        ProviderAPIError,
      );

      try {
        await client.getDeploymentStatus("bad");
      } catch (err) {
        // second call for assertion
        fetcher.mockResolvedValueOnce(
          errorResponse(404, {
            message: "deployment not found",
            code: "NOT_FOUND",
          }),
        );
      }
    });

    it("ProviderAPIError exposes status and code", async () => {
      const { client, fetcher } = createClient();
      fetcher.mockResolvedValueOnce(
        errorResponse(403, { message: "forbidden", code: "AUTH_FAILED" }),
      );

      try {
        await client.getDeploymentStatus("x");
        expect.fail("should have thrown");
      } catch (err) {
        const apiErr = err as ProviderAPIError;
        expect(apiErr).toBeInstanceOf(ProviderAPIError);
        expect(apiErr.status).toBe(403);
        expect(apiErr.code).toBe("AUTH_FAILED");
      }
    });

    it("does not retry on 4xx errors", async () => {
      const { client, fetcher } = createClient({ retries: 2 });
      fetcher.mockResolvedValue(
        errorResponse(401, { message: "unauthorized" }),
      );

      await expect(client.health()).rejects.toThrow(ProviderAPIError);

      // Health doesn't require auth so fetch is only called once
      // even with retries=2, because 401 < 500
      expect(fetcher).toHaveBeenCalledTimes(1);
    });

    it("retries on 5xx errors", async () => {
      const { client, fetcher } = createClient({ retries: 1, retryDelayMs: 0 });
      fetcher
        .mockResolvedValueOnce(errorResponse(503, { message: "unavailable" }))
        .mockResolvedValueOnce(jsonResponse({ status: "ok", version: "1.0" }));

      const result = await client.health();

      expect(result.status).toBe("ok");
      expect(fetcher).toHaveBeenCalledTimes(2);
    });
  });

  // -----------------------------------------------------------------------
  // AC-1: Timeout handling
  // -----------------------------------------------------------------------
  describe("timeout", () => {
    it("aborts request when timeout expires", async () => {
      const { client, fetcher } = createClient({ timeoutMs: 1 });
      fetcher.mockImplementation(
        (_url: string, init: RequestInit) =>
          new Promise((resolve, reject) => {
            const signal = init?.signal as AbortSignal | undefined;
            if (signal) {
              signal.addEventListener("abort", () =>
                reject(new DOMException("Aborted", "AbortError")),
              );
            }
            // Never resolve naturally — rely on abort
          }),
      );

      await expect(client.health()).rejects.toThrow();
    });
  });
});

describe("ProviderAPIError", () => {
  it("extends Error with name ProviderAPIError", () => {
    const err = new ProviderAPIError("fail", 500);
    expect(err).toBeInstanceOf(Error);
    expect(err.name).toBe("ProviderAPIError");
    expect(err.message).toBe("fail");
    expect(err.status).toBe(500);
  });

  it("extracts code from object payload", () => {
    const err = new ProviderAPIError("fail", 400, {
      code: "INVALID",
      message: "bad",
    });
    expect(err.code).toBe("INVALID");
  });
});

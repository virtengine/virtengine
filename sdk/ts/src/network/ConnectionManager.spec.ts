import { afterEach, beforeEach, describe, expect, it, jest } from "@jest/globals";

import { ConnectionManager, type EndpointConfig } from "./ConnectionManager.ts";

function createMockFetch(responses: Map<string, { ok: boolean; status: number }>) {
  return jest.fn<typeof globalThis.fetch>().mockImplementation(async (input: RequestInfo | URL) => {
    const url = typeof input === "string" ? input : input.toString();
    const match = responses.get(url) ?? { ok: true, status: 200 };
    return { ok: match.ok, status: match.status, headers: new Headers() } as Response;
  });
}

describe("ConnectionManager", () => {
  let manager: ConnectionManager;
  let mockFetch: jest.MockedFunction<typeof globalThis.fetch>;

  const rpc1: EndpointConfig = { url: "https://rpc1.example.com", type: "rpc" };
  const rpc2: EndpointConfig = { url: "https://rpc2.example.com", type: "rpc", priority: 1 };
  const rest1: EndpointConfig = { url: "https://rest1.example.com", type: "rest" };

  beforeEach(() => {
    mockFetch = createMockFetch(new Map([
      ["https://rpc1.example.com/status", { ok: true, status: 200 }],
      ["https://rpc2.example.com/status", { ok: true, status: 200 }],
      ["https://rest1.example.com/cosmos/base/tendermint/v1beta1/node_info", { ok: true, status: 200 }],
    ]));

    manager = new ConnectionManager({
      healthCheckIntervalMs: 60_000,
      healthCheckTimeoutMs: 2_000,
      fetch: mockFetch,
    });
  });

  afterEach(() => {
    manager.stop();
  });

  describe("addEndpoint / removeEndpoint", () => {
    it("adds endpoints and lists them", () => {
      manager.addEndpoint(rpc1);
      manager.addEndpoint(rest1);
      expect(manager.getAllEndpoints()).toHaveLength(2);
    });

    it("ignores duplicate endpoints", () => {
      manager.addEndpoint(rpc1);
      manager.addEndpoint(rpc1);
      expect(manager.getAllEndpoints()).toHaveLength(1);
    });

    it("removes an endpoint by URL", () => {
      manager.addEndpoint(rpc1);
      manager.addEndpoint(rpc2);
      expect(manager.removeEndpoint(rpc1.url)).toBe(true);
      expect(manager.getAllEndpoints()).toHaveLength(1);
    });

    it("returns false removing non-existent endpoint", () => {
      expect(manager.removeEndpoint("https://nope.example.com")).toBe(false);
    });
  });

  describe("getBestEndpoint", () => {
    it("throws when no endpoints of the requested type exist", () => {
      expect(() => manager.getBestEndpoint("rpc")).toThrow("No rpc endpoints registered");
    });

    it("returns the only endpoint when one is registered", () => {
      manager.addEndpoint(rpc1);
      expect(manager.getBestEndpoint("rpc")).toBe(rpc1.url);
    });

    it("returns unknown endpoints (initial state) when none are checked", () => {
      manager.addEndpoint(rpc1);
      manager.addEndpoint(rpc2);
      // Both are unknown, prefer by priority (lower wins)
      expect(manager.getBestEndpoint("rpc")).toBe(rpc1.url);
    });

    it("prefers healthy over degraded endpoint", async () => {
      manager.addEndpoint(rpc1);
      manager.addEndpoint(rpc2);

      // Mark rpc1 degraded, rpc2 healthy
      manager.reportFailure(rpc1.url);
      manager.reportSuccess(rpc2.url);

      expect(manager.getBestEndpoint("rpc")).toBe(rpc2.url);
    });

    it("prefers lower priority when both healthy", async () => {
      manager.addEndpoint(rpc1); // priority 0
      manager.addEndpoint(rpc2); // priority 1

      manager.reportSuccess(rpc1.url);
      manager.reportSuccess(rpc2.url);

      expect(manager.getBestEndpoint("rpc")).toBe(rpc1.url);
    });

    it("falls back to degraded when no healthy endpoints", () => {
      manager.addEndpoint(rpc1);
      manager.reportFailure(rpc1.url); // 1 failure = degraded

      expect(manager.getBestEndpoint("rpc")).toBe(rpc1.url);
      expect(manager.getEndpointHealth(rpc1.url)).toBe("degraded");
    });
  });

  describe("health checks", () => {
    it("marks endpoint healthy after successful check", async () => {
      manager.addEndpoint(rpc1);
      await manager.checkAllEndpoints();

      expect(manager.getEndpointHealth(rpc1.url)).toBe("healthy");
    });

    it("marks endpoint degraded after one failure", async () => {
      const failFetch = createMockFetch(new Map([
        ["https://rpc1.example.com/status", { ok: false, status: 503 }],
      ]));

      const m = new ConnectionManager({ fetch: failFetch, degradedThreshold: 1 });
      m.addEndpoint(rpc1);

      await m.checkAllEndpoints();
      expect(m.getEndpointHealth(rpc1.url)).toBe("degraded");
    });

    it("marks endpoint unhealthy after consecutive failures", async () => {
      const failFetch = createMockFetch(new Map([
        ["https://rpc1.example.com/status", { ok: false, status: 503 }],
      ]));

      const m = new ConnectionManager({ fetch: failFetch, unhealthyThreshold: 2 });
      m.addEndpoint(rpc1);

      await m.checkAllEndpoints();
      await m.checkAllEndpoints();
      expect(m.getEndpointHealth(rpc1.url)).toBe("unhealthy");
    });

    it("recovers from unhealthy when check succeeds", async () => {
      manager.addEndpoint(rpc1);
      // Manually mark unhealthy
      manager.reportFailure(rpc1.url);
      manager.reportFailure(rpc1.url);
      manager.reportFailure(rpc1.url);

      // Now health check succeeds (mockFetch returns 200)
      await manager.checkAllEndpoints();

      expect(manager.getEndpointHealth(rpc1.url)).toBe("healthy");
    });

    it("checks correct URL per endpoint type", async () => {
      manager.addEndpoint(rpc1);
      manager.addEndpoint(rest1);

      await manager.checkAllEndpoints();

      expect(mockFetch).toHaveBeenCalledWith(
        "https://rpc1.example.com/status",
        expect.anything(),
      );
      expect(mockFetch).toHaveBeenCalledWith(
        "https://rest1.example.com/cosmos/base/tendermint/v1beta1/node_info",
        expect.anything(),
      );
    });

    it("handles timeout as failure", async () => {
      const abortFetch = jest.fn<typeof globalThis.fetch>().mockImplementation(async (_input, init) => {
        // Simulate timeout by checking signal
        return new Promise((_resolve, reject) => {
          if (init?.signal) {
            init.signal.addEventListener("abort", () => {
              const err = new DOMException("Aborted", "AbortError");
              reject(err);
            });
          }
          // Never resolves naturally - timeout fires
        });
      });

      const m = new ConnectionManager({
        fetch: abortFetch,
        healthCheckTimeoutMs: 50,
      });
      m.addEndpoint(rpc1);

      await m.checkAllEndpoints();

      const state = m.getAllEndpoints()[0];
      expect(state.consecutiveFailures).toBe(1);
      expect(state.lastError).toContain("timed out");
    });
  });

  describe("checkEndpointByUrl", () => {
    it("checks and returns health for a specific endpoint", async () => {
      manager.addEndpoint(rpc1);
      const health = await manager.checkEndpointByUrl(rpc1.url);
      expect(health).toBe("healthy");
    });

    it("throws for non-registered endpoint", async () => {
      await expect(manager.checkEndpointByUrl("https://nope.example.com"))
        .rejects.toThrow("Endpoint not registered");
    });
  });

  describe("reportFailure / reportSuccess", () => {
    it("tracks failures and updates health", () => {
      manager.addEndpoint(rpc1);
      expect(manager.getEndpointHealth(rpc1.url)).toBe("unknown");

      manager.reportFailure(rpc1.url, "connection refused");
      expect(manager.getEndpointHealth(rpc1.url)).toBe("degraded");
    });

    it("resets failure count on success", () => {
      manager.addEndpoint(rpc1);
      manager.reportFailure(rpc1.url);
      manager.reportFailure(rpc1.url);
      manager.reportSuccess(rpc1.url, 42);

      const state = manager.getAllEndpoints()[0];
      expect(state.health).toBe("healthy");
      expect(state.consecutiveFailures).toBe(0);
      expect(state.latencyMs).toBe(42);
    });
  });

  describe("start / stop", () => {
    it("reports running state", () => {
      expect(manager.isRunning()).toBe(false);
      manager.start();
      expect(manager.isRunning()).toBe(true);
      manager.stop();
      expect(manager.isRunning()).toBe(false);
    });

    it("is idempotent on start", () => {
      manager.start();
      manager.start();
      expect(manager.isRunning()).toBe(true);
      manager.stop();
    });
  });

  describe("reset", () => {
    it("resets all endpoints to unknown", () => {
      manager.addEndpoint(rpc1);
      manager.reportSuccess(rpc1.url);
      expect(manager.getEndpointHealth(rpc1.url)).toBe("healthy");

      manager.reset();
      expect(manager.getEndpointHealth(rpc1.url)).toBe("unknown");
    });
  });

  describe("getEndpointsByType", () => {
    it("filters by type", () => {
      manager.addEndpoint(rpc1);
      manager.addEndpoint(rpc2);
      manager.addEndpoint(rest1);

      expect(manager.getEndpointsByType("rpc")).toHaveLength(2);
      expect(manager.getEndpointsByType("rest")).toHaveLength(1);
      expect(manager.getEndpointsByType("grpc")).toHaveLength(0);
    });
  });
});

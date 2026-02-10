import { NetworkError } from "../utils/errors.ts";

/**
 * Health status for a single endpoint
 */
export type EndpointHealth = "healthy" | "degraded" | "unhealthy" | "unknown";

/**
 * Endpoint configuration with optional health check settings
 */
export interface EndpointConfig {
  /** Endpoint URL */
  url: string;
  /** Endpoint type */
  type: "rpc" | "rest" | "grpc";
  /** Priority (lower = preferred, default: 0) */
  priority?: number;
}

/**
 * Tracked state for a single endpoint
 */
export interface EndpointState {
  config: EndpointConfig;
  health: EndpointHealth;
  latencyMs: number;
  lastChecked: number;
  consecutiveFailures: number;
  lastError?: string;
}

/**
 * ConnectionManager configuration
 */
export interface ConnectionManagerOptions {
  /** Health check interval in milliseconds (default: 30000) */
  healthCheckIntervalMs?: number;
  /** Timeout for health check requests in milliseconds (default: 5000) */
  healthCheckTimeoutMs?: number;
  /** Number of consecutive failures before marking endpoint as unhealthy (default: 3) */
  unhealthyThreshold?: number;
  /** Number of consecutive failures before marking endpoint as degraded (default: 1) */
  degradedThreshold?: number;
  /** Whether to start health checks automatically (default: false) */
  autoStart?: boolean;
  /** Custom fetch implementation (for testing or Node.js) */
  fetch?: typeof globalThis.fetch;
}

const DEFAULT_OPTIONS: Required<Omit<ConnectionManagerOptions, "fetch" | "autoStart">> = {
  healthCheckIntervalMs: 30_000,
  healthCheckTimeoutMs: 5_000,
  unhealthyThreshold: 3,
  degradedThreshold: 1,
};

/**
 * Manages multiple chain endpoints with health checking and automatic failover.
 *
 * @example
 * ```typescript
 * const manager = new ConnectionManager({
 *   healthCheckIntervalMs: 15000,
 * });
 *
 * manager.addEndpoint({ url: "https://rpc1.virtengine.network", type: "rpc" });
 * manager.addEndpoint({ url: "https://rpc2.virtengine.network", type: "rpc", priority: 1 });
 *
 * manager.start();
 *
 * // Get the best available endpoint
 * const rpc = manager.getBestEndpoint("rpc");
 * console.log(rpc); // "https://rpc1.virtengine.network" (or fallback)
 *
 * manager.stop();
 * ```
 */
export class ConnectionManager {
  private endpoints: Map<string, EndpointState> = new Map();
  private intervalId: ReturnType<typeof setInterval> | null = null;
  private options: Required<Omit<ConnectionManagerOptions, "fetch" | "autoStart">>;
  private fetchFn: typeof globalThis.fetch;
  private running = false;

  constructor(options?: ConnectionManagerOptions) {
    this.options = {
      ...DEFAULT_OPTIONS,
      ...options,
    };
    this.fetchFn = options?.fetch ?? globalThis.fetch?.bind(globalThis);

    if (options?.autoStart) {
      this.start();
    }
  }

  /**
   * Add an endpoint to be managed
   */
  addEndpoint(config: EndpointConfig): void {
    const key = this.endpointKey(config);
    if (this.endpoints.has(key)) return;

    this.endpoints.set(key, {
      config: { priority: 0, ...config },
      health: "unknown",
      latencyMs: 0,
      lastChecked: 0,
      consecutiveFailures: 0,
    });
  }

  /**
   * Remove an endpoint from management
   */
  removeEndpoint(url: string): boolean {
    for (const [key, state] of this.endpoints) {
      if (state.config.url === url) {
        this.endpoints.delete(key);
        return true;
      }
    }
    return false;
  }

  /**
   * Get the best available endpoint URL for a given type.
   * Selection prefers healthy endpoints sorted by priority (ascending) then latency.
   * Falls back to degraded, then unknown, then unhealthy.
   * Throws if no endpoints of the requested type are registered.
   */
  getBestEndpoint(type: "rpc" | "rest" | "grpc"): string {
    const candidates = this.getEndpointsByType(type);

    if (candidates.length === 0) {
      throw new NetworkError(`No ${type} endpoints registered`);
    }

    // Prefer healthy > degraded > unknown > unhealthy
    const healthOrder: EndpointHealth[] = ["healthy", "degraded", "unknown", "unhealthy"];

    for (const health of healthOrder) {
      const matching = candidates
        .filter((s) => s.health === health)
        .sort((a, b) => {
          const prioA = a.config.priority ?? 0;
          const prioB = b.config.priority ?? 0;
          if (prioA !== prioB) return prioA - prioB;
          return a.latencyMs - b.latencyMs;
        });

      if (matching.length > 0) {
        return matching[0].config.url;
      }
    }

    // Should not reach here given the length check above, but satisfy TypeScript
    return candidates[0].config.url;
  }

  /**
   * Get all endpoint states for a given type
   */
  getEndpointsByType(type: "rpc" | "rest" | "grpc"): EndpointState[] {
    const result: EndpointState[] = [];
    for (const state of this.endpoints.values()) {
      if (state.config.type === type) {
        result.push(state);
      }
    }
    return result;
  }

  /**
   * Get all endpoint states
   */
  getAllEndpoints(): EndpointState[] {
    return Array.from(this.endpoints.values());
  }

  /**
   * Get the health status of a specific endpoint
   */
  getEndpointHealth(url: string): EndpointHealth | undefined {
    for (const state of this.endpoints.values()) {
      if (state.config.url === url) {
        return state.health;
      }
    }
    return undefined;
  }

  /**
   * Start periodic health checks
   */
  start(): void {
    if (this.running) return;
    this.running = true;

    // Run initial health check immediately
    void this.checkAllEndpoints();

    this.intervalId = setInterval(
      () => void this.checkAllEndpoints(),
      this.options.healthCheckIntervalMs,
    );
  }

  /**
   * Stop periodic health checks
   */
  stop(): void {
    this.running = false;
    if (this.intervalId !== null) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
  }

  /**
   * Whether health checks are currently running
   */
  isRunning(): boolean {
    return this.running;
  }

  /**
   * Manually trigger a health check for all endpoints
   */
  async checkAllEndpoints(): Promise<void> {
    const checks = Array.from(this.endpoints.keys()).map((key) =>
      this.checkEndpoint(key),
    );
    await Promise.allSettled(checks);
  }

  /**
   * Manually trigger a health check for a specific endpoint by URL
   */
  async checkEndpointByUrl(url: string): Promise<EndpointHealth> {
    for (const [key, state] of this.endpoints) {
      if (state.config.url === url) {
        await this.checkEndpoint(key);
        return state.health;
      }
    }
    throw new NetworkError(`Endpoint not registered: ${url}`);
  }

  /**
   * Mark an endpoint as unhealthy (e.g., after a request failure)
   */
  reportFailure(url: string, error?: string): void {
    for (const state of this.endpoints.values()) {
      if (state.config.url === url) {
        state.consecutiveFailures++;
        state.lastError = error;
        this.updateHealth(state);
        return;
      }
    }
  }

  /**
   * Mark an endpoint as healthy (e.g., after a successful request)
   */
  reportSuccess(url: string, latencyMs?: number): void {
    for (const state of this.endpoints.values()) {
      if (state.config.url === url) {
        state.consecutiveFailures = 0;
        state.lastError = undefined;
        state.health = "healthy";
        if (latencyMs !== undefined) {
          state.latencyMs = latencyMs;
        }
        return;
      }
    }
  }

  /**
   * Reset all endpoints to unknown state
   */
  reset(): void {
    for (const state of this.endpoints.values()) {
      state.health = "unknown";
      state.latencyMs = 0;
      state.lastChecked = 0;
      state.consecutiveFailures = 0;
      state.lastError = undefined;
    }
  }

  /**
   * Internal: check a single endpoint's health
   */
  private async checkEndpoint(key: string): Promise<void> {
    const state = this.endpoints.get(key);
    if (!state) return;

    const start = Date.now();
    const url = this.buildHealthCheckUrl(state.config);

    try {
      const controller = new AbortController();
      const timeout = setTimeout(
        () => controller.abort(),
        this.options.healthCheckTimeoutMs,
      );

      try {
        const response = await this.fetchFn(url, {
          method: "GET",
          signal: controller.signal,
          headers: { Accept: "application/json" },
        });

        const latency = Date.now() - start;
        state.latencyMs = latency;
        state.lastChecked = Date.now();

        if (response.ok) {
          state.consecutiveFailures = 0;
          state.lastError = undefined;
          state.health = "healthy";
        } else {
          state.consecutiveFailures++;
          state.lastError = `HTTP ${response.status}`;
          this.updateHealth(state);
        }
      } finally {
        clearTimeout(timeout);
      }
    } catch (error) {
      state.lastChecked = Date.now();
      state.latencyMs = Date.now() - start;
      state.consecutiveFailures++;

      if (error instanceof DOMException && error.name === "AbortError") {
        state.lastError = `Health check timed out after ${this.options.healthCheckTimeoutMs}ms`;
      } else {
        state.lastError = error instanceof Error ? error.message : String(error);
      }

      this.updateHealth(state);
    }
  }

  /**
   * Update health classification based on consecutive failures
   */
  private updateHealth(state: EndpointState): void {
    if (state.consecutiveFailures >= this.options.unhealthyThreshold) {
      state.health = "unhealthy";
    } else if (state.consecutiveFailures >= this.options.degradedThreshold) {
      state.health = "degraded";
    }
  }

  /**
   * Build the health check URL for an endpoint.
   * - RPC: /status
   * - REST: /cosmos/base/tendermint/v1beta1/node_info
   * - gRPC: same as REST (assumes gRPC-gateway is co-located)
   */
  private buildHealthCheckUrl(config: EndpointConfig): string {
    const base = config.url.replace(/\/+$/, "");
    switch (config.type) {
      case "rpc":
        return `${base}/status`;
      case "rest":
      case "grpc":
        return `${base}/cosmos/base/tendermint/v1beta1/node_info`;
    }
  }

  private endpointKey(config: EndpointConfig): string {
    return `${config.type}:${config.url}`;
  }
}

import { serializeRequestBody, signRequest } from "../auth/wallet-sign";
import type { WalletRequestSigner } from "../auth/wallet-sign";
import type {
  Deployment,
  DeploymentAction,
  DeploymentListResponse,
  DeploymentStatus,
  LogOptions,
  ProviderAPIErrorDetails,
  ProviderHealth,
  ResourceMetrics,
  ShellSessionResponse,
} from "./types";

export interface ProviderAPIClientOptions {
  endpoint: string;
  timeoutMs?: number;
  retries?: number;
  retryDelayMs?: number;
  wallet?: {
    signer: WalletRequestSigner;
    address: string;
    chainId: string;
  };
  hmac?: {
    secret: string;
    principal: string;
  };
  fetcher?: typeof fetch;
}

export interface ProviderAPIRequestOptions {
  params?: Record<string, string | number | boolean | undefined>;
  requiresAuth?: boolean;
  body?: unknown;
}

export class ProviderAPIError extends Error {
  status: number;
  code?: string;
  payload?: ProviderAPIErrorDetails | string;

  constructor(
    message: string,
    status: number,
    payload?: ProviderAPIErrorDetails | string,
  ) {
    super(message);
    this.name = "ProviderAPIError";
    this.status = status;
    this.payload = payload;
    this.code = typeof payload === "object" ? payload?.code : undefined;
  }
}

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

const normalizeDate = (
  value: string | number | Date | undefined,
): Date | undefined => {
  if (!value) return undefined;
  if (value instanceof Date) return value;
  if (typeof value === "number") {
    return value < 1_000_000_000_000 ? new Date(value * 1000) : new Date(value);
  }
  const parsed = Date.parse(value);
  return Number.isNaN(parsed) ? undefined : new Date(parsed);
};

const hmacSha256Hex = async (
  payload: string,
  secret: string,
): Promise<string> => {
  const encoder = new TextEncoder();
  const data = encoder.encode(payload);
  if (globalThis.crypto?.subtle) {
    const key = await globalThis.crypto.subtle.importKey(
      "raw",
      encoder.encode(secret),
      { name: "HMAC", hash: "SHA-256" },
      false,
      ["sign"],
    );
    const signature = await globalThis.crypto.subtle.sign("HMAC", key, data);
    return Array.from(new Uint8Array(signature))
      .map((byte) => byte.toString(16).padStart(2, "0"))
      .join("");
  }

  const nodeCrypto = await import("crypto");
  return nodeCrypto.createHmac("sha256", secret).update(data).digest("hex");
};

const buildQuery = (
  params?: Record<string, string | number | boolean | undefined>,
) => {
  if (!params) return "";
  const entries = Object.entries(params).filter(
    ([, value]) => value !== undefined,
  );
  if (entries.length === 0) return "";
  const search = new URLSearchParams();
  entries.forEach(([key, value]) => {
    if (value !== undefined) {
      search.set(key, String(value));
    }
  });
  const queryString = search.toString();
  return queryString ? `?${queryString}` : "";
};

class ReconnectingWebSocket {
  private url: string;
  private retryDelayMs: number;
  private maxRetries: number;
  private attempts = 0;
  private shouldReconnect = true;
  private ws: WebSocket | null = null;

  onOpen?: () => void;
  onMessage?: (event: MessageEvent) => void;
  onClose?: (event: CloseEvent) => void;
  onError?: (event: Event) => void;

  constructor(
    url: string,
    options?: { retryDelayMs?: number; maxRetries?: number },
  ) {
    this.url = url;
    this.retryDelayMs = options?.retryDelayMs ?? 2000;
    this.maxRetries = options?.maxRetries ?? 5;
    this.connect();
  }

  private connect() {
    if (typeof WebSocket === "undefined") {
      throw new Error("WebSocket is not available in this environment");
    }

    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      this.attempts = 0;
      this.onOpen?.();
    };

    this.ws.onmessage = (event) => {
      this.onMessage?.(event);
    };

    this.ws.onerror = (event) => {
      this.onError?.(event);
    };

    this.ws.onclose = (event) => {
      this.onClose?.(event);
      if (this.shouldReconnect && this.attempts < this.maxRetries) {
        this.attempts += 1;
        setTimeout(() => this.connect(), this.retryDelayMs * this.attempts);
      }
    };
  }

  send(data: string | ArrayBufferLike | Blob | ArrayBufferView) {
    this.ws?.send(data);
  }

  close() {
    this.shouldReconnect = false;
    this.ws?.close();
  }
}

export class LogStream {
  private socket: ReconnectingWebSocket;

  constructor(url: string) {
    this.socket = new ReconnectingWebSocket(url, {
      retryDelayMs: 1500,
      maxRetries: 10,
    });
  }

  onMessage(handler: (line: string) => void) {
    this.socket.onMessage = (event) => handler(String(event.data));
  }

  onOpen(handler: () => void) {
    this.socket.onOpen = handler;
  }

  onClose(handler: (event: CloseEvent) => void) {
    this.socket.onClose = handler;
  }

  onError(handler: (event: Event) => void) {
    this.socket.onError = handler;
  }

  close() {
    this.socket.close();
  }
}

export class ShellConnection {
  private socket: ReconnectingWebSocket;

  constructor(url: string) {
    this.socket = new ReconnectingWebSocket(url, {
      retryDelayMs: 1500,
      maxRetries: 10,
    });
  }

  onMessage(handler: (data: ArrayBuffer) => void) {
    this.socket.onMessage = (event) => {
      if (event.data instanceof ArrayBuffer) {
        handler(event.data);
      } else if (event.data instanceof Blob) {
        event.data
          .arrayBuffer()
          .then(handler)
          .catch(() => undefined);
      } else {
        const encoder = new TextEncoder();
        handler(encoder.encode(String(event.data)).buffer);
      }
    };
  }

  onOpen(handler: () => void) {
    this.socket.onOpen = handler;
  }

  onClose(handler: (event: CloseEvent) => void) {
    this.socket.onClose = handler;
  }

  onError(handler: (event: Event) => void) {
    this.socket.onError = handler;
  }

  send(data: ArrayBufferLike | ArrayBufferView) {
    this.socket.send(data as ArrayBufferLike);
  }

  close() {
    this.socket.close();
  }
}

export class ProviderAPIClient {
  private endpoint: string;
  private timeoutMs: number;
  private retries: number;
  private retryDelayMs: number;
  private wallet?: ProviderAPIClientOptions["wallet"];
  private hmac?: ProviderAPIClientOptions["hmac"];
  private fetcher: typeof fetch;

  constructor(options: ProviderAPIClientOptions) {
    this.endpoint = options.endpoint.replace(/\/$/, "");
    this.timeoutMs = options.timeoutMs ?? 30000;
    this.retries = options.retries ?? 2;
    this.retryDelayMs = options.retryDelayMs ?? 1000;
    this.wallet = options.wallet;
    this.hmac = options.hmac;
    this.fetcher = options.fetcher ?? fetch;
  }

  async health(): Promise<ProviderHealth> {
    return this.request<ProviderHealth>("GET", "/api/v1/health", {
      requiresAuth: false,
    });
  }

  async listDeployments(
    options: { limit?: number; cursor?: string; status?: string } = {},
  ): Promise<DeploymentListResponse> {
    const response = await this.request<{
      deployments?: Deployment[];
      next_cursor?: string;
    }>("GET", "/api/v1/deployments", {
      params: {
        limit: options.limit,
        cursor: options.cursor,
        status: options.status,
      },
    });

    return {
      deployments: (response.deployments ?? []).map(this.normalizeDeployment),
      nextCursor: response.next_cursor ?? null,
    };
  }

  async getDeployment(deploymentId: string): Promise<Deployment> {
    const response = await this.request<Deployment>(
      "GET",
      `/api/v1/deployments/${deploymentId}`,
    );
    return this.normalizeDeployment(response);
  }

  async getDeploymentStatus(leaseId: string): Promise<DeploymentStatus> {
    const response = await this.request<DeploymentStatus>(
      "GET",
      `/api/v1/deployments/${leaseId}/status`,
    );
    return {
      ...response,
      leaseId: response.leaseId ?? (response as any).lease_id ?? leaseId,
      lastUpdated: normalizeDate(
        (response as any).lastUpdated ?? (response as any).last_updated,
      ),
    };
  }

  async getDeploymentLogs(
    leaseId: string,
    options?: LogOptions,
  ): Promise<string[]> {
    const query = buildQuery({
      tail: options?.tail,
      since: options?.since ? options.since.toISOString() : undefined,
      timestamps: options?.timestamps,
      level: options?.level,
      search: options?.search,
      follow: options?.follow,
    });
    const response = await this.request<string | string[]>(
      "GET",
      `/api/v1/deployments/${leaseId}/logs${query}`,
    );

    if (Array.isArray(response)) {
      return response;
    }

    if (typeof response === "string") {
      return response
        .split("\n")
        .map((line) => line.trim())
        .filter(Boolean);
    }

    return [];
  }

  async getDeploymentMetrics(leaseId: string): Promise<ResourceMetrics> {
    const response = await this.request<
      ResourceMetrics & { timestamp?: string | number }
    >("GET", `/api/v1/deployments/${leaseId}/metrics`);

    return {
      ...response,
      timestamp: normalizeDate(response.timestamp),
    };
  }

  async performAction(
    leaseId: string,
    action: DeploymentAction,
  ): Promise<{ success: boolean; message?: string }> {
    return this.request("POST", `/api/v1/deployments/${leaseId}/actions`, {
      body: { action },
    });
  }

  async createShellSession(
    leaseId: string,
    container?: string,
  ): Promise<ShellSessionResponse> {
    const response = await this.request<ShellSessionResponse>(
      "POST",
      `/api/v1/deployments/${leaseId}/shell/session`,
      {
        body: container ? { container } : {},
      },
    );

    return {
      ...response,
      expiresAt: response.expiresAt ?? (response as any).expires_at,
      sessionTtl: response.sessionTtl ?? (response as any).session_ttl,
    };
  }

  async connectLogStream(
    leaseId: string,
    options?: LogOptions,
  ): Promise<LogStream> {
    const query = buildQuery({
      tail: options?.tail,
      since: options?.since ? options.since.toISOString() : undefined,
      timestamps: options?.timestamps,
      level: options?.level,
      search: options?.search,
      follow: options?.follow,
    });
    const url = await this.buildWebSocketUrl(
      `/api/v1/deployments/${leaseId}/logs/stream${query}`,
    );
    return new LogStream(url);
  }

  async connectShell(
    leaseId: string,
    sessionToken?: string,
    container?: string,
  ): Promise<ShellConnection> {
    const query = buildQuery({
      token: sessionToken,
      container,
    });
    const url = await this.buildWebSocketUrl(
      `/api/v1/deployments/${leaseId}/shell${query}`,
    );
    return new ShellConnection(url);
  }

  private normalizeDeployment = (
    deployment: Deployment & { created_at?: string; createdAt?: string | Date },
  ): Deployment => {
    return {
      ...deployment,
      createdAt: normalizeDate(deployment.createdAt ?? deployment.created_at),
    };
  };

  private async request<T>(
    method: string,
    path: string,
    options: ProviderAPIRequestOptions = {},
  ): Promise<T> {
    const requiresAuth = options.requiresAuth ?? true;
    const url = new URL(`${this.endpoint}${path}`);
    if (options.params) {
      Object.entries(options.params).forEach(([key, value]) => {
        if (value !== undefined) {
          url.searchParams.set(key, String(value));
        }
      });
    }

    let lastError: Error | null = null;

    const bodyPayload =
      options.body === undefined
        ? undefined
        : serializeRequestBody(options.body);

    for (let attempt = 0; attempt <= this.retries; attempt += 1) {
      try {
        const headers: Record<string, string> = {
          "Content-Type": "application/json",
        };

        if (requiresAuth) {
          if (this.wallet) {
            const signed = await signRequest({
              method,
              path: url.pathname,
              body: options.body,
              signer: this.wallet.signer,
              address: this.wallet.address,
              chainId: this.wallet.chainId,
            });
            Object.assign(headers, signed);
          } else if (this.hmac) {
            const timestamp = Math.floor(Date.now() / 1000).toString();
            const query = url.searchParams.toString();
            const payload = [
              method,
              url.pathname,
              query,
              this.hmac.principal,
              timestamp,
            ].join("\n");
            const signature = await hmacSha256Hex(payload, this.hmac.secret);
            headers["X-VE-Signature"] = signature;
            headers["X-VE-Timestamp"] = timestamp;
            headers["X-VE-Principal"] = this.hmac.principal;
          }
        }

        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeoutMs);

        const response = await this.fetcher(url.toString(), {
          method,
          headers,
          body: bodyPayload,
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        const contentType = response.headers.get("content-type") ?? "";
        let payload: unknown = null;

        if (contentType.includes("application/json")) {
          payload = await response.json();
        } else {
          payload = await response.text();
        }

        if (!response.ok) {
          throw new ProviderAPIError(
            typeof payload === "string"
              ? payload
              : ((payload as ProviderAPIErrorDetails)?.message ??
                  response.statusText),
            response.status,
            payload as ProviderAPIErrorDetails,
          );
        }

        return payload as T;
      } catch (error) {
        lastError = error as Error;
        if (attempt >= this.retries || !this.shouldRetry(error)) {
          break;
        }
        await sleep(this.retryDelayMs * Math.max(1, attempt + 1));
      }
    }

    throw lastError ?? new Error("Provider API request failed");
  }

  private shouldRetry(error: unknown): boolean {
    if (error instanceof ProviderAPIError) {
      return error.status >= 500;
    }
    return true;
  }

  private async buildWebSocketUrl(path: string): Promise<string> {
    const base = this.endpoint.replace(/^http/, "ws");
    const url = new URL(`${base}${path}`);

    if (this.wallet) {
      const signed = await signRequest({
        method: "GET",
        path: url.pathname,
        signer: this.wallet.signer,
        address: this.wallet.address,
        chainId: this.wallet.chainId,
      });
      url.searchParams.set("ve_address", signed["X-VE-Address"]);
      url.searchParams.set("ve_ts", signed["X-VE-Timestamp"]);
      url.searchParams.set("ve_nonce", signed["X-VE-Nonce"]);
      url.searchParams.set("ve_sig", signed["X-VE-Signature"]);
      url.searchParams.set("ve_pub", signed["X-VE-PubKey"]);
    } else if (this.hmac) {
      const timestamp = Math.floor(Date.now() / 1000).toString();
      const query = url.searchParams.toString();
      const payload = [
        "GET",
        url.pathname,
        query,
        this.hmac.principal,
        timestamp,
      ].join("\n");
      const signature = await hmacSha256Hex(payload, this.hmac.secret);
      url.searchParams.set("sig", signature);
      url.searchParams.set("ts", timestamp);
      url.searchParams.set("principal", this.hmac.principal);
    }

    return url.toString();
  }
}

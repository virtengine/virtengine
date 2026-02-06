# Task 29D: Provider API TypeScript Client

**ID:** 29D  
**Title:** feat(portal): Provider API TypeScript client  
**Priority:** P0 (Critical)  
**Wave:** 2 (Sequential after 29C)  
**Estimated LOC:** 2000-3000  
**Dependencies:** 29C (lib/portal foundation)  
**Blocking:** 29E, 29G, 29H, 29I, 29J  

---

## Problem Statement

The provider daemon exposes a Portal API at `pkg/provider_daemon/portal_api.go` (688 lines) with REST and WebSocket endpoints, but there's **no TypeScript client** to consume these from the frontend. The hybrid architecture requires:

1. Direct communication between browser and provider instances
2. WebSocket connections for real-time logs and shell
3. HMAC-signed requests for authentication
4. Proper error handling and retries

### Current State Analysis

```
pkg/provider_daemon/portal_api.go  ✅ Complete (688 lines)
├── GET  /api/v1/health
├── GET  /api/v1/deployments/:id/logs
├── GET  /api/v1/deployments/:id/status
├── WS   /api/v1/deployments/:id/shell
├── GET  /api/v1/deployments/:id/metrics
└── POST /api/v1/deployments/:id/actions

lib/portal/src/provider-api/       ❌ MISSING
```

---

## Acceptance Criteria

### AC-1: ProviderAPIClient Class
- [ ] Create `ProviderAPIClient` TypeScript class
- [ ] Accept provider endpoint URL in constructor
- [ ] Support request timeout configuration
- [ ] Implement retry logic with exponential backoff
- [ ] Handle provider offline/unreachable states

### AC-2: REST Endpoint Methods
- [ ] `health()` - Check provider health
- [ ] `getDeploymentLogs(leaseId, opts)` - Fetch container logs
- [ ] `getDeploymentStatus(leaseId)` - Get deployment status
- [ ] `getDeploymentMetrics(leaseId)` - Get resource metrics
- [ ] `performAction(leaseId, action)` - Start/stop/restart
- [ ] All methods properly typed with request/response interfaces

### AC-3: WebSocket Connections
- [ ] `connectShell(leaseId)` - Interactive shell WebSocket
- [ ] `streamLogs(leaseId)` - Real-time log streaming
- [ ] Automatic reconnection with backoff
- [ ] Connection state management
- [ ] Proper cleanup on disconnect

### AC-4: HMAC Authentication
- [ ] Implement HMAC-SHA256 signature generation
- [ ] Sign all requests with shared secret
- [ ] Include timestamp in signature
- [ ] Handle signature verification failures

### AC-5: React Hooks
- [ ] `useProviderAPI(endpoint)` - Client instance hook
- [ ] `useDeploymentLogs(leaseId)` - Log streaming hook
- [ ] `useDeploymentShell(leaseId)` - Interactive shell hook
- [ ] `useDeploymentMetrics(leaseId)` - Metrics polling hook
- [ ] `useDeploymentStatus(leaseId)` - Status polling hook
- [ ] Proper cleanup on unmount

### AC-6: Error Handling
- [ ] Define error types for all failure modes
- [ ] Handle network errors gracefully
- [ ] Handle authentication failures
- [ ] Handle provider-side errors
- [ ] Provide meaningful error messages

### AC-7: Testing
- [ ] Unit tests with mocked HTTP responses
- [ ] Unit tests for WebSocket reconnection
- [ ] Unit tests for HMAC signing
- [ ] Integration tests against real provider daemon

---

## Technical Requirements

### ProviderAPIClient Implementation

```typescript
// lib/portal/src/provider-api/client.ts

export interface ProviderAPIClientOptions {
  endpoint: string;
  timeout?: number;
  retries?: number;
  hmacSecret?: string;
}

export interface LogOptions {
  follow?: boolean;
  tail?: number;
  since?: Date;
  timestamps?: boolean;
}

export interface DeploymentStatus {
  leaseId: string;
  state: 'pending' | 'running' | 'stopped' | 'failed';
  replicas: { ready: number; total: number };
  services: ServiceStatus[];
  lastUpdated: Date;
}

export interface ServiceStatus {
  name: string;
  state: string;
  replicas: number;
  ports: { port: number; protocol: string }[];
}

export interface ResourceMetrics {
  cpu: { usage: number; limit: number };
  memory: { usage: number; limit: number };
  storage: { usage: number; limit: number };
  network: { rxBytes: number; txBytes: number };
  timestamp: Date;
}

export class ProviderAPIClient {
  private readonly endpoint: string;
  private readonly timeout: number;
  private readonly maxRetries: number;
  private readonly hmacSecret?: string;
  
  constructor(options: ProviderAPIClientOptions) {
    this.endpoint = options.endpoint.replace(/\/$/, '');
    this.timeout = options.timeout ?? 30_000;
    this.maxRetries = options.retries ?? 3;
    this.hmacSecret = options.hmacSecret;
  }

  /**
   * Check provider health
   */
  async health(): Promise<{ status: 'ok' | 'degraded' | 'down'; version: string }> {
    return this.request<{ status: string; version: string }>('GET', '/api/v1/health');
  }

  /**
   * Get deployment status
   */
  async getDeploymentStatus(leaseId: string): Promise<DeploymentStatus> {
    return this.request<DeploymentStatus>('GET', `/api/v1/deployments/${leaseId}/status`);
  }

  /**
   * Fetch deployment logs (non-streaming)
   */
  async getDeploymentLogs(leaseId: string, options?: LogOptions): Promise<string[]> {
    const params = new URLSearchParams();
    if (options?.tail) params.set('tail', options.tail.toString());
    if (options?.since) params.set('since', options.since.toISOString());
    if (options?.timestamps) params.set('timestamps', 'true');
    
    const query = params.toString();
    const path = `/api/v1/deployments/${leaseId}/logs${query ? '?' + query : ''}`;
    
    return this.request<string[]>('GET', path);
  }

  /**
   * Get resource metrics
   */
  async getDeploymentMetrics(leaseId: string): Promise<ResourceMetrics> {
    const response = await this.request<any>('GET', `/api/v1/deployments/${leaseId}/metrics`);
    return {
      ...response,
      timestamp: new Date(response.timestamp),
    };
  }

  /**
   * Perform action on deployment (start, stop, restart)
   */
  async performAction(
    leaseId: string,
    action: 'start' | 'stop' | 'restart',
  ): Promise<{ success: boolean; message: string }> {
    return this.request('POST', `/api/v1/deployments/${leaseId}/actions`, { action });
  }

  /**
   * Connect to real-time log stream
   */
  connectLogStream(leaseId: string, options?: LogOptions): LogStream {
    const url = this.buildWebSocketUrl(`/api/v1/deployments/${leaseId}/logs/stream`);
    return new LogStream(url, this.hmacSecret);
  }

  /**
   * Connect to interactive shell
   */
  connectShell(leaseId: string, service?: string): ShellConnection {
    const path = service
      ? `/api/v1/deployments/${leaseId}/shell?service=${service}`
      : `/api/v1/deployments/${leaseId}/shell`;
    const url = this.buildWebSocketUrl(path);
    return new ShellConnection(url, this.hmacSecret);
  }

  /**
   * Internal HTTP request method with retries
   */
  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    let lastError: Error | null = null;
    
    for (let attempt = 0; attempt <= this.maxRetries; attempt++) {
      try {
        const headers: HeadersInit = {
          'Content-Type': 'application/json',
        };

        // Add HMAC signature if secret is configured
        if (this.hmacSecret) {
          const timestamp = Date.now().toString();
          const signature = this.signRequest(method, path, timestamp, body);
          headers['X-Timestamp'] = timestamp;
          headers['X-Signature'] = signature;
        }

        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeout);

        const response = await fetch(`${this.endpoint}${path}`, {
          method,
          headers,
          body: body ? JSON.stringify(body) : undefined,
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        if (!response.ok) {
          const error = await response.json().catch(() => ({}));
          throw new ProviderAPIError(
            error.message || `HTTP ${response.status}`,
            response.status,
            error.code,
          );
        }

        return response.json();
      } catch (error) {
        lastError = error as Error;
        
        // Don't retry on authentication errors
        if (error instanceof ProviderAPIError && error.status === 401) {
          throw error;
        }
        
        // Exponential backoff
        if (attempt < this.maxRetries) {
          await this.delay(Math.pow(2, attempt) * 1000);
        }
      }
    }

    throw lastError || new Error('Request failed');
  }

  private signRequest(
    method: string,
    path: string,
    timestamp: string,
    body?: unknown,
  ): string {
    const payload = [method, path, timestamp, body ? JSON.stringify(body) : ''].join('\n');
    return hmacSha256(this.hmacSecret!, payload);
  }

  private buildWebSocketUrl(path: string): string {
    const url = new URL(this.endpoint);
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
    url.pathname = path;
    return url.toString();
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

### WebSocket Connections

```typescript
// lib/portal/src/provider-api/websocket.ts

export interface WebSocketOptions {
  reconnect?: boolean;
  maxReconnectAttempts?: number;
  reconnectInterval?: number;
}

export type LogStreamCallback = (log: string) => void;

export class LogStream {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private readonly maxReconnectAttempts: number;
  private readonly reconnectInterval: number;
  private callbacks: LogStreamCallback[] = [];

  constructor(
    private readonly url: string,
    private readonly hmacSecret?: string,
    options?: WebSocketOptions,
  ) {
    this.maxReconnectAttempts = options?.maxReconnectAttempts ?? 5;
    this.reconnectInterval = options?.reconnectInterval ?? 1000;
  }

  connect(): void {
    const wsUrl = this.hmacSecret
      ? this.addAuthToUrl(this.url)
      : this.url;

    this.ws = new WebSocket(wsUrl);

    this.ws.onmessage = (event) => {
      const log = event.data as string;
      this.callbacks.forEach(cb => cb(log));
    };

    this.ws.onclose = (event) => {
      if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        setTimeout(() => this.connect(), this.reconnectInterval * this.reconnectAttempts);
      }
    };

    this.ws.onerror = (error) => {
      console.error('LogStream error:', error);
    };
  }

  onLog(callback: LogStreamCallback): void {
    this.callbacks.push(callback);
  }

  close(): void {
    this.reconnectAttempts = this.maxReconnectAttempts; // Prevent reconnection
    this.ws?.close();
    this.ws = null;
    this.callbacks = [];
  }

  private addAuthToUrl(url: string): string {
    const timestamp = Date.now().toString();
    const signature = hmacSha256(this.hmacSecret!, `WS\n${url}\n${timestamp}`);
    return `${url}${url.includes('?') ? '&' : '?'}ts=${timestamp}&sig=${signature}`;
  }
}

export class ShellConnection {
  private ws: WebSocket | null = null;
  private onDataCallback?: (data: string) => void;
  private onCloseCallback?: () => void;

  constructor(
    private readonly url: string,
    private readonly hmacSecret?: string,
  ) {}

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      const wsUrl = this.hmacSecret
        ? this.addAuthToUrl(this.url)
        : this.url;

      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => resolve();
      this.ws.onerror = (error) => reject(error);

      this.ws.onmessage = (event) => {
        this.onDataCallback?.(event.data as string);
      };

      this.ws.onclose = () => {
        this.onCloseCallback?.();
      };
    });
  }

  send(data: string): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(data);
    }
  }

  resize(cols: number, rows: number): void {
    this.send(JSON.stringify({ type: 'resize', cols, rows }));
  }

  onData(callback: (data: string) => void): void {
    this.onDataCallback = callback;
  }

  onClose(callback: () => void): void {
    this.onCloseCallback = callback;
  }

  close(): void {
    this.ws?.close();
    this.ws = null;
  }

  private addAuthToUrl(url: string): string {
    const timestamp = Date.now().toString();
    const signature = hmacSha256(this.hmacSecret!, `WS\n${url}\n${timestamp}`);
    return `${url}${url.includes('?') ? '&' : '?'}ts=${timestamp}&sig=${signature}`;
  }
}
```

### React Hooks

```typescript
// lib/portal/src/hooks/useProviderAPI.ts
import { useMemo } from 'react';
import { ProviderAPIClient, ProviderAPIClientOptions } from '../provider-api/client';

export function useProviderAPI(options: ProviderAPIClientOptions): ProviderAPIClient {
  return useMemo(() => new ProviderAPIClient(options), [options.endpoint, options.hmacSecret]);
}

// lib/portal/src/hooks/useDeploymentLogs.ts
import { useEffect, useState, useCallback } from 'react';
import { LogStream } from '../provider-api/websocket';
import { useProviderAPI } from './useProviderAPI';

export interface UseDeploymentLogsOptions {
  endpoint: string;
  leaseId: string;
  hmacSecret?: string;
  maxLines?: number;
}

export function useDeploymentLogs(options: UseDeploymentLogsOptions) {
  const [logs, setLogs] = useState<string[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  
  const client = useProviderAPI({
    endpoint: options.endpoint,
    hmacSecret: options.hmacSecret,
  });

  useEffect(() => {
    const stream = client.connectLogStream(options.leaseId);
    
    stream.onLog((log) => {
      setLogs(prev => {
        const next = [...prev, log];
        if (options.maxLines && next.length > options.maxLines) {
          return next.slice(-options.maxLines);
        }
        return next;
      });
    });

    try {
      stream.connect();
      setIsConnected(true);
    } catch (err) {
      setError(err as Error);
    }

    return () => {
      stream.close();
      setIsConnected(false);
    };
  }, [options.endpoint, options.leaseId, options.hmacSecret]);

  const clear = useCallback(() => setLogs([]), []);

  return { logs, isConnected, error, clear };
}

// lib/portal/src/hooks/useDeploymentShell.ts
import { useEffect, useRef, useState, useCallback } from 'react';
import { ShellConnection } from '../provider-api/websocket';
import { useProviderAPI } from './useProviderAPI';

export interface UseDeploymentShellOptions {
  endpoint: string;
  leaseId: string;
  service?: string;
  hmacSecret?: string;
}

export function useDeploymentShell(options: UseDeploymentShellOptions) {
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const shellRef = useRef<ShellConnection | null>(null);
  const onDataRef = useRef<((data: string) => void) | null>(null);
  
  const client = useProviderAPI({
    endpoint: options.endpoint,
    hmacSecret: options.hmacSecret,
  });

  useEffect(() => {
    const shell = client.connectShell(options.leaseId, options.service);
    shellRef.current = shell;

    shell.onData((data) => {
      onDataRef.current?.(data);
    });

    shell.onClose(() => {
      setIsConnected(false);
    });

    shell.connect()
      .then(() => setIsConnected(true))
      .catch((err) => setError(err));

    return () => {
      shell.close();
      shellRef.current = null;
    };
  }, [options.endpoint, options.leaseId, options.service, options.hmacSecret]);

  const send = useCallback((data: string) => {
    shellRef.current?.send(data);
  }, []);

  const resize = useCallback((cols: number, rows: number) => {
    shellRef.current?.resize(cols, rows);
  }, []);

  const onData = useCallback((callback: (data: string) => void) => {
    onDataRef.current = callback;
  }, []);

  return { isConnected, error, send, resize, onData };
}

// lib/portal/src/hooks/useDeploymentMetrics.ts
import { useQuery } from '@tanstack/react-query';
import { useProviderAPI } from './useProviderAPI';
import { ResourceMetrics } from '../provider-api/client';

export interface UseDeploymentMetricsOptions {
  endpoint: string;
  leaseId: string;
  hmacSecret?: string;
  pollingInterval?: number;
}

export function useDeploymentMetrics(options: UseDeploymentMetricsOptions) {
  const client = useProviderAPI({
    endpoint: options.endpoint,
    hmacSecret: options.hmacSecret,
  });

  return useQuery<ResourceMetrics>({
    queryKey: ['deployment-metrics', options.endpoint, options.leaseId],
    queryFn: () => client.getDeploymentMetrics(options.leaseId),
    refetchInterval: options.pollingInterval ?? 5000,
    staleTime: 1000,
  });
}
```

---

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `lib/portal/src/provider-api/client.ts` | Main API client | 350 |
| `lib/portal/src/provider-api/websocket.ts` | WebSocket connections | 250 |
| `lib/portal/src/provider-api/auth.ts` | HMAC signing utilities | 80 |
| `lib/portal/src/provider-api/errors.ts` | Error types | 60 |
| `lib/portal/src/provider-api/types.ts` | API type definitions | 150 |
| `lib/portal/src/provider-api/index.ts` | Module exports | 20 |
| `lib/portal/src/hooks/useProviderAPI.ts` | Client hook | 30 |
| `lib/portal/src/hooks/useDeploymentLogs.ts` | Log streaming hook | 70 |
| `lib/portal/src/hooks/useDeploymentShell.ts` | Shell hook | 80 |
| `lib/portal/src/hooks/useDeploymentMetrics.ts` | Metrics hook | 40 |
| `lib/portal/src/hooks/useDeploymentStatus.ts` | Status hook | 40 |
| `lib/portal/src/provider-api/__tests__/client.test.ts` | Client tests | 300 |
| `lib/portal/src/provider-api/__tests__/websocket.test.ts` | WebSocket tests | 200 |

**Total: ~1670 lines**

---

## Implementation Steps

### Step 1: Create Type Definitions
Define all request/response types

### Step 2: Implement HMAC Signing
Create `auth.ts` with signing utilities

### Step 3: Implement ProviderAPIClient
Create main client class with all REST methods

### Step 4: Implement WebSocket Classes
Create LogStream and ShellConnection

### Step 5: Create React Hooks
Implement all hooks with proper cleanup

### Step 6: Write Tests
Unit tests with mocked responses

### Step 7: Integration Test
Test against running provider daemon

---

## Validation Checklist

- [ ] Client can connect to provider health endpoint
- [ ] Client can fetch deployment logs
- [ ] Client can get deployment status
- [ ] WebSocket log streaming works
- [ ] WebSocket shell connection works
- [ ] HMAC signing is correct
- [ ] Retry logic works for transient failures
- [ ] Hooks properly cleanup on unmount
- [ ] All tests pass

---

## Vibe-Kanban Task ID

`98867444-1226-4fc0-9d35-bf9af06b992f`

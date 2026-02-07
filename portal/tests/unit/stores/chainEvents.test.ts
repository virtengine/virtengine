/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ChainEventClient } from '@/lib/chain-events';
import type { ChainEvent, ConnectionStatus } from '@/types/chain-events';

// ---------------------------------------------------------------------------
// Mock WebSocket
// ---------------------------------------------------------------------------

class MockWebSocket {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;

  readyState = MockWebSocket.CONNECTING;
  onopen: (() => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  onmessage: ((e: { data: string }) => void) | null = null;

  sent: string[] = [];

  constructor(_url: string) {
    // Auto-open after a tick.
    setTimeout(() => {
      this.readyState = MockWebSocket.OPEN;
      this.onopen?.();
    }, 0);
  }

  send(data: string) {
    this.sent.push(data);
  }

  close() {
    this.readyState = MockWebSocket.CLOSED;
    this.onclose?.();
  }
}

let lastMockWs: MockWebSocket | null = null;

beforeEach(() => {
  lastMockWs = null;
  vi.stubGlobal(
    'WebSocket',
    class extends MockWebSocket {
      constructor(url: string) {
        super(url);
        lastMockWs = this;
      }
    }
  );
  (globalThis as Record<string, unknown>).WebSocket = (
    globalThis as Record<string, unknown>
  ).WebSocket;
  // Set the static constants on the stub
  (globalThis.WebSocket as unknown as typeof MockWebSocket).OPEN = MockWebSocket.OPEN;
  (globalThis.WebSocket as unknown as typeof MockWebSocket).CONNECTING = MockWebSocket.CONNECTING;
});

afterEach(() => {
  vi.restoreAllMocks();
  vi.useRealTimers();
});

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('ChainEventClient', () => {
  it('connects and sets status to connected', async () => {
    const statuses: ConnectionStatus[] = [];
    const client = new ChainEventClient({ autoReconnect: false });
    client.onStatusChange((s) => statuses.push(s));

    client.connect();
    expect(statuses).toContain('connecting');

    // Wait for mock WS to open.
    await new Promise((r) => setTimeout(r, 10));
    expect(statuses).toContain('connected');
    expect(client.getStatus()).toBe('connected');

    client.disconnect();
  });

  it('sends subscribe messages for each configured event type', async () => {
    const client = new ChainEventClient({
      subscriptions: ['order.created', 'bid.created'],
      autoReconnect: false,
    });
    client.connect();
    await new Promise((r) => setTimeout(r, 10));

    expect(lastMockWs).not.toBeNull();
    expect(lastMockWs!.sent.length).toBe(2);

    const msg0 = JSON.parse(lastMockWs!.sent[0]!);
    expect(msg0.method).toBe('subscribe');
    expect(msg0.params.query).toContain('CreateOrder');

    const msg1 = JSON.parse(lastMockWs!.sent[1]!);
    expect(msg1.params.query).toContain('CreateBid');

    client.disconnect();
  });

  it('parses incoming events and notifies handlers', async () => {
    const received: ChainEvent[] = [];
    const client = new ChainEventClient({ autoReconnect: false });
    client.onEvent((e) => received.push(e));
    client.connect();
    await new Promise((r) => setTimeout(r, 10));

    const msg = JSON.stringify({
      result: {
        events: { 'tx.hash': ['ABC123'] },
        data: {
          value: {
            TxResult: {
              height: '42',
              result: {
                events: [
                  {
                    type: 'create_order',
                    attributes: [{ key: 'order_id', value: '1001' }],
                  },
                ],
              },
            },
          },
        },
      },
    });

    lastMockWs!.onmessage?.({ data: msg });

    expect(received.length).toBe(1);
    expect(received[0]!.type).toBe('order.created');
    expect(received[0]!.blockHeight).toBe(42);
    expect(received[0]!.attributes['order_id']).toBe('1001');

    client.disconnect();
  });

  it('ignores malformed messages', async () => {
    const received: ChainEvent[] = [];
    const client = new ChainEventClient({ autoReconnect: false });
    client.onEvent((e) => received.push(e));
    client.connect();
    await new Promise((r) => setTimeout(r, 10));

    lastMockWs!.onmessage?.({ data: 'not json' });
    lastMockWs!.onmessage?.({ data: '{"result":{}}' });

    expect(received.length).toBe(0);
    client.disconnect();
  });

  it('reconnects with exponential backoff', async () => {
    vi.useFakeTimers();

    const statuses: ConnectionStatus[] = [];
    const client = new ChainEventClient({
      autoReconnect: true,
      reconnectDelayMs: 100,
      maxReconnectAttempts: 3,
    });
    client.onStatusChange((s) => statuses.push(s));

    client.connect();
    await vi.advanceTimersByTimeAsync(10); // open
    expect(client.getStatus()).toBe('connected');

    // Simulate close — scheduleReconnect runs synchronously, setting reconnecting.
    lastMockWs!.close();
    expect(client.getStatus()).toBe('reconnecting');

    // First reconnect after 100ms.
    await vi.advanceTimersByTimeAsync(110);
    expect(statuses).toContain('reconnecting');

    client.disconnect();
  });

  it('disconnect cleans up and prevents reconnection', async () => {
    const client = new ChainEventClient({ autoReconnect: true });
    client.connect();
    await new Promise((r) => setTimeout(r, 10));

    client.disconnect();
    expect(client.getStatus()).toBe('disconnected');

    // No reconnect should happen.
    await new Promise((r) => setTimeout(r, 50));
    expect(client.getStatus()).toBe('disconnected');
  });

  it('onEvent returns an unsubscribe function', async () => {
    const received: ChainEvent[] = [];
    const client = new ChainEventClient({ autoReconnect: false });
    const unsub = client.onEvent((e) => received.push(e));

    client.connect();
    await new Promise((r) => setTimeout(r, 10));

    unsub();

    const msg = JSON.stringify({
      result: {
        events: { 'tx.hash': ['XYZ'] },
        data: {
          value: {
            TxResult: {
              height: '1',
              result: {
                events: [{ type: 'create_order', attributes: [] }],
              },
            },
          },
        },
      },
    });

    lastMockWs!.onmessage?.({ data: msg });
    expect(received.length).toBe(0);

    client.disconnect();
  });
});

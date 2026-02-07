/**
 * WebSocket classes unit tests
 * AC-3, AC-7 — LogStream and ShellConnection with reconnection
 */
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { LogStream, ShellConnection } from "../../src/provider-api/client";

// ---------------------------------------------------------------------------
// Mock WebSocket
// ---------------------------------------------------------------------------

interface MockWSInstance {
  url: string;
  onopen: ((ev?: Event) => void) | null;
  onmessage: ((ev: MessageEvent) => void) | null;
  onclose: ((ev: CloseEvent) => void) | null;
  onerror: ((ev: Event) => void) | null;
  send: ReturnType<typeof vi.fn>;
  close: ReturnType<typeof vi.fn>;
  readyState: number;
}

let wsInstances: MockWSInstance[] = [];

class MockWebSocket implements MockWSInstance {
  url: string;
  onopen: ((ev?: Event) => void) | null = null;
  onmessage: ((ev: MessageEvent) => void) | null = null;
  onclose: ((ev: CloseEvent) => void) | null = null;
  onerror: ((ev: Event) => void) | null = null;
  send = vi.fn();
  close = vi.fn();
  readyState = 1; // OPEN

  static readonly OPEN = 1;
  static readonly CLOSED = 3;

  constructor(url: string) {
    this.url = url;
    wsInstances.push(this);
  }
}

// Install mock before each test
beforeEach(() => {
  wsInstances = [];
  (globalThis as any).WebSocket = MockWebSocket;
});

afterEach(() => {
  delete (globalThis as any).WebSocket;
});

// ---------------------------------------------------------------------------
// LogStream Tests
// ---------------------------------------------------------------------------

describe("LogStream", () => {
  it("creates a WebSocket connection to the given URL", () => {
    const stream = new LogStream("wss://provider.example.com/logs");
    expect(wsInstances).toHaveLength(1);
    expect(wsInstances[0].url).toBe("wss://provider.example.com/logs");
    stream.close();
  });

  it("forwards messages via onMessage handler", () => {
    const handler = vi.fn();
    const stream = new LogStream("wss://p.test/logs");
    stream.onMessage(handler);

    const ws = wsInstances[0];
    ws.onmessage?.({ data: "hello world" } as MessageEvent);

    expect(handler).toHaveBeenCalledWith("hello world");
    stream.close();
  });

  it("fires onOpen callback when socket connects", () => {
    const handler = vi.fn();
    const stream = new LogStream("wss://p.test/logs");
    stream.onOpen(handler);

    wsInstances[0].onopen?.();

    expect(handler).toHaveBeenCalledTimes(1);
    stream.close();
  });

  it("fires onClose callback when socket closes", () => {
    const handler = vi.fn();
    const stream = new LogStream("wss://p.test/logs");
    stream.onClose(handler);

    const closeEvent = new Event("close") as unknown as CloseEvent;
    wsInstances[0].onclose?.(closeEvent);

    expect(handler).toHaveBeenCalledTimes(1);
    stream.close();
  });

  it("fires onError callback on socket error", () => {
    const handler = vi.fn();
    const stream = new LogStream("wss://p.test/logs");
    stream.onError(handler);

    wsInstances[0].onerror?.(new Event("error"));

    expect(handler).toHaveBeenCalledTimes(1);
    stream.close();
  });

  it("attempts reconnection on unclean close", () => {
    vi.useFakeTimers();
    try {
      const stream = new LogStream("wss://p.test/logs");
      expect(wsInstances).toHaveLength(1);

      // Simulate unclean close (server disconnect)
      const closeEvent = { wasClean: false } as unknown as CloseEvent;
      wsInstances[0].onclose?.(closeEvent);

      // Advance timer past first reconnect delay
      vi.advanceTimersByTime(5_000);

      // A new WebSocket should have been created for reconnection
      expect(wsInstances.length).toBeGreaterThanOrEqual(2);

      stream.close();
    } finally {
      vi.useRealTimers();
    }
  });

  it("stops reconnecting after close() is called", () => {
    vi.useFakeTimers();
    try {
      const stream = new LogStream("wss://p.test/logs");
      stream.close();

      expect(wsInstances[0].close).toHaveBeenCalled();

      // No new WebSocket should be created
      const count = wsInstances.length;
      vi.advanceTimersByTime(60_000);
      expect(wsInstances.length).toBe(count);
    } finally {
      vi.useRealTimers();
    }
  });
});

// ---------------------------------------------------------------------------
// ShellConnection Tests
// ---------------------------------------------------------------------------

describe("ShellConnection", () => {
  it("creates a WebSocket connection to the given URL", () => {
    const shell = new ShellConnection("wss://provider.example.com/shell");
    expect(wsInstances).toHaveLength(1);
    expect(wsInstances[0].url).toBe("wss://provider.example.com/shell");
    shell.close();
  });

  it("sends binary data to the WebSocket", () => {
    const shell = new ShellConnection("wss://p.test/shell");
    const data = new TextEncoder().encode("ls -la");

    shell.send(data);

    expect(wsInstances[0].send).toHaveBeenCalledWith(data);
    shell.close();
  });

  it("converts text messages to ArrayBuffer via onMessage", () => {
    const handler = vi.fn();
    const shell = new ShellConnection("wss://p.test/shell");
    shell.onMessage(handler);

    wsInstances[0].onmessage?.({
      data: "text output",
    } as MessageEvent);

    expect(handler).toHaveBeenCalledTimes(1);
    const received = handler.mock.calls[0][0];
    // TextEncoder.encode().buffer may be ArrayBuffer or a typed-array buffer
    const decoded = new TextDecoder().decode(new Uint8Array(received));
    expect(decoded).toBe("text output");
    shell.close();
  });

  it("passes ArrayBuffer data directly via onMessage", () => {
    const handler = vi.fn();
    const shell = new ShellConnection("wss://p.test/shell");
    shell.onMessage(handler);

    const buf = new ArrayBuffer(4);
    wsInstances[0].onmessage?.({ data: buf } as MessageEvent);

    expect(handler).toHaveBeenCalledWith(buf);
    shell.close();
  });

  it("fires onOpen and onClose callbacks", () => {
    const openHandler = vi.fn();
    const closeHandler = vi.fn();
    const shell = new ShellConnection("wss://p.test/shell");
    shell.onOpen(openHandler);
    shell.onClose(closeHandler);

    wsInstances[0].onopen?.();
    expect(openHandler).toHaveBeenCalledTimes(1);

    const closeEvent = new Event("close") as unknown as CloseEvent;
    wsInstances[0].onclose?.(closeEvent);
    expect(closeHandler).toHaveBeenCalledTimes(1);

    shell.close();
  });

  it("closes the WebSocket on close()", () => {
    const shell = new ShellConnection("wss://p.test/shell");
    shell.close();
    expect(wsInstances[0].close).toHaveBeenCalled();
  });
});

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { ChainQueryClient } from "../../src/chain/client";
import { createChainConfig } from "../../src/chain/config";

function createClient(restEndpoints: string[]) {
  const config = createChainConfig({
    chainId: "virtengine-testnet-1",
    restEndpoints,
    rpcEndpoints: [],
  });
  return new ChainQueryClient(config);
}

describe("ChainQueryClient", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("falls back to the next endpoint on 404", async () => {
    const client = createClient(["https://rest-1", "https://rest-2"]);
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response("not found", { status: 404 }))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ ok: true }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
      );

    vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

    const result = await client.getJson<{ ok: boolean }>("/status");

    expect(result.data.ok).toBe(true);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(result.endpoint).toContain("rest-2");
  });

  it("retries on retryable status codes", async () => {
    const client = new ChainQueryClient(
      createChainConfig({
        chainId: "virtengine-testnet-1",
        restEndpoints: ["https://rest-1"],
        retry: {
          maxRetries: 1,
          baseDelayMs: 0,
          maxDelayMs: 0,
          jitterMs: 0,
          retryableStatusCodes: [500],
        },
      }),
    );

    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response("server error", { status: 500 }))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ ok: true }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
      );

    vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

    const result = await client.getJson<{ ok: boolean }>("/retry");

    expect(result.data.ok).toBe(true);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
});

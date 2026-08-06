import { describe, expect, it, jest } from "@jest/globals";

import { createProviderWebSDK } from "../../src/index.web.ts";
import { TransportError } from "../../src/sdk/transport/TransportError.ts";

describe("provider web SDK contract", () => {
  it("routes gateway-backed provider queries through the public web entry", async () => {
    const fetch = jest.fn<typeof globalThis.fetch>().mockImplementation(async (url) => {
      switch ((url as URL).pathname) {
        case "/v1/node":
          return new Response(JSON.stringify({ name: "node-1" }));
        case "/v1/inventory":
          return new Response(JSON.stringify({ nodes: [{ name: "node-1" }] }));
        case "/v1/status":
          return new Response(JSON.stringify({ errors: [], public_hostnames: ["provider.example.com"] }));
        default:
          throw new Error(`Unexpected request path: ${(url as URL).pathname}`);
      }
    });
    const sdk = createProviderWebSDK({
      baseUrl: "https://provider.example.com",
      transportOptions: { fetch },
    });

    const [node, cluster, status] = await Promise.all([
      sdk.virtengine.inventory.v1.queryNode({}),
      sdk.virtengine.inventory.v1.queryCluster({}),
      sdk.virtengine.provider.v1.getStatus({}),
    ]);

    expect(node).toEqual(expect.objectContaining({ name: "node-1" }));
    expect(cluster).toEqual(expect.objectContaining({ nodes: [expect.objectContaining({ name: "node-1" })] }));
    expect(status).toEqual(expect.objectContaining({ publicHostnames: ["provider.example.com"] }));
    expect(fetch.mock.calls.map(([url]) => (url as URL).pathname)).toEqual([
      "/v1/node",
      "/v1/inventory",
      "/v1/status",
    ]);
  });

  it("applies retry handling to gateway-backed provider methods", async () => {
    const fetch = jest.fn<typeof globalThis.fetch>()
      .mockRejectedValueOnce(new TransportError("temporary outage", TransportError.Code.Unavailable))
      .mockResolvedValueOnce(new Response(JSON.stringify({ errors: [] })));
    const sdk = createProviderWebSDK({
      baseUrl: "https://provider.example.com",
      transportOptions: {
        fetch,
        retry: { maxAttempts: 2, maxDelayMs: 1 },
      },
    });

    const status = await sdk.virtengine.provider.v1.getStatus({});

    expect(status).toEqual(expect.objectContaining({ errors: [] }));
    expect(fetch).toHaveBeenCalledTimes(2);
  });

  it("fails closed for provider methods that require direct gRPC or streaming", async () => {
    const fetch = jest.fn<typeof globalThis.fetch>();
    const sdk = createProviderWebSDK({
      baseUrl: "https://provider.example.com",
      transportOptions: { fetch },
    });

    await expect(sdk.virtengine.provider.lease.v1.sendManifest({})).rejects.toMatchObject({
      code: TransportError.Code.Unimplemented,
      message: expect.stringContaining("Use createProviderSDK() for direct gRPC access"),
    });
    await expect(
      sdk.virtengine.provider.lease.v1.streamServiceLogs({}).then(collectAsyncIterable),
    ).rejects.toMatchObject({
      code: TransportError.Code.Unimplemented,
      message: expect.stringContaining("Use createProviderSDK() for gRPC streaming access"),
    });
    expect(fetch).not.toHaveBeenCalled();
  });
});

async function collectAsyncIterable<T>(stream: AsyncIterable<T>) {
  const values: T[] = [];

  for await (const value of stream) {
    values.push(value);
  }

  return values;
}

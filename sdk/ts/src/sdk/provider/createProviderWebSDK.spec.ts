import { describe, expect, it, jest } from "@jest/globals";

import { TransportError } from "../transport/TransportError.ts";
import { createProviderWebSDK } from "./createProviderWebSDK.ts";

describe(createProviderWebSDK.name, () => {
  it("exposes the generated provider contract while routing gateway-backed unary methods over HTTP", async () => {
    const fetch = jest.fn<typeof globalThis.fetch>().mockResolvedValue(
      new Response(JSON.stringify({ name: "node-1" })),
    );
    const sdk = createProviderWebSDK({
      baseUrl: "https://provider.example.com",
      transportOptions: { fetch },
    });

    const node = await sdk.virtengine.inventory.v1.queryNode({});

    expect(sdk.virtengine.inventory.v1.queryCluster).toBeInstanceOf(Function);
    expect(sdk.virtengine.provider.lease.v1.sendManifest).toBeInstanceOf(Function);
    expect(sdk.virtengine.provider.v1.getStatus).toBeInstanceOf(Function);
    expect(node).toEqual(expect.objectContaining({ name: "node-1" }));
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(fetch.mock.calls[0][0]).toEqual(expect.objectContaining({ pathname: "/v1/node" }));
  });

  it("rejects provider lease methods that do not expose an HTTP gateway route", async () => {
    const fetch = jest.fn<typeof globalThis.fetch>();
    const sdk = createProviderWebSDK({
      baseUrl: "https://provider.example.com",
      transportOptions: { fetch },
    });

    await expect(sdk.virtengine.provider.lease.v1.sendManifest({})).rejects.toMatchObject({
      code: TransportError.Code.Unimplemented,
      message: expect.stringContaining("does not expose an HTTP gateway route"),
    });
    expect(fetch).not.toHaveBeenCalled();
  });

  it("rejects provider streaming methods for browser transports", async () => {
    const fetch = jest.fn<typeof globalThis.fetch>();
    const sdk = createProviderWebSDK({
      baseUrl: "https://provider.example.com",
      transportOptions: { fetch },
    });

    await expect(
      sdk.virtengine.provider.v1.streamStatus({}).then(collectAsyncIterable),
    ).rejects.toMatchObject({
      code: TransportError.Code.Unimplemented,
      message: expect.stringContaining("streaming provider methods are not available"),
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

import { describe, expect, it } from "@jest/globals";

import { createProviderSDK } from "./createProviderSDK.ts";

describe(createProviderSDK.name, () => {
  it("exposes inventory, lease, and provider services from the generated contract", () => {
    const sdk = createProviderSDK({
      baseUrl: "http://localhost:1317",
      transportOptions: {
        pingIdleConnection: true,
        pingIntervalMs: 1000,
        pingTimeoutMs: 1000,
        idleConnectionTimeoutMs: 1000,
        defaultTimeoutMs: 1000,
      },
    });

    expect(sdk.virtengine.inventory.v1.queryNode).toBeInstanceOf(Function);
    expect(sdk.virtengine.inventory.v1.queryCluster).toBeInstanceOf(Function);
    expect(sdk.virtengine.provider.lease.v1.sendManifest).toBeInstanceOf(Function);
    expect(sdk.virtengine.provider.lease.v1.serviceStatus).toBeInstanceOf(Function);
    expect(sdk.virtengine.provider.v1.getStatus).toBeInstanceOf(Function);
    expect(sdk.virtengine.provider.v1.streamStatus).toBeInstanceOf(Function);
  });
});

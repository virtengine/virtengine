import { describe, expect, it } from "@jest/globals";

import { createProviderSDK } from "./createProviderSDK.ts";

describe(createProviderSDK.name, () => {
  it("creates ProviderSDK", () => {
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

    expect(sdk).toBeDefined();
  });
});

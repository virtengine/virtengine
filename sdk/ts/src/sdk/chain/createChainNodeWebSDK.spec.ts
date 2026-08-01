import { describe, expect, it } from "@jest/globals";
import { mock } from "jest-mock-extended";

import type { TxClient } from "../transport/tx/TxClient.ts";
import { ChainCapability, ChainCapabilityError, ChainCapabilityErrorReason } from "./ChainCapability.ts";
import { createChainNodeWebSDK } from "./createChainNodeWebSDK.ts";

describe(createChainNodeWebSDK.name, () => {
  it("creates ChainNodeSDK with tx transport", () => {
    const sdk = createChainNodeWebSDK({
      query: { baseUrl: "http://localhost:1317" },
      tx: {
        signer: mock<TxClient>(),
      },
    });

    expect(sdk.virtengine).toBeDefined();
    expect(sdk.cosmos).toBeDefined();
    expect(sdk.capability.state).toBe(ChainCapability.SigningReady);
  });

  it("creates a query-only SDK and rejects transactions with a typed capability error", async () => {
    const sdk = createChainNodeWebSDK({
      query: { baseUrl: "http://localhost:1317" },
    });

    expect(sdk.virtengine).toBeDefined();
    expect(sdk.cosmos).toBeDefined();
    expect(sdk.capability.state).toBe(ChainCapability.QueryOnly);
    const transaction = sdk.virtengine.provider.v1beta4.createProvider({
      attributes: [],
      hostUri: "http://localhost:26657",
      info: undefined,
      owner: "virt1...",
    });
    await expect(transaction).rejects.toBeInstanceOf(ChainCapabilityError);
    await expect(transaction).rejects.toMatchObject({
      reason: ChainCapabilityErrorReason.SigningRequired,
      currentCapability: ChainCapability.QueryOnly,
      requiredCapability: ChainCapability.SigningReady,
      operation: "virtengine.provider.v1beta4.Msg.CreateProvider",
    });
  });

  it("rejects gateway queries after disconnect", async () => {
    const sdk = createChainNodeWebSDK({
      query: { baseUrl: "http://localhost:1317" },
    });

    sdk.capability.disconnect();

    await expect(sdk.cosmos.base.tendermint.v1beta1.getNodeInfo()).rejects.toMatchObject({
      reason: ChainCapabilityErrorReason.Disconnected,
      currentCapability: ChainCapability.Disconnected,
      requiredCapability: ChainCapability.QueryOnly,
      operation: "cosmos.base.tendermint.v1beta1.Service.GetNodeInfo",
    });
  });
});

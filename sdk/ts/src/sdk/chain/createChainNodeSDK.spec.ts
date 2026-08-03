import { describe, expect, it } from "@jest/globals";
import { mock } from "jest-mock-extended";

import type { TxClient } from "../transport/tx/TxClient.ts";
import { ChainCapability, ChainCapabilityError, ChainCapabilityErrorReason } from "./ChainCapability.ts";
import { createChainNodeSDK } from "./createChainNodeSDK.ts";

describe(createChainNodeSDK.name, () => {
  it("creates ChainNodeSDK with tx transport", () => {
    const sdk = createChainNodeSDK({
      query: { baseUrl: "http://localhost:1317" },
      tx: {
        signer: mock<TxClient>(),
      },
    });

    expect(sdk.virtengine).toBeDefined();
    expect(sdk.cosmos).toBeDefined();
    expect(sdk.capability.state).toBe(ChainCapability.SigningReady);
  });

  it("submits transactions through the injected signer", async () => {
    const signer = mock<TxClient>();
    signer.signAndBroadcast.mockResolvedValue({
      height: 1,
      txIndex: 0,
      code: 0,
      transactionHash: "committed-transaction-hash",
      events: [],
      msgResponses: [],
      gasUsed: 1n,
      gasWanted: 1n,
    });
    const sdk = createChainNodeSDK({
      query: { baseUrl: "http://localhost:1317" },
      tx: { signer },
    });

    await sdk.virtengine.provider.v1beta4.createProvider({
      attributes: [],
      hostUri: "http://localhost:26657",
      info: undefined,
      owner: "virt1...",
    });

    expect(signer.signAndBroadcast).toHaveBeenCalledTimes(1);
    expect(signer.signAndBroadcast).toHaveBeenCalledWith(
      [expect.objectContaining({ typeUrl: "/virtengine.provider.v1beta4.MsgCreateProvider" })],
      expect.objectContaining({ memo: "virtengine: CreateProvider" }),
    );
  });

  it("creates a query-only SDK and rejects transactions with a typed capability error", async () => {
    const sdk = createChainNodeSDK({
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

  it("rejects queries after disconnect and clears the signer", async () => {
    const sdk = createChainNodeSDK({
      query: { baseUrl: "http://localhost:1317" },
      tx: { signer: mock<TxClient>() },
    });

    sdk.capability.authorizeMFA({
      authorizationId: "mfa-authorization-1",
      expiresAt: Date.now() + 60_000,
    });
    expect(sdk.capability.state).toBe(ChainCapability.MfaAuthorized);
    sdk.capability.disconnect();

    expect(sdk.capability.state).toBe(ChainCapability.Disconnected);
    await expect(sdk.cosmos.base.tendermint.v1beta1.getNodeInfo()).rejects.toMatchObject({
      reason: ChainCapabilityErrorReason.Disconnected,
      currentCapability: ChainCapability.Disconnected,
      requiredCapability: ChainCapability.QueryOnly,
      operation: "cosmos.base.tendermint.v1beta1.Service.GetNodeInfo",
    });
  });
});

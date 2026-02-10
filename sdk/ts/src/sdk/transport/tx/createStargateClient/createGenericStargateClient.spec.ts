import type {
  EncodeObject,
  OfflineSigner,
} from "@cosmjs/proto-signing";
import type { SigningStargateClient, StdFee } from "@cosmjs/stargate";
import { describe, expect, it, jest } from "@jest/globals";
import { mock } from "jest-mock-extended";

import { createGenericStargateClient, type StargateTxClient } from "./createGenericStargateClient.ts";

describe(createGenericStargateClient.name, () => {
  const MESSAGE_TYPE = "/test.type";

  describe("sign", () => {
    includeSigningTests(async (client) => {
      const messages: EncodeObject[] = [{ typeUrl: MESSAGE_TYPE, value: {} }];
      const fee: StdFee = { amount: [], gas: "1000" };
      await client.sign(messages, fee, "test memo");
    });
  });

  describe("estimateFee", () => {
    includeSigningTests(async (client) => {
      const messages: EncodeObject[] = [{ typeUrl: MESSAGE_TYPE, value: {} }];
      await client.estimateFee(messages, "test memo");
    });

    it("uses specified gas multiplier", async () => {
      const client = createGenericStargateClient({
        baseUrl: "https://rpc.virtengine.network",
        signer: createOfflineSigner(),
        gasMultiplier: 2,
        getMessageType: () => ({
          typeUrl: MESSAGE_TYPE,
          encode: () => new Uint8Array(0),
          decode: () => ({}),
          fromPartial: () => ({}),
        }),
        createClient: async () => mock<SigningStargateClient>({
          simulate: jest.fn(async () => 1),
        } as unknown as SigningStargateClient),
      });

      const messages: EncodeObject[] = [{ typeUrl: MESSAGE_TYPE, value: {} }];
      const fee = await client.estimateFee(messages, "test memo");
      expect(fee.gas).toBe("2");
    });

    it("floors the final gas value", async () => {
      const client = createGenericStargateClient({
        baseUrl: "https://rpc.virtengine.network",
        signer: createOfflineSigner(),
        gasMultiplier: 1.9,
        getMessageType: () => ({
          typeUrl: MESSAGE_TYPE,
          encode: () => new Uint8Array(0),
          decode: () => ({}),
          fromPartial: () => ({}),
        }),
        createClient: async () => mock<SigningStargateClient>({
          simulate: jest.fn(async () => 2),
        } as unknown as SigningStargateClient),
      });

      const messages: EncodeObject[] = [{ typeUrl: MESSAGE_TYPE, value: {} }];
      const fee = await client.estimateFee(messages, "test memo");
      expect(fee.gas).toBe("3"); // 1.9 * 2 = 3.8, floored to 3
    });
  });

  function includeSigningTests(sign: (client: StargateTxClient) => Promise<unknown>) {
    it("does not calls `getMessageType` when signing message with types that are already registered", async () => {
      const getMessageType = jest.fn(() => {
        throw new Error("no types");
      });
      const client = createGenericStargateClient({
        baseUrl: "https://rpc.virtengine.network",
        signer: createOfflineSigner(),
        builtInTypes: [{
          typeUrl: MESSAGE_TYPE,
          encode: () => new Uint8Array(0),
          decode: () => ({}),
          fromPartial: () => ({}),
        }],
        getMessageType,
        createClient: async () => ({
          sign: jest.fn(),
          simulate: jest.fn(async () => 1),
          broadcastTx: jest.fn(),
        } as unknown as SigningStargateClient),
      });

      await sign(client);

      expect(getMessageType).not.toHaveBeenCalled();
    });

    it("calls `getMessageType` when signing message with types that are not registered", async () => {
      const getMessageType = jest.fn(() => ({
        typeUrl: MESSAGE_TYPE,
        encode: () => new Uint8Array(0),
        decode: () => ({}),
        fromPartial: () => ({}),
      }));
      const client = createGenericStargateClient({
        baseUrl: "https://rpc.virtengine.network",
        signer: createOfflineSigner(),
        getMessageType,
        createClient: async () => ({
          sign: jest.fn(),
          simulate: jest.fn(async () => 1),
          broadcastTx: jest.fn(),
        } as unknown as SigningStargateClient),
      });

      await sign(client);

      expect(getMessageType).toHaveBeenCalledWith(MESSAGE_TYPE);
    });
  }

  function createOfflineSigner(): OfflineSigner {
    return {
      getAccounts: async () => [{
        address: "test-address",
        algo: "secp256k1",
        pubkey: new Uint8Array(),
      }],
      signDirect: async (_, signDoc) => ({
        signed: signDoc,
        signature: {
          pub_key: {
            type: "tendermint/PubKeySecp256k1",
            value: new Uint8Array(0),
          },
          signature: "test-signature",
        },
      }),
    };
  }
});

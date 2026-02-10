import type { DescMessage, DescMethodBiDiStreaming, DescMethodUnary } from "@bufbuild/protobuf";
import type { DeliverTxResponse, StdFee } from "@cosmjs/stargate";
import { describe, expect, it, jest } from "@jest/globals";

import { proto } from "../../../../test/helpers/proto.ts";
import { createAsyncIterable } from "../../client/stream.ts";
import { createTxTransport } from "./createTxTransport.ts";
import type { TxClient, TxRaw, TxSignAndBroadcastOptions } from "./TxClient.ts";
import { TxError } from "./TxError.ts";

describe(createTxTransport.name, () => {
  describe("stream", () => {
    it("throws when `stream` method is called", async () => {
      const { TestServiceSchema } = await setup();
      const transport = createTxTransport({
        client: createMockTxClient(),
        getMessageType,
      });

      await expect(transport.stream(TestServiceSchema.methods.streamMethod, createAsyncIterable([{ test: "input" }]))).rejects.toThrow(/unimplemented/i);
    });
  });

  describe("unary", () => {
    it("signs and broadcasts a transaction with single message", async () => {
      const { TestServiceSchema } = await setup();
      const txResponse: DeliverTxResponse = {
        height: 1,
        txIndex: 0,
        code: 0,
        transactionHash: "123",
        events: [],
        msgResponses: [],
        gasUsed: 1n,
        gasWanted: 1n,
      };
      const txRaw: TxRaw = {
        bodyBytes: new Uint8Array(0),
        authInfoBytes: new Uint8Array(0),
        signatures: [],
      };
      const afterSign = jest.fn();
      const afterBroadcast = jest.fn();
      const fee: StdFee = {
        amount: [{ denom: "uve", amount: "100000" }],
        gas: "100000",
      };
      const client = createMockTxClient({
        signAndBroadcast: jest.fn((_, options?: TxSignAndBroadcastOptions) => {
          options?.afterSign?.(txRaw);
          return Promise.resolve(txResponse);
        }),
      });
      const transport = createTxTransport({
        client,
        getMessageType,
      });

      const result = await transport.unary(TestServiceSchema.methods.testMethod, { test: "input" }, {
        memo: "test",
        fee,
        afterSign,
        afterBroadcast,
      });
      const messages = [{
        typeUrl: `/${TestServiceSchema.methods.testMethod.input.$type}`,
        value: { test: "input" },
      }];

      expect(client.signAndBroadcast).toHaveBeenCalledWith(messages, {
        afterSign,
        fee,
        memo: "test",
      });
      expect(afterSign).toHaveBeenCalledWith(txRaw);
      expect(afterBroadcast).toHaveBeenCalledWith(txResponse);
      expect(result.message).toEqual({});
    });

    it("uses default memo if not provided", async () => {
      const { TestServiceSchema } = await setup();
      const client = createMockTxClient();
      const transport = createTxTransport({
        client,
        getMessageType,
      });

      await transport.unary(TestServiceSchema.methods.testMethod, { test: "input" });

      expect(client.signAndBroadcast).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({ memo: "virtengine: TestMethod" }),
      );
    });

    it("decodes response if `msgResponses` has a response", async () => {
      const { TestServiceSchema } = await setup();
      const client = createMockTxClient({
        signAndBroadcast: jest.fn(() => Promise.resolve({
          height: 1,
          txIndex: 0,
          code: 0,
          transactionHash: "123",
          events: [],
          msgResponses: [{
            typeUrl: `/${TestServiceSchema.methods.testMethod.output.$type}`,
            value: new Uint8Array(0),
          }],
          gasUsed: 1n,
          gasWanted: 1n,
        })),
      });
      const transport = createTxTransport({
        client,
        getMessageType: (typeUrl) => ({
          ...getMessageType(),
          decode: jest.fn(() => (typeUrl === `/${TestServiceSchema.methods.testMethod.output.$type}` ? { test: "output", ok: true } : null)),
        }),
      });

      const result = await transport.unary(TestServiceSchema.methods.testMethod, { test: "input" });
      expect(result.message).toEqual({ test: "output", ok: true });
    });

    it("throws if tx response has a non-zero code", async () => {
      const { TestServiceSchema } = await setup();
      const client = createMockTxClient({
        signAndBroadcast: jest.fn(() => Promise.resolve({
          height: 1,
          txIndex: 0,
          code: 1,
          transactionHash: "123",
          events: [],
          msgResponses: [],
          gasUsed: 1n,
          gasWanted: 1n,
        })),
      });
      const transport = createTxTransport({
        client,
        getMessageType,
      });

      await expect(transport.unary(TestServiceSchema.methods.testMethod, { test: "input" })).rejects.toThrow(TxError);
    });
  });

  async function setup() {
    const def = await proto`
      service TestService {
        rpc TestMethod(TestInput) returns (TestOutput);
        rpc StreamMethod(stream TestInput) returns (stream TestOutput);
      }

      message TestInput {
        string test = 1;
      }

      message TestOutput {
        string result = 1;
      }
    `;
    const TestServiceSchema = def.getTsProtoService<{
      testMethod: DescMethodUnary<DescMessage, DescMessage>;
      streamMethod: DescMethodBiDiStreaming<DescMessage, DescMessage>;
    }>("TestService");

    return { TestServiceSchema };
  }

  function createMockTxClient(overrides?: Partial<TxClient>): TxClient {
    return {
      signAndBroadcast: jest.fn(() => Promise.resolve({
        height: 1,
        txIndex: 0,
        code: 0,
        transactionHash: "123",
        events: [],
        msgResponses: [],
        gasUsed: 1n,
        gasWanted: 1n,
      })),
      ...overrides,
    };
  }

  function getMessageType() {
    return {
      decode: jest.fn(),
      encode: jest.fn(),
      fromPartial: jest.fn(),
    };
  }
});

import type { DescMethodStreaming, DescMethodUnary } from "@bufbuild/protobuf";
import type { UniversalClientFn } from "@connectrpc/connect/protocol";
import { describe, expect, it, jest } from "@jest/globals";

import { proto } from "../../../../test/helpers/proto.ts";
import { createAsyncIterable } from "../../client/stream.ts";
import type { RetryOptions } from "../interceptors/retry.ts";
import { createRetryInterceptor, isRetryEnabled } from "../interceptors/retry.ts";
import { TransportError } from "../TransportError.ts";
import { createGrpcTransport, type GrpcTransportOptions } from "./createGrpcTransport.ts";

describe(createGrpcTransport.name, () => {
  it("has `requiresTypePatching` set to true", async () => {
    const transport = createGrpcTransport({ baseUrl: "https://api.example.com" });
    expect(transport.requiresTypePatching).toBe(true);
  });

  describe("unary method", () => {
    it("makes POST request to correct URL", async () => {
      const { transport, httpClient, TestMethodSchema } = await setup();
      const message = { userId: "123", status: "active" };

      await transport.unary(TestMethodSchema, message);

      expect(httpClient).toHaveBeenCalledWith(
        expect.objectContaining({
          url: "https://api.example.com/virtengine.test.unit.TestService/TestMethod",
          method: "POST",
        }),
      );
    });

    it("serializes message in request body", async () => {
      const { transport, httpClient, TestMethodSchema } = await setup();
      const message = { userId: "123", status: "active" };

      await transport.unary(TestMethodSchema, message);

      expect(httpClient).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            [Symbol.asyncIterator]: expect.any(Function),
          }),
        }),
      );
    });

    it("sets default grpc headers and custom headers from call options", async () => {
      const { transport, httpClient, TestMethodSchema } = await setup();
      const message = { test: "data" };
      const callOptions = {
        header: { "X-Custom-Header": "custom-value" },
      };

      await transport.unary(TestMethodSchema, message, callOptions);

      const [options] = httpClient.mock.calls[0];
      const headers = options?.header as Headers;
      expect(headers.get("Content-Type")).toBe("application/grpc+proto");
      expect(headers.get("X-Custom-Header")).toBe("custom-value");
    });

    it("handles timeout from call options", async () => {
      const { transport, httpClient, TestMethodSchema } = await setup();
      const message = { test: "data" };

      await transport.unary(TestMethodSchema, message, { timeoutMs: 2000 });

      const [options] = httpClient.mock.calls[0];
      const headers = options?.header as Headers;
      expect(headers.get("Grpc-Timeout")).toBe("2000m");
    });

    it("uses default timeout when no call timeout provided", async () => {
      const { transport, httpClient, TestMethodSchema } = await setup({ defaultTimeoutMs: 5000 });
      const message = { test: "data" };

      await transport.unary(TestMethodSchema, message);

      const [options] = httpClient.mock.calls[0];
      const headers = options?.header as Headers;
      expect(headers.get("Grpc-Timeout")).toBe("5000m");
    });

    it("handles abort signal from call options", async () => {
      const { transport, httpClient, TestMethodSchema } = await setup();
      const message = { test: "data" };
      const controller = new AbortController();

      await transport.unary(TestMethodSchema, message, { signal: controller.signal });

      expect(httpClient).toHaveBeenCalledWith(
        expect.objectContaining({
          signal: expect.any(AbortSignal),
        }),
      );
    });

    it("returns correctly formatted unary response", async () => {
      const { transport, TestMethodSchema } = await setup();
      const message = { test: "data" };

      const result = await transport.unary(TestMethodSchema, message);

      expect(result.header).toBeInstanceOf(Headers);
      expect(result.trailer).toBeInstanceOf(Headers);
      expect(result.message).toBeDefined();
      expect(result.method).toBe(TestMethodSchema);
      expect(result.stream).toBe(false);
    });

    it("throws error when response contains extra messages", async () => {
      const { transport, TestMethodSchema } = await setup({ multipleMessages: true });
      const message = { test: "data" };

      await expect(transport.unary(TestMethodSchema, message)).rejects.toThrow(
        expect.objectContaining({
          message: expect.stringContaining("extra output message"),
        }),
      );
    });

    it("throws error when response contains no message", async () => {
      const { transport, TestMethodSchema } = await setup({ emptyResponse: true });
      const message = { test: "data" };

      await expect(transport.unary(TestMethodSchema, message)).rejects.toThrow(
        expect.objectContaining({
          message: expect.stringContaining("missing output message"),
        }),
      );
    });

    describe("with retry interceptor", () => {
      it("retries failed requests up to maxAttempts", async () => {
        const { transport, httpClient, TestMethodSchema } = await setup({ retryOptions: { maxAttempts: 3, maxDelayMs: 10 } });
        const message = { test: "data" };
        const responseMessageBytes = TestMethodSchema.output.encode({ result: "success" }).finish();

        httpClient
          .mockRejectedValueOnce(new TransportError("Network error"))
          .mockRejectedValueOnce(new TransportError("Network error"))
          .mockRejectedValueOnce(new TransportError("Network error"))
          .mockImplementationOnce(createMockHttpClient(responseMessageBytes));

        const result = await transport.unary(TestMethodSchema, message);

        expect(httpClient).toHaveBeenCalledTimes(4);
        expect(result.message).toBeDefined();
      });

      it("throws after exhausting all retry attempts", async () => {
        const { transport, httpClient, TestMethodSchema } = await setup({ retryOptions: { maxAttempts: 3, maxDelayMs: 10 } });
        const message = { test: "data" };

        httpClient
          .mockRejectedValueOnce(new TransportError("Network error 1"))
          .mockRejectedValueOnce(new TransportError("Network error 2"))
          .mockRejectedValueOnce(new TransportError("Network error 3"))
          .mockRejectedValueOnce(new TransportError("Network error 4"));

        await expect(transport.unary(TestMethodSchema, message)).rejects.toThrow("Network error 4");
        expect(httpClient).toHaveBeenCalledTimes(4);
      });

      it("succeeds on first attempt without retry", async () => {
        const { transport, httpClient, TestMethodSchema } = await setup({ retryOptions: { maxAttempts: 3, maxDelayMs: 10 } });
        const message = { test: "data" };

        const result = await transport.unary(TestMethodSchema, message);

        expect(httpClient).toHaveBeenCalledTimes(1);
        expect(result.message).toBeDefined();
      });

      it("uses max 3 attempts when maxAttempts is bigger than 3", async () => {
        const { transport, httpClient, TestMethodSchema } = await setup({ retryOptions: { maxAttempts: 5, maxDelayMs: 10 } });
        const message = { test: "data" };
        const responseMessageBytes = TestMethodSchema.output.encode({ result: "success" }).finish();

        httpClient
          .mockRejectedValueOnce(new TransportError("Network error", TransportError.Code.Unavailable))
          .mockRejectedValueOnce(new TransportError("Network error", TransportError.Code.Internal))
          .mockRejectedValueOnce(new TransportError("Network error", TransportError.Code.DeadlineExceeded))
          .mockRejectedValueOnce(new TransportError("Network error 4", TransportError.Code.Unknown))
          .mockImplementationOnce(createMockHttpClient(responseMessageBytes));

        await expect(transport.unary(TestMethodSchema, message)).rejects.toThrow("Network error 4");
        expect(httpClient).toHaveBeenCalledTimes(4);
      });
    });

    async function setup(
      input?: Partial<GrpcTransportOptions> & {
        multipleMessages?: boolean;
        emptyResponse?: boolean;
        headerError?: boolean;
        retryOptions?: RetryOptions;
      },
    ) {
      const { TestMethodSchema } = await setupMethod();
      const responseMessageBytes = TestMethodSchema.output.encode({ result: "success" }).finish();
      const { multipleMessages, emptyResponse, headerError, retryOptions, ...transportOptions } = input ?? {};

      const httpClient = jest.fn<UniversalClientFn>(createMockHttpClient(responseMessageBytes, {
        multipleMessages, emptyResponse, headerError,
      }));
      const options: GrpcTransportOptions = {
        baseUrl: "https://api.example.com",
        httpClient,
        defaultTimeoutMs: undefined,
        ...transportOptions,
      };

      if (isRetryEnabled(retryOptions)) {
        options.interceptors = [createRetryInterceptor(retryOptions)];
      }
      const transport = createGrpcTransport(options);

      return { transport, httpClient, TestMethodSchema };
    }

    async function setupMethod() {
      const def = await proto`
        service TestService {
          rpc TestMethod(TestInput) returns (TestOutput);
        }

        message TestInput {
          string userId = 1;
          string resourceId = 2;
          string status = 3;
          string name = 4;
          int32 value = 5;
          string test = 6;
        }

        message TestOutput {
          string result = 1;
          string id = 2;
          bool created = 3;
        }
      `;
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const TestInputSchema = def.getMessage<"TestInput", { userId: string; resourceId: string; status: string; name: string; value: number; test: string }>("TestInput");
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const TestOutputSchema = def.getMessage<"TestOutput", { result: string; id: string; created: boolean }>("TestOutput");
      const TestServiceSchema = def.getTsProtoService<{
        testMethod: DescMethodUnary<typeof TestInputSchema, typeof TestOutputSchema>;
      }>("TestService");

      const TestMethodSchema = {
        ...TestServiceSchema.methods.testMethod,
        kind: "unary" as const,
      };

      return { TestMethodSchema };
    }
  });

  describe("stream method", () => {
    it("makes POST request to correct URL", async () => {
      const { transport, httpClient, TestStreamMethodSchema } = await setup();

      await transport.stream(TestStreamMethodSchema, createAsyncIterable([{ test: "data" }]));

      expect(httpClient).toHaveBeenCalledWith(
        expect.objectContaining({
          url: "https://api.example.com/virtengine.test.unit.TestService/TestStreamMethod",
          method: "POST",
        }),
      );
    });

    it("sets default grpc headers and custom headers from call options", async () => {
      const { transport, httpClient, TestStreamMethodSchema } = await setup();
      const callOptions = {
        header: { "X-Custom-Header": "custom-value" },
      };

      await transport.stream(TestStreamMethodSchema, createAsyncIterable([{ test: "data" }]), callOptions);

      const [options] = httpClient.mock.calls[0];
      const headers = options?.header as Headers;
      expect(headers.get("X-Custom-Header")).toBe("custom-value");
      expect(headers.get("Content-Type")).toBe("application/grpc+proto");
    });

    it("handles timeout from call options", async () => {
      const { transport, httpClient, TestStreamMethodSchema } = await setup();
      const callOptions = { timeoutMs: 2000 };

      await transport.stream(TestStreamMethodSchema, createAsyncIterable([{ test: "data" }]), callOptions);

      const [options] = httpClient.mock.calls[0];
      const headers = options?.header as Headers;
      expect(headers.get("Grpc-Timeout")).toBe("2000m");
    });

    it("returns correctly formatted stream response", async () => {
      const { transport, TestStreamMethodSchema } = await setup();

      const result = await transport.stream(TestStreamMethodSchema, createAsyncIterable([{ test: "data" }]));

      expect(result.header).toBeInstanceOf(Headers);
      expect(result.trailer).toBeInstanceOf(Headers);
      expect(result.message).toBeDefined();
      expect(result.method).toBe(TestStreamMethodSchema);
      expect(result.stream).toBe(true);
    });

    it("yields multiple messages from stream", async () => {
      const { transport, TestStreamMethodSchema } = await setup({ multipleMessages: true });

      const result = await transport.stream(TestStreamMethodSchema, createAsyncIterable([{ test: "data" }]));

      const messages: unknown[] = [];
      for await (const message of result.message) {
        messages.push(message);
      }

      expect(messages.length).toBe(2);
    });

    it("throws error on header error", async () => {
      const { transport, TestStreamMethodSchema } = await setup({ headerError: true });

      await expect(transport.stream(TestStreamMethodSchema, createAsyncIterable([{ test: "data" }]))).rejects.toThrow(TransportError);
    });

    describe("with retry interceptor", () => {
      it("retries failed requests up to maxAttempts", async () => {
        const { transport, httpClient, TestStreamMethodSchema } = await setup({ retryOptions: { maxAttempts: 3, maxDelayMs: 10 } });
        const responseMessageBytes = TestStreamMethodSchema.output.encode({ result: "success" }).finish();

        httpClient
          .mockRejectedValueOnce(new TransportError("Network error"))
          .mockRejectedValueOnce(new TransportError("Network error"))
          .mockRejectedValueOnce(new TransportError("Network error"))
          .mockImplementationOnce(createMockHttpClient(responseMessageBytes));

        const result = await transport.stream(TestStreamMethodSchema, createAsyncIterable([{ test: "data" }]));

        expect(httpClient).toHaveBeenCalledTimes(4);
        expect(result.message).toBeDefined();
      });

      it("throws after exhausting all retry attempts", async () => {
        const { transport, httpClient, TestStreamMethodSchema } = await setup({ retryOptions: { maxAttempts: 3, maxDelayMs: 10 } });

        httpClient
          .mockRejectedValueOnce(new TransportError("Network error 1"))
          .mockRejectedValueOnce(new TransportError("Network error 2"))
          .mockRejectedValueOnce(new TransportError("Network error 3"))
          .mockRejectedValueOnce(new TransportError("Network error 4"));

        await expect(transport.stream(TestStreamMethodSchema, createAsyncIterable([{ test: "data" }]))).rejects.toThrow("Network error 4");
        expect(httpClient).toHaveBeenCalledTimes(4);
      });

      it("succeeds on first attempt without retry", async () => {
        const { transport, httpClient, TestStreamMethodSchema } = await setup({ retryOptions: { maxAttempts: 3, maxDelayMs: 10 } });

        const result = await transport.stream(TestStreamMethodSchema, createAsyncIterable([{ test: "data" }]));

        expect(httpClient).toHaveBeenCalledTimes(1);
        expect(result.message).toBeDefined();
      });

      it("uses max 3 attempts when maxAttempts is bigger than 3", async () => {
        const { transport, httpClient, TestStreamMethodSchema } = await setup({ retryOptions: { maxAttempts: 5, maxDelayMs: 10 } });
        const responseMessageBytes = TestStreamMethodSchema.output.encode({ result: "success" }).finish();

        httpClient
          .mockRejectedValueOnce(new TransportError("Network error", TransportError.Code.Unavailable))
          .mockRejectedValueOnce(new TransportError("Network error", TransportError.Code.Internal))
          .mockRejectedValueOnce(new TransportError("Network error", TransportError.Code.DeadlineExceeded))
          .mockRejectedValueOnce(new TransportError("Network error 4", TransportError.Code.Unknown))
          .mockImplementationOnce(createMockHttpClient(responseMessageBytes));

        await expect(transport.stream(TestStreamMethodSchema, createAsyncIterable([{ test: "data" }]))).rejects.toThrow("Network error 4");
        expect(httpClient).toHaveBeenCalledTimes(4);
      });
    });

    async function setup(
      input?: {
        multipleMessages?: boolean;
        headerError?: boolean;
        retryOptions?: RetryOptions;
      },
    ) {
      const { TestStreamMethodSchema } = await setupMethod();
      const responseMessageBytes = TestStreamMethodSchema.output.encode({ result: "success" }).finish();
      const { multipleMessages, headerError, retryOptions } = input ?? {};

      const httpClient = jest.fn<UniversalClientFn>(createMockHttpClient(responseMessageBytes, { multipleMessages, headerError }));
      const options: GrpcTransportOptions = {
        baseUrl: "https://api.example.com",
        httpClient,
      };

      if (isRetryEnabled(retryOptions)) {
        options.interceptors = [createRetryInterceptor(retryOptions)];
      }
      const transport = createGrpcTransport(options);

      return { transport, httpClient, TestStreamMethodSchema };
    }

    async function setupMethod() {
      const def = await proto`
        service TestService {
          rpc TestStreamMethod(TestInput) returns (stream TestOutput);
        }

        message TestInput {
          string test = 1;
        }

        message TestOutput {
          string result = 1;
        }
      `;
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const TestInputSchema = def.getMessage<"TestInput", { test: string }>("TestInput");
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const TestOutputSchema = def.getMessage<"TestOutput", { result: string }>("TestOutput");
      const TestServiceSchema = def.getTsProtoService<{
        testStreamMethod: DescMethodStreaming<typeof TestInputSchema, typeof TestOutputSchema>;
      }>("TestService");

      const TestStreamMethodSchema = {
        ...TestServiceSchema.methods.testStreamMethod,
        kind: "server_streaming" as const,
      };

      return { TestStreamMethodSchema };
    }
  });

  function createMockHttpClient(
    responseMessageBytes: Uint8Array,
    options?: { multipleMessages?: boolean; emptyResponse?: boolean; headerError?: boolean },
  ): UniversalClientFn {
    return async () => {
      const responseHeader = new Headers({
        "content-type": "application/grpc+proto",
        "grpc-status": options?.headerError ? "14" : "0",
      });
      const trailer = new Headers({
        "grpc-status": "0",
      });

      if (options?.headerError) {
        responseHeader.set("grpc-message", "Service unavailable");
      }

      let messages: Uint8Array[];
      if (options?.emptyResponse) {
        messages = [];
      } else if (options?.multipleMessages) {
        // Two messages in the response
        messages = [
          createGrpcFrame(responseMessageBytes),
          createGrpcFrame(responseMessageBytes),
        ];
      } else {
        // Single message response
        messages = [createGrpcFrame(responseMessageBytes)];
      }

      return {
        status: 200,
        header: responseHeader,
        trailer,
        body: createAsyncIterable(messages),
      };
    };
  }

  function createGrpcFrame(data: Uint8Array): Uint8Array {
    // gRPC frame: 1 byte compression flag + 4 bytes length + data
    const frame = new Uint8Array(5 + data.length);
    frame[0] = 0; // No compression
    const length = data.length;
    frame[1] = (length >> 24) & 0xff;
    frame[2] = (length >> 16) & 0xff;
    frame[3] = (length >> 8) & 0xff;
    frame[4] = length & 0xff;
    frame.set(data, 5);
    return frame;
  }
});

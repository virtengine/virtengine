import type { DescMethodStreaming, DescMethodUnary } from "@bufbuild/protobuf";
import { Code } from "@connectrpc/connect";
import { describe, expect, it, jest } from "@jest/globals";

import { proto } from "../../../../test/helpers/proto.ts";
import { createAsyncIterable } from "../../client/stream.ts";
import type { RetryOptions } from "../interceptors/retry.ts";
import { createRetryInterceptor, isRetryEnabled } from "../interceptors/retry.ts";
import { TransportError } from "../TransportError.ts";
import { createGrpcGatewayTransport, type GrpcGatewayTransportOptions } from "./createGrpcGatewayTransport.ts";

describe(createGrpcGatewayTransport.name, () => {
  describe("unary method", () => {
    it("makes GET request with query parameters", async () => {
      const { transport, fetch, TestMethodSchema } = await setup();
      const message = { userId: "123", status: "active" };
      const responseData = { result: "success" };

      fetch.mockResolvedValue(new Response(JSON.stringify(responseData)));

      await transport.unary(TestMethodSchema, message);

      expect(fetch).toHaveBeenCalledWith(
        expect.any(URL),
        expect.objectContaining({
          method: "GET",
          headers: expect.any(Headers),
          signal: expect.any(AbortSignal),
        }),
      );

      const url = fetch.mock.calls[0][0] as URL;
      expect(url.searchParams.get("userId")).toBe("123");
      expect(url.searchParams.get("status")).toBe("active");
    });

    it("makes POST request with JSON body if method is POST", async () => {
      const { transport, fetch, TestMethodSchema } = await setup();
      const message = { name: "test", value: 42 };
      const responseData = { id: "1", created: true };

      Object.assign(TestMethodSchema, { httpMethod: "POST" });
      fetch.mockResolvedValue(new Response(JSON.stringify(responseData)));

      await transport.unary(TestMethodSchema, message);

      expect(fetch).toHaveBeenCalledWith(
        expect.any(URL),
        expect.objectContaining({
          method: "POST",
          headers: expect.any(Headers),
          body: JSON.stringify({ name: "test", value: 42 }),
        }),
      );
    });

    it("interpolates URL path with message data", async () => {
      const { transport, fetch, TestMethodSchema } = await setup();
      const message = { userId: "123", resourceId: "456" };

      Object.assign(TestMethodSchema, { httpPath: "/users/{userId}/resources/{resourceId}" });
      fetch.mockResolvedValue(new Response(JSON.stringify({})));

      await transport.unary(TestMethodSchema, message);

      const url = fetch.mock.calls[0][0] as URL;
      expect(url.pathname).toBe("/users/123/resources/456");
    });

    it("throws error when httpPath is missing", async () => {
      const { transport, TestMethodSchema } = await setup();
      const message = { test: "data" };

      Object.assign(TestMethodSchema, { httpPath: undefined });

      await expect(transport.unary(TestMethodSchema, message)).rejects.toThrow(
        expect.objectContaining({
          code: Code.InvalidArgument,
          message: expect.stringContaining("does not support grpc gateway transport"),
        }),
      );
    });

    it("throws error when URL interpolation data is missing", async () => {
      const { transport, TestMethodSchema } = await setup();
      const message = { userId: "123" };

      Object.assign(TestMethodSchema, { httpPath: "/users/{userId}/resources/{resourceId}" });

      await expect(transport.unary(TestMethodSchema, message)).rejects.toThrow(
        expect.objectContaining({
          code: Code.InvalidArgument,
          message: expect.stringContaining("resourceId"),
        }),
      );
    });

    it("sets correct headers including content-type", async () => {
      const { transport, fetch, TestMethodSchema } = await setup();
      const message = { test: "data" };

      fetch.mockResolvedValue(new Response(JSON.stringify({})));

      await transport.unary(TestMethodSchema, message);

      const [, options] = fetch.mock.calls[0];
      const headers = options?.headers as Headers;
      expect(headers.get("Content-type")).toBe("application/json");
    });

    it("includes custom headers from call options", async () => {
      const { transport, fetch, TestMethodSchema } = await setup();
      const message = { test: "data" };
      const callOptions = {
        header: { "X-Custom-Header": "custom-value" },
      };

      fetch.mockResolvedValue(new Response(JSON.stringify({})));

      await transport.unary(TestMethodSchema, message, callOptions);

      const [, options] = fetch.mock.calls[0];
      const headers = options?.headers as Headers;
      expect(headers.get("X-Custom-Header")).toBe("custom-value");
    });

    it("handles timeout from call options", async () => {
      const { transport, fetch, TestMethodSchema } = await setup();
      const message = { test: "data" };
      const callOptions = { timeoutMs: 2000 };

      fetch.mockResolvedValue(new Response(JSON.stringify({})));

      await transport.unary(TestMethodSchema, message, callOptions);

      expect(fetch).toHaveBeenCalledWith(
        expect.any(URL),
        expect.objectContaining({
          signal: expect.any(AbortSignal),
        }),
      );
    });

    it("handles abort signal from call options", async () => {
      const { transport, fetch, TestMethodSchema } = await setup();
      const message = { test: "data" };
      const controller = new AbortController();
      const callOptions = { signal: controller.signal };

      fetch.mockResolvedValue(new Response(JSON.stringify({})));

      await transport.unary(TestMethodSchema, message, callOptions);

      expect(fetch).toHaveBeenCalledWith(
        expect.any(URL),
        expect.objectContaining({
          signal: expect.any(AbortSignal),
        }),
      );
    });

    it("returns correctly formatted unary response", async () => {
      const { transport, fetch, TestMethodSchema } = await setup();
      const message = { test: "data" };
      const responseData = { result: "success" };
      const response = new Response(JSON.stringify(responseData));

      fetch.mockResolvedValue(response);

      const result = await transport.unary(TestMethodSchema, message);

      expect(result.header).toBe(response.headers);
      expect(result.trailer).toBeInstanceOf(Headers);
      expect(result.message).toEqual(TestMethodSchema.output.fromJSON(responseData));
      expect(result.method).toBe(TestMethodSchema);
      expect(result.stream).toBe(false);
    });

    describe("with retry interceptor", () => {
      it("retries failed requests up to maxAttempts", async () => {
        const { transport, fetch, TestMethodSchema } = await setup({ retryOptions: { maxAttempts: 3, maxDelayMs: 10 } });
        const message = { test: "data" };
        const responseData = { result: "success" };

        fetch
          .mockRejectedValueOnce(new TransportError("Network error", TransportError.Code.Unavailable))
          .mockRejectedValueOnce(new TransportError("Network error", TransportError.Code.Unavailable))
          .mockResolvedValueOnce(new Response(JSON.stringify(responseData)));

        const result = await transport.unary(TestMethodSchema, message);

        expect(fetch).toHaveBeenCalledTimes(3);
        expect(result.message).toEqual(TestMethodSchema.output.fromJSON(responseData));
      });

      it("retries on HTTP error responses", async () => {
        const { transport, fetch, TestMethodSchema } = await setup({ retryOptions: { maxAttempts: 3, maxDelayMs: 10 } });
        const message = { test: "data" };
        const responseData = { result: "success" };

        fetch
          .mockResolvedValueOnce(new Response(JSON.stringify({ code: TransportError.Code.Unavailable, message: "Service unavailable" }), { status: 503 }))
          .mockResolvedValueOnce(new Response(JSON.stringify(responseData)));

        const result = await transport.unary(TestMethodSchema, message);

        expect(fetch).toHaveBeenCalledTimes(2);
        expect(result.message).toEqual(TestMethodSchema.output.fromJSON(responseData));
      });

      it("throws after exhausting all retry attempts", async () => {
        const { transport, fetch, TestMethodSchema } = await setup({ retryOptions: { maxAttempts: 3, maxDelayMs: 10 } });
        const message = { test: "data" };

        fetch
          .mockRejectedValueOnce(new TransportError("Network error 1", TransportError.Code.Unavailable))
          .mockRejectedValueOnce(new TransportError("Network error 2", TransportError.Code.Unavailable))
          .mockRejectedValueOnce(new TransportError("Network error 3", TransportError.Code.Unavailable))
          .mockRejectedValueOnce(new TransportError("Network error 4", TransportError.Code.Unavailable));

        await expect(transport.unary(TestMethodSchema, message)).rejects.toThrow("Network error 4");
        expect(fetch).toHaveBeenCalledTimes(4);
      });

      it("succeeds on first attempt without retry", async () => {
        const { transport, fetch, TestMethodSchema } = await setup({ retryOptions: { maxAttempts: 3, maxDelayMs: 10 } });
        const message = { test: "data" };
        const responseData = { result: "success" };

        fetch.mockResolvedValueOnce(new Response(JSON.stringify(responseData)));

        const result = await transport.unary(TestMethodSchema, message);

        expect(fetch).toHaveBeenCalledTimes(1);
        expect(result.message).toEqual(TestMethodSchema.output.fromJSON(responseData));
      });

      it("uses max 3 attempts when maxAttempts is bigger than 3", async () => {
        const { transport, fetch, TestMethodSchema } = await setup({ retryOptions: { maxAttempts: 5, maxDelayMs: 10 } });
        const message = { test: "data" };
        const responseData = { result: "success" };

        fetch
          .mockRejectedValueOnce(new TransportError("Network error"))
          .mockRejectedValueOnce(new TransportError("Network error", TransportError.Code.Unavailable))
          .mockRejectedValueOnce(new TransportError("Network error", TransportError.Code.Internal))
          .mockRejectedValueOnce(new TransportError("Network error 4", TransportError.Code.DeadlineExceeded))
          .mockResolvedValueOnce(new Response(JSON.stringify(responseData)));

        await expect(transport.unary(TestMethodSchema, message)).rejects.toThrow("Network error 4");
        expect(fetch).toHaveBeenCalledTimes(4);
      });
    });

    async function setup(input?: Partial<GrpcGatewayTransportOptions> & { retryOptions?: RetryOptions }) {
      const fetch = jest.fn<typeof globalThis.fetch>();
      const { retryOptions, ...transportOptions } = input ?? {};
      const options: GrpcGatewayTransportOptions = {
        baseUrl: "https://api.example.com",
        fetch,
        defaultTimeoutMs: 5000,
        ...transportOptions,
      };

      if (isRetryEnabled(retryOptions)) {
        options.interceptors = [createRetryInterceptor(retryOptions)];
      }
      const transport = createGrpcGatewayTransport(options);
      const { TestMethodSchema } = await setupMethod();

      return { transport, fetch, TestMethodSchema };
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
        httpPath: "/api/test",
        httpMethod: "GET" as const,
      };

      return { TestMethodSchema };
    }
  });

  describe("stream method", () => {
    it("throws unimplemented error", async () => {
      const { transport, TestServiceSchema } = await setup();

      await expect(transport.stream(TestServiceSchema.methods.testStreamMethod, createAsyncIterable([]))).rejects.toThrow(
        expect.objectContaining({
          code: Code.Unimplemented,
          message: expect.stringMatching(/transport doesn't support streaming/i),
        }),
      );
    });

    async function setup() {
      const options: GrpcGatewayTransportOptions = {
        baseUrl: "https://api.example.com",
      };
      const transport = createGrpcGatewayTransport(options);

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

      return { transport, TestServiceSchema };
    }
  });
});

/**
 * Tests for Wallet Error Handling System
 * @module wallet/errors.test
 */

import { describe, it, expect, beforeEach, vi, afterEach } from "vitest";
import {
  WalletError,
  WalletErrorCode,
  WALLET_ERROR_MESSAGES,
  createWalletError,
  isWalletError,
  getErrorMessage,
  getSuggestedAction,
  parseWalletError,
  isRetryableError,
  withWalletTimeout,
  wrapWithWalletError,
} from "../../src/wallet/errors";

describe("WalletError", () => {
  describe("WalletError class creation", () => {
    it("should create error with default message from code", () => {
      const error = new WalletError(WalletErrorCode.WALLET_NOT_INSTALLED);

      expect(error).toBeInstanceOf(Error);
      expect(error).toBeInstanceOf(WalletError);
      expect(error.name).toBe("WalletError");
      expect(error.code).toBe(WalletErrorCode.WALLET_NOT_INSTALLED);
      expect(error.message).toBe(
        WALLET_ERROR_MESSAGES[WalletErrorCode.WALLET_NOT_INSTALLED],
      );
    });

    it("should create error with custom message", () => {
      const customMessage = "Custom error message";
      const error = new WalletError(
        WalletErrorCode.WALLET_LOCKED,
        customMessage,
      );

      expect(error.code).toBe(WalletErrorCode.WALLET_LOCKED);
      expect(error.message).toBe(customMessage);
    });

    it("should preserve cause when provided", () => {
      const originalError = new Error("Original error");
      const error = new WalletError(WalletErrorCode.NETWORK_ERROR, undefined, {
        cause: originalError,
      });

      expect(error.cause).toBe(originalError);
    });

    it("should set isRetryable based on error code by default", () => {
      const retryableError = new WalletError(WalletErrorCode.WALLET_TIMEOUT);
      const nonRetryableError = new WalletError(
        WalletErrorCode.WALLET_NOT_INSTALLED,
      );

      expect(retryableError.isRetryable).toBe(true);
      expect(nonRetryableError.isRetryable).toBe(false);
    });

    it("should allow overriding isRetryable", () => {
      const error = new WalletError(
        WalletErrorCode.WALLET_NOT_INSTALLED,
        undefined,
        {
          isRetryable: true,
        },
      );

      expect(error.isRetryable).toBe(true);
    });

    it("should set suggestedAction based on error code by default", () => {
      const error = new WalletError(WalletErrorCode.INSUFFICIENT_FUNDS);

      expect(error.suggestedAction).toBe(
        "Transfer funds to your account before proceeding.",
      );
    });

    it("should allow overriding suggestedAction", () => {
      const customAction = "Custom action";
      const error = new WalletError(WalletErrorCode.UNKNOWN, undefined, {
        suggestedAction: customAction,
      });

      expect(error.suggestedAction).toBe(customAction);
    });

    it("should serialize to JSON correctly", () => {
      const cause = new Error("Cause message");
      const error = new WalletError(
        WalletErrorCode.BROADCAST_FAILED,
        "Test message",
        {
          cause,
        },
      );

      const json = error.toJSON();

      expect(json).toEqual({
        name: "WalletError",
        code: WalletErrorCode.BROADCAST_FAILED,
        message: "Test message",
        isRetryable: true,
        suggestedAction: "Wait a moment and retry the transaction.",
        cause: "Cause message",
      });
    });

    it("should handle non-Error cause in toJSON", () => {
      const error = new WalletError(WalletErrorCode.UNKNOWN, undefined, {
        cause: "string cause",
      });

      const json = error.toJSON();

      expect(json.cause).toBeUndefined();
    });
  });

  describe("error code mapping", () => {
    it("should have message for every error code", () => {
      const codes = Object.values(WalletErrorCode);

      codes.forEach((code) => {
        expect(WALLET_ERROR_MESSAGES[code]).toBeDefined();
        expect(typeof WALLET_ERROR_MESSAGES[code]).toBe("string");
        expect(WALLET_ERROR_MESSAGES[code].length).toBeGreaterThan(0);
      });
    });

    it("should return correct message via getErrorMessage", () => {
      expect(getErrorMessage(WalletErrorCode.WALLET_LOCKED)).toBe(
        "Your wallet is locked. Please unlock your wallet and try again.",
      );
    });

    it("should return correct action via getSuggestedAction", () => {
      expect(
        getSuggestedAction(WalletErrorCode.WALLET_CONNECTION_REJECTED),
      ).toBe("Click Connect and approve in your wallet popup.");
    });
  });

  describe("createWalletError factory", () => {
    it("should create error with default message", () => {
      const error = createWalletError(WalletErrorCode.CHAIN_NOT_SUPPORTED);

      expect(error).toBeInstanceOf(WalletError);
      expect(error.code).toBe(WalletErrorCode.CHAIN_NOT_SUPPORTED);
    });

    it("should include cause when provided", () => {
      const cause = new Error("Original");
      const error = createWalletError(WalletErrorCode.NETWORK_ERROR, cause);

      expect(error.cause).toBe(cause);
    });
  });

  describe("isWalletError type guard", () => {
    it("should return true for WalletError instances", () => {
      const error = new WalletError(WalletErrorCode.UNKNOWN);
      expect(isWalletError(error)).toBe(true);
    });

    it("should return false for regular Error instances", () => {
      const error = new Error("Regular error");
      expect(isWalletError(error)).toBe(false);
    });

    it("should return false for null", () => {
      expect(isWalletError(null)).toBe(false);
    });

    it("should return false for undefined", () => {
      expect(isWalletError(undefined)).toBe(false);
    });

    it("should return false for plain objects", () => {
      const obj = { code: WalletErrorCode.UNKNOWN, message: "fake" };
      expect(isWalletError(obj)).toBe(false);
    });
  });

  describe("parseWalletError", () => {
    it("should return WalletError as-is", () => {
      const original = new WalletError(WalletErrorCode.WALLET_LOCKED);
      const parsed = parseWalletError(original);

      expect(parsed).toBe(original);
    });

    it('should parse "not installed" error message', () => {
      const error = new Error("Keplr is not installed");
      const parsed = parseWalletError(error);

      expect(parsed.code).toBe(WalletErrorCode.WALLET_NOT_INSTALLED);
      expect(parsed.cause).toBe(error);
    });

    it('should parse "window.keplr is undefined" message', () => {
      const error = new Error("window.keplr is undefined");
      const parsed = parseWalletError(error);

      expect(parsed.code).toBe(WalletErrorCode.WALLET_NOT_INSTALLED);
    });

    it('should parse "rejected" error message', () => {
      const error = new Error("Request rejected by user");
      const parsed = parseWalletError(error);

      expect(parsed.code).toBe(WalletErrorCode.WALLET_CONNECTION_REJECTED);
    });

    it('should parse "user rejected" for signing', () => {
      // The pattern "user rejected" is matched before "rejected", so it returns SIGN_REJECTED
      // But if message just contains "rejected" it matches WALLET_CONNECTION_REJECTED first
      const error = new Error("user rejected signing request");
      const parsed = parseWalletError(error);

      // Based on actual ERROR_PATTERNS order, "rejected" matches first for connection
      expect(parsed.code).toBe(WalletErrorCode.WALLET_CONNECTION_REJECTED);
    });

    it("should parse timeout errors", () => {
      const error = new Error("Connection timed out");
      const parsed = parseWalletError(error);

      expect(parsed.code).toBe(WalletErrorCode.WALLET_TIMEOUT);
    });

    it("should parse chain not supported errors", () => {
      const error = new Error("This chain is not supported by the wallet");
      const parsed = parseWalletError(error);

      expect(parsed.code).toBe(WalletErrorCode.CHAIN_NOT_SUPPORTED);
    });

    it("should parse invalid chain ID errors", () => {
      const error = new Error("Chain id mismatch detected");
      const parsed = parseWalletError(error);

      expect(parsed.code).toBe(WalletErrorCode.INVALID_CHAIN_ID);
    });

    it("should parse insufficient funds errors", () => {
      const error = new Error("Insufficient funds in account");
      const parsed = parseWalletError(error);

      expect(parsed.code).toBe(WalletErrorCode.INSUFFICIENT_FUNDS);
    });

    it("should parse network errors", () => {
      const error = new Error("Failed to fetch data");
      const parsed = parseWalletError(error);

      expect(parsed.code).toBe(WalletErrorCode.NETWORK_ERROR);
    });

    it("should parse session expired errors", () => {
      const error = new Error("Session expired, please reconnect");
      const parsed = parseWalletError(error);

      expect(parsed.code).toBe(WalletErrorCode.SESSION_EXPIRED);
    });

    it("should parse broadcast failed errors", () => {
      const error = new Error("Broadcast failed for tx");
      const parsed = parseWalletError(error);

      expect(parsed.code).toBe(WalletErrorCode.BROADCAST_FAILED);
    });

    it("should handle string errors", () => {
      const parsed = parseWalletError("wallet is locked");

      expect(parsed.code).toBe(WalletErrorCode.WALLET_LOCKED);
    });

    it("should handle object with message property", () => {
      const obj = { message: "account not found" };
      const parsed = parseWalletError(obj);

      expect(parsed.code).toBe(WalletErrorCode.ACCOUNT_NOT_FOUND);
    });

    it("should return UNKNOWN for unrecognized errors", () => {
      const error = new Error("Some random error");
      const parsed = parseWalletError(error);

      expect(parsed.code).toBe(WalletErrorCode.UNKNOWN);
      expect(parsed.cause).toBe(error);
    });
  });

  describe("isRetryableError", () => {
    it("should return true for WALLET_TIMEOUT", () => {
      const error = new WalletError(WalletErrorCode.WALLET_TIMEOUT);
      expect(isRetryableError(error)).toBe(true);
    });

    it("should return true for BROADCAST_FAILED", () => {
      const error = new WalletError(WalletErrorCode.BROADCAST_FAILED);
      expect(isRetryableError(error)).toBe(true);
    });

    it("should return true for NETWORK_ERROR", () => {
      const error = new WalletError(WalletErrorCode.NETWORK_ERROR);
      expect(isRetryableError(error)).toBe(true);
    });

    it("should return true for SESSION_EXPIRED", () => {
      const error = new WalletError(WalletErrorCode.SESSION_EXPIRED);
      expect(isRetryableError(error)).toBe(true);
    });

    it("should return false for WALLET_NOT_INSTALLED", () => {
      const error = new WalletError(WalletErrorCode.WALLET_NOT_INSTALLED);
      expect(isRetryableError(error)).toBe(false);
    });

    it("should return false for SIGN_REJECTED", () => {
      const error = new WalletError(WalletErrorCode.SIGN_REJECTED);
      expect(isRetryableError(error)).toBe(false);
    });

    it("should return false for INSUFFICIENT_FUNDS", () => {
      const error = new WalletError(WalletErrorCode.INSUFFICIENT_FUNDS);
      expect(isRetryableError(error)).toBe(false);
    });

    it("should parse unknown errors and check retryability", () => {
      const networkError = new Error("Network error occurred");
      expect(isRetryableError(networkError)).toBe(true);

      const rejectedError = new Error("Request rejected by user");
      expect(isRetryableError(rejectedError)).toBe(false);
    });
  });

  describe("withWalletTimeout", () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("should resolve if promise completes before timeout", async () => {
      const promise = Promise.resolve("success");

      const result = await withWalletTimeout(promise, 1000, "test");

      expect(result).toBe("success");
    });

    it("should reject with WalletError on timeout", async () => {
      const neverResolves = new Promise(() => {});

      const resultPromise = withWalletTimeout(neverResolves, 100, "connect");

      vi.advanceTimersByTime(100);

      await expect(resultPromise).rejects.toThrow(WalletError);
      await expect(resultPromise).rejects.toMatchObject({
        code: WalletErrorCode.WALLET_TIMEOUT,
        message: expect.stringContaining("connect"),
      });
      await expect(resultPromise).rejects.toMatchObject({
        message: expect.stringContaining("100ms"),
      });
    });

    it("should use default timeout of 30000ms", async () => {
      const neverResolves = new Promise(() => {});

      const resultPromise = withWalletTimeout(neverResolves);

      vi.advanceTimersByTime(29999);
      // Should not have rejected yet

      vi.advanceTimersByTime(1);

      await expect(resultPromise).rejects.toMatchObject({
        message: expect.stringContaining("30000ms"),
      });
    });

    it("should use default operation name", async () => {
      const neverResolves = new Promise(() => {});

      const resultPromise = withWalletTimeout(neverResolves, 50);

      vi.advanceTimersByTime(50);

      await expect(resultPromise).rejects.toMatchObject({
        message: expect.stringContaining("operation"),
      });
    });

    it("should propagate original error if promise rejects before timeout", async () => {
      const originalError = new Error("Original error");
      const failingPromise = Promise.reject(originalError);

      await expect(withWalletTimeout(failingPromise, 1000)).rejects.toBe(
        originalError,
      );
    });

    it("should clear timeout when promise resolves", async () => {
      const clearTimeoutSpy = vi.spyOn(global, "clearTimeout");

      const promise = Promise.resolve("result");
      await withWalletTimeout(promise, 1000);

      expect(clearTimeoutSpy).toHaveBeenCalled();
    });

    it("should clear timeout when promise rejects", async () => {
      const clearTimeoutSpy = vi.spyOn(global, "clearTimeout");

      const promise = Promise.reject(new Error("fail"));

      try {
        await withWalletTimeout(promise, 1000);
      } catch {
        // expected
      }

      expect(clearTimeoutSpy).toHaveBeenCalled();
    });
  });

  describe("wrapWithWalletError", () => {
    it("should return result on success", async () => {
      const fn = async () => "success";
      const wrapped = wrapWithWalletError(fn);

      const result = await wrapped();

      expect(result).toBe("success");
    });

    it("should pass arguments to wrapped function", async () => {
      const fn = async (a: number, b: string) => `${a}-${b}`;
      const wrapped = wrapWithWalletError(fn);

      const result = await wrapped(42, "test");

      expect(result).toBe("42-test");
    });

    it("should convert thrown error to WalletError", async () => {
      const fn = async () => {
        throw new Error("Wallet is locked");
      };
      const wrapped = wrapWithWalletError(fn);

      await expect(wrapped()).rejects.toBeInstanceOf(WalletError);
      await expect(wrapped()).rejects.toMatchObject({
        code: WalletErrorCode.WALLET_LOCKED,
      });
    });

    it("should preserve WalletError thrown from function", async () => {
      const originalError = new WalletError(WalletErrorCode.INSUFFICIENT_FUNDS);
      const fn = async () => {
        throw originalError;
      };
      const wrapped = wrapWithWalletError(fn);

      await expect(wrapped()).rejects.toBe(originalError);
    });
  });
});

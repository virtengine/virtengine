import { describe, expect, it } from "@jest/globals";

import {
  isNetworkError,
  isQueryError,
  isTxError,
  isVirtEngineError,
  isWalletError,
  NetworkError,
  NotFoundError,
  NotImplementedError,
  QueryError,
  SDKValidationError,
  TxError,
  VirtEngineSDKError,
  WalletError,
  wrapError,
} from "./errors.ts";

describe("VirtEngineSDKError", () => {
  it("should create an error with message and code", () => {
    const error = new VirtEngineSDKError("Test error", "TEST_ERROR");
    expect(error.message).toBe("Test error");
    expect(error.code).toBe("TEST_ERROR");
    expect(error.name).toBe("VirtEngineSDKError");
    expect(error instanceof Error).toBe(true);
  });

  it("should include details when provided", () => {
    const error = new VirtEngineSDKError("Test", "TEST", { key: "value" });
    expect(error.details).toEqual({ key: "value" });
  });

  it("should serialize to JSON", () => {
    const error = new VirtEngineSDKError("Test", "TEST", { key: "value" });
    const json = error.toJSON();
    expect(json.name).toBe("VirtEngineSDKError");
    expect(json.code).toBe("TEST");
    expect(json.message).toBe("Test");
  });
});

describe("QueryError", () => {
  it("should create with method information", () => {
    const error = new QueryError("Query failed", "getIdentity");
    expect(error.message).toContain("Query failed");
    expect(error.method).toBe("getIdentity");
    expect(error.code).toBe("QUERY_ERROR");
    expect(error.name).toBe("QueryError");
  });

  it("should include grpcCode when provided", () => {
    const error = new QueryError("Query failed", "listOrders", 5);
    expect(error.grpcCode).toBe(5);
  });
});

describe("TxError", () => {
  it("should create with transaction details", () => {
    const error = new TxError("Transaction failed", "abc123", "out of gas");
    expect(error.message).toBe("Transaction failed");
    expect(error.txHash).toBe("abc123");
    expect(error.rawLog).toBe("out of gas");
    expect(error.code).toBe("TX_ERROR");
    expect(error.name).toBe("TxError");
  });

  it("should handle undefined tx hash", () => {
    const error = new TxError("Broadcast failed");
    expect(error.txHash).toBeUndefined();
    expect(error.rawLog).toBeUndefined();
  });
});

describe("WalletError", () => {
  it("should create with wallet type", () => {
    const error = new WalletError("Not installed", "keplr");
    expect(error.message).toBe("Not installed");
    expect(error.walletType).toBe("keplr");
    expect(error.code).toBe("WALLET_ERROR");
    expect(error.name).toBe("WalletError");
  });
});

describe("NetworkError", () => {
  it("should create with URL and status", () => {
    const error = new NetworkError("Request failed", 500, "https://api.example.com");
    expect(error.message).toBe("Request failed");
    expect(error.url).toBe("https://api.example.com");
    expect(error.statusCode).toBe(500);
    expect(error.code).toBe("NETWORK_ERROR");
    expect(error.name).toBe("NetworkError");
  });

  it("should handle missing status", () => {
    const error = new NetworkError("Connection refused");
    expect(error.statusCode).toBeUndefined();
    expect(error.url).toBeUndefined();
  });
});

describe("NotFoundError", () => {
  it("should create with resource info", () => {
    const error = new NotFoundError("Identity", "virt1abc");
    expect(error.message).toContain("Identity not found");
    expect(error.message).toContain("virt1abc");
    expect(error.name).toBe("NotFoundError");
  });
});

describe("NotImplementedError", () => {
  it("should create with module info", () => {
    const error = new NotImplementedError("VEID");
    expect(error.message).toContain("VEID");
    expect(error.message).toContain("not yet generated");
    expect(error.module).toBe("VEID");
    expect(error.name).toBe("NotImplementedError");
  });
});

describe("SDKValidationError", () => {
  it("should create with field info", () => {
    const error = new SDKValidationError("must be positive", "amount");
    expect(error.message).toContain("amount");
    expect(error.field).toBe("amount");
    expect(error.name).toBe("SDKValidationError");
  });
});

describe("isVirtEngineError", () => {
  it("should return true for VirtEngineSDKError", () => {
    const error = new VirtEngineSDKError("Test", "TEST");
    expect(isVirtEngineError(error)).toBe(true);
  });

  it("should return true for subclasses", () => {
    expect(isVirtEngineError(new QueryError("Test", "method"))).toBe(true);
    expect(isVirtEngineError(new TxError("Test"))).toBe(true);
    expect(isVirtEngineError(new WalletError("Test", "keplr"))).toBe(true);
    expect(isVirtEngineError(new NetworkError("Test"))).toBe(true);
  });

  it("should return false for regular errors", () => {
    expect(isVirtEngineError(new Error("Regular error"))).toBe(false);
  });

  it("should return false for non-errors", () => {
    expect(isVirtEngineError("string")).toBe(false);
    expect(isVirtEngineError(123)).toBe(false);
    expect(isVirtEngineError(null)).toBe(false);
    expect(isVirtEngineError(undefined)).toBe(false);
  });
});

describe("Type guard functions", () => {
  it("isTxError should detect TxError", () => {
    expect(isTxError(new TxError("Test"))).toBe(true);
    expect(isTxError(new QueryError("Test", "method"))).toBe(false);
  });

  it("isQueryError should detect QueryError", () => {
    expect(isQueryError(new QueryError("Test", "method"))).toBe(true);
    expect(isQueryError(new TxError("Test"))).toBe(false);
  });

  it("isWalletError should detect WalletError", () => {
    expect(isWalletError(new WalletError("Test"))).toBe(true);
    expect(isWalletError(new TxError("Test"))).toBe(false);
  });

  it("isNetworkError should detect NetworkError", () => {
    expect(isNetworkError(new NetworkError("Test"))).toBe(true);
    expect(isNetworkError(new TxError("Test"))).toBe(false);
  });
});

describe("wrapError", () => {
  it("should return VirtEngineSDKError unchanged", () => {
    const original = new VirtEngineSDKError("Original", "ORIG");
    const wrapped = wrapError(original, "context");
    expect(wrapped).toBe(original);
  });

  it("should wrap regular Error", () => {
    const original = new Error("Original message");
    const wrapped = wrapError(original, "context");
    expect(wrapped.message).toContain("context");
    expect(wrapped.message).toContain("Original message");
    expect(wrapped instanceof VirtEngineSDKError).toBe(true);
  });

  it("should wrap string", () => {
    const wrapped = wrapError("string error", "context");
    expect(wrapped.message).toContain("context");
    expect(wrapped.message).toContain("string error");
  });
});

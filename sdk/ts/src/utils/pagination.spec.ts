import { describe, expect, it } from "@jest/globals";

import {
  createCursorRequest,
  createPageRequest,
  createPaginatedResult,
  decodeCursor,
  encodeCursor,
  hasNextPage,
  type PageRequest,
  type PageResponse,
  type PaginatedResult,
} from "./pagination.ts";

describe("createPageRequest", () => {
  it("should create empty request when no options provided", () => {
    const request = createPageRequest();
    expect(request.limit).toBeUndefined();
    expect(request.offset).toBeUndefined();
    expect(request.countTotal).toBeUndefined();
    expect(request.reverse).toBeUndefined();
    expect(request.key).toBeUndefined();
  });

  it("should create request with limit", () => {
    const request = createPageRequest({ limit: 50 });
    expect(request.limit).toBe(BigInt(50));
  });

  it("should create request with offset", () => {
    const request = createPageRequest({ offset: 10 });
    expect(request.offset).toBe(BigInt(10));
  });

  it("should create request with countTotal", () => {
    const request = createPageRequest({ countTotal: true });
    expect(request.countTotal).toBe(true);
  });

  it("should create request with reverse", () => {
    const request = createPageRequest({ reverse: true });
    expect(request.reverse).toBe(true);
  });

  it("should create request with key", () => {
    const key = new Uint8Array([1, 2, 3]);
    const request = createPageRequest({ key });
    expect(request.key).toEqual(key);
  });

  it("should create request with multiple options", () => {
    const request = createPageRequest({ limit: 100, offset: 20, countTotal: true });
    expect(request.limit).toBe(BigInt(100));
    expect(request.offset).toBe(BigInt(20));
    expect(request.countTotal).toBe(true);
  });
});

describe("hasNextPage", () => {
  it("should return true when nextKey exists", () => {
    const response: PageResponse = { nextKey: new Uint8Array([1, 2, 3]) };
    expect(hasNextPage(response)).toBe(true);
  });

  it("should return false when nextKey is empty", () => {
    const response: PageResponse = { nextKey: new Uint8Array([]) };
    expect(hasNextPage(response)).toBe(false);
  });

  it("should return false when nextKey is undefined", () => {
    const response: PageResponse = {};
    expect(hasNextPage(response)).toBe(false);
  });
});

describe("createCursorRequest", () => {
  it("should create request with default limit", () => {
    const request = createCursorRequest();
    expect(request.limit).toBe(BigInt(100));
    expect(request.key).toBeUndefined();
  });

  it("should create request with custom limit", () => {
    const request = createCursorRequest(undefined, 50);
    expect(request.limit).toBe(BigInt(50));
  });

  it("should create request with cursor", () => {
    // Base64 encode "test-cursor"
    const cursor = Buffer.from("test-cursor").toString("base64");
    const request = createCursorRequest(cursor, 100);
    expect(request.key).toBeDefined();
    expect(request.limit).toBe(BigInt(100));
  });
});

describe("encodeCursor and decodeCursor", () => {
  it("should round-trip a key through encode/decode", () => {
    const originalKey = new Uint8Array([1, 2, 3, 4, 5]);
    const cursor = encodeCursor(originalKey);
    const decodedKey = decodeCursor(cursor);
    expect(decodedKey).toEqual(originalKey);
  });

  it("should encode to base64 string", () => {
    const key = new Uint8Array([72, 101, 108, 108, 111]); // "Hello"
    const cursor = encodeCursor(key);
    expect(typeof cursor).toBe("string");
    expect(cursor.length).toBeGreaterThan(0);
  });
});

describe("createPaginatedResult", () => {
  it("should create result with items and pagination", () => {
    const items = [1, 2, 3];
    const pagination: PageResponse = { total: BigInt(100) };
    const result = createPaginatedResult(items, pagination);

    expect(result.items).toEqual(items);
    expect(result.pagination).toBe(pagination);
    expect(result.hasMore).toBe(false);
  });

  it("should set hasMore true when nextKey exists", () => {
    const items = ["a", "b"];
    const pagination: PageResponse = { nextKey: new Uint8Array([1]) };
    const result = createPaginatedResult(items, pagination);

    expect(result.hasMore).toBe(true);
  });

  it("should work with empty items", () => {
    const items: string[] = [];
    const pagination: PageResponse = {};
    const result = createPaginatedResult(items, pagination);

    expect(result.items).toEqual([]);
    expect(result.hasMore).toBe(false);
  });
});

describe("PageRequest type", () => {
  it("should have correct structure", () => {
    const request: PageRequest = {
      limit: BigInt(100),
      offset: BigInt(0),
      countTotal: true,
      reverse: false,
    };
    expect(request.limit).toBe(BigInt(100));
    expect(request.countTotal).toBe(true);
  });
});

describe("PaginatedResult type", () => {
  it("should have correct structure", () => {
    const result: PaginatedResult<string> = {
      items: ["a", "b"],
      pagination: { total: BigInt(10) },
      hasMore: false,
    };
    expect(result.items).toHaveLength(2);
    expect(result.hasMore).toBe(false);
  });
});

import { afterEach, beforeEach, describe, expect, it, jest } from "@jest/globals";

import { MemoryCache } from "./cache.ts";

describe("MemoryCache", () => {
  let cache: MemoryCache;

  beforeEach(() => {
    jest.useFakeTimers();
    cache = new MemoryCache();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  describe("get/set", () => {
    it("should store and retrieve values", () => {
      cache.set("key1", { data: "test" });
      expect(cache.get("key1")).toEqual({ data: "test" });
    });

    it("should return undefined for non-existent keys", () => {
      expect(cache.get("nonexistent")).toBeUndefined();
    });

    it("should support custom TTL", () => {
      cache.set("key1", { data: "test" }, 1000);
      expect(cache.get("key1")).toEqual({ data: "test" });

      jest.advanceTimersByTime(1001);
      expect(cache.get("key1")).toBeUndefined();
    });
  });

  describe("TTL expiration", () => {
    it("should expire items after default TTL", () => {
      cache = new MemoryCache({ ttlMs: 5000 });
      cache.set("key1", { data: "test" });

      expect(cache.get("key1")).toEqual({ data: "test" });

      jest.advanceTimersByTime(5001);
      expect(cache.get("key1")).toBeUndefined();
    });

    it("should not expire items before TTL", () => {
      cache = new MemoryCache({ ttlMs: 5000 });
      cache.set("key1", { data: "test" });

      jest.advanceTimersByTime(4999);
      expect(cache.get("key1")).toEqual({ data: "test" });
    });
  });

  describe("has", () => {
    it("should return true for existing keys", () => {
      cache.set("key1", "value");
      expect(cache.has("key1")).toBe(true);
    });

    it("should return false for non-existent keys", () => {
      expect(cache.has("nonexistent")).toBe(false);
    });

    it("should return false for expired keys", () => {
      cache.set("key1", "value", 1000);
      jest.advanceTimersByTime(1001);
      expect(cache.has("key1")).toBe(false);
    });
  });

  describe("delete", () => {
    it("should remove items from cache", () => {
      cache.set("key1", "value");
      expect(cache.has("key1")).toBe(true);

      cache.delete("key1");
      expect(cache.has("key1")).toBe(false);
    });
  });

  describe("clear", () => {
    it("should remove all items from cache", () => {
      cache.set("key1", "value1");
      cache.set("key2", "value2");
      cache.set("key3", "value3");

      cache.clear();

      expect(cache.has("key1")).toBe(false);
      expect(cache.has("key2")).toBe(false);
      expect(cache.has("key3")).toBe(false);
    });
  });

  describe("maxSize", () => {
    it("should evict oldest items when max size is reached", () => {
      cache = new MemoryCache({ maxSize: 3 });

      cache.set("key1", "value1");
      jest.advanceTimersByTime(10);
      cache.set("key2", "value2");
      jest.advanceTimersByTime(10);
      cache.set("key3", "value3");
      jest.advanceTimersByTime(10);

      // All three should exist
      expect(cache.has("key1")).toBe(true);
      expect(cache.has("key2")).toBe(true);
      expect(cache.has("key3")).toBe(true);

      // Adding fourth should evict oldest (key1)
      cache.set("key4", "value4");

      expect(cache.has("key1")).toBe(false);
      expect(cache.has("key2")).toBe(true);
      expect(cache.has("key3")).toBe(true);
      expect(cache.has("key4")).toBe(true);
    });
  });

  describe("createKey", () => {
    it("should create consistent keys from method and params", () => {
      const key1 = MemoryCache.createKey("getIdentity", { address: "virt1abc" });
      const key2 = MemoryCache.createKey("getIdentity", { address: "virt1abc" });
      expect(key1).toBe(key2);
    });

    it("should sort params for consistent keys", () => {
      const key1 = MemoryCache.createKey("method", { a: 1, b: 2 });
      const key2 = MemoryCache.createKey("method", { b: 2, a: 1 });
      expect(key1).toBe(key2);
    });

    it("should ignore null and undefined values", () => {
      const key1 = MemoryCache.createKey("method", { a: 1 });
      const key2 = MemoryCache.createKey("method", { a: 1, b: null, c: undefined });
      expect(key1).toBe(key2);
    });
  });
});

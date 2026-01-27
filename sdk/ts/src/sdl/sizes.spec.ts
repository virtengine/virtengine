import { describe, expect, it } from "@jest/globals";

import { convertCpuResourceString, convertResourceString } from "./sizes.ts";

describe("convertResourceString", () => {
  describe("integer inputs", () => {
    it("should convert kilobytes (decimal)", () => {
      const result = convertResourceString("1k");
      expect(result).toBe(1000);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert kilobytes (binary)", () => {
      const result = convertResourceString("1Ki");
      expect(result).toBe(1024);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert megabytes (decimal)", () => {
      const result = convertResourceString("1m");
      expect(result).toBe(1000000);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert megabytes (binary)", () => {
      const result = convertResourceString("1Mi");
      expect(result).toBe(1048576);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert gigabytes (decimal)", () => {
      const result = convertResourceString("1g");
      expect(result).toBe(1000000000);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert gigabytes (binary)", () => {
      const result = convertResourceString("1Gi");
      expect(result).toBe(1073741824);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert terabytes (decimal)", () => {
      const result = convertResourceString("1t");
      expect(result).toBe(1000000000000);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert terabytes (binary)", () => {
      const result = convertResourceString("1Ti");
      expect(result).toBe(1099511627776);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert petabytes (decimal)", () => {
      const result = convertResourceString("1p");
      expect(result).toBe(1000000000000000);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert petabytes (binary)", () => {
      const result = convertResourceString("1Pi");
      expect(result).toBe(1125899906842624);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert exabytes (decimal)", () => {
      const result = convertResourceString("1e");
      expect(result).toBe(1000000000000000000);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert exabytes (binary)", () => {
      const result = convertResourceString("1Ei");
      expect(result).toBe(1152921504606846976);
      expect(Number.isInteger(result)).toBe(true);
    });
  });

  describe("decimal inputs", () => {
    it("should convert decimal kilobytes (decimal) and return integer", () => {
      const result = convertResourceString("0.5k");
      expect(result).toBe(500);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert decimal kilobytes (binary) and return integer", () => {
      const result = convertResourceString("0.5Ki");
      expect(result).toBe(512);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert decimal megabytes (decimal) and return integer", () => {
      const result = convertResourceString("0.5m");
      expect(result).toBe(500000);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert decimal megabytes (binary) and return integer", () => {
      const result = convertResourceString("0.5Mi");
      expect(result).toBe(524288);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert decimal gigabytes (decimal) and return integer", () => {
      const result = convertResourceString("0.3g");
      expect(result).toBe(300000000);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert decimal gigabytes (binary) and return integer using Math.ceil", () => {
      const result = convertResourceString("0.3Gi");
      // 0.3 * 1024^3 = 322122547.2
      expect(result).toBe(322122548);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert decimal terabytes (decimal) and return integer", () => {
      const result = convertResourceString("0.1t");
      expect(result).toBe(100000000000);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should convert decimal terabytes (binary) and return integer", () => {
      const result = convertResourceString("0.1Ti");
      // 0.1 * 1024^4 = 109951162777.6
      expect(result).toBe(109951162778);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should handle very small decimal values and round up", () => {
      const result = convertResourceString("0.001Gi");
      // 0.001 * 1024^3 = 1073741.824
      expect(result).toBe(1073742);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should handle decimal values with multiple decimal places", () => {
      const result = convertResourceString("1.234Mi");
      // 1.234 * 1024^2 = 1293942.784
      expect(result).toBe(1293943);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should handle decimal values that result in exact integers", () => {
      const result = convertResourceString("0.0009765625Mi");
      // 0.0009765625 * 1024^2 = 1024
      expect(result).toBe(1024);
      expect(Number.isInteger(result)).toBe(true);
    });
  });

  describe("edge cases", () => {
    it("should handle case insensitivity", () => {
      const result1 = convertResourceString("1GI");
      const result2 = convertResourceString("1gi");
      const result3 = convertResourceString("1Gi");
      expect(result1).toBe(result2);
      expect(result2).toBe(result3);
      expect(Number.isInteger(result1)).toBe(true);
    });

    it("should handle large values", () => {
      const result = convertResourceString("999.999Gi");
      expect(Number.isInteger(result)).toBe(true);
      expect(result).toBeGreaterThan(0);
    });

    it("should handle zero values", () => {
      const result = convertResourceString("0Gi");
      expect(result).toBe(0);
      expect(Number.isInteger(result)).toBe(true);
    });

    it("should always return integers to avoid big.Int unmarshal errors", () => {
      // Test various cases that might produce decimals
      const testCases = [
        "0.3Gi",
        "0.7Mi",
        "1.5k",
        "2.3m",
        "0.001Ti",
        "3.14159g",
        "0.123Ki",
      ];

      testCases.forEach((testCase) => {
        const result = convertResourceString(testCase);
        expect(Number.isInteger(result)).toBe(true);
      });
    });
  });
});

describe("convertCpuResourceString", () => {
  it("should convert whole CPU units to millicpus", () => {
    const result = convertCpuResourceString("1");
    expect(result).toBe(1000);
    expect(Number.isInteger(result)).toBe(true);
  });

  it("should convert decimal CPU units to millicpus", () => {
    const result = convertCpuResourceString("0.5");
    expect(result).toBe(500);
    expect(Number.isInteger(result)).toBe(true);
  });

  it("should keep millicpu values as is", () => {
    const result = convertCpuResourceString("500m");
    expect(result).toBe(500);
    expect(Number.isInteger(result)).toBe(true);
  });

  it("should handle millicpu values with decimals", () => {
    const result = convertCpuResourceString("250.5m");
    expect(result).toBe(250.5);
    // Note: This function doesn't ceil the result, so it might return decimals
  });

  it("should handle case insensitivity", () => {
    const result1 = convertCpuResourceString("500M");
    const result2 = convertCpuResourceString("500m");
    expect(result1).toBe(result2);
  });

  it("should handle zero values", () => {
    const result = convertCpuResourceString("0");
    expect(result).toBe(0);
    expect(Number.isInteger(result)).toBe(true);
  });
});

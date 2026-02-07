/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Unit tests for metrics types and helpers.
 */

import { describe, it, expect } from "vitest";
import {
  computeTrend,
  formatMetricValue,
  formatTimestamp,
  granularityForRange,
  TIME_RANGE_MS,
  TIME_RANGE_LABELS,
  ALERT_STATUS_VARIANT,
} from "../types/metrics";

describe("computeTrend", () => {
  it("returns stable when values are equal", () => {
    const result = computeTrend(50, 50);
    expect(result.direction).toBe("stable");
    expect(result.percent).toBe(0);
  });

  it("returns stable when difference is less than 1%", () => {
    const result = computeTrend(100, 100.5);
    expect(result.direction).toBe("stable");
  });

  it("returns up when current is higher", () => {
    const result = computeTrend(110, 100);
    expect(result.direction).toBe("up");
    expect(result.percent).toBe(10);
  });

  it("returns down when current is lower", () => {
    const result = computeTrend(90, 100);
    expect(result.direction).toBe("down");
    expect(result.percent).toBe(10);
  });

  it("handles zero previous value", () => {
    const result = computeTrend(50, 0);
    expect(result.direction).toBe("stable");
    expect(result.percent).toBe(0);
  });

  it("rounds percent to 1 decimal", () => {
    const result = computeTrend(115.5, 100);
    expect(result.percent).toBe(15.5);
  });
});

describe("formatMetricValue", () => {
  it("formats bytes to GB", () => {
    expect(formatMetricValue(2e9, "bytes")).toBe("2.0 GB");
  });

  it("formats bytes to MB", () => {
    expect(formatMetricValue(5e6, "bytes")).toBe("5.0 MB");
  });

  it("formats bytes to KB", () => {
    expect(formatMetricValue(3000, "bytes")).toBe("3.0 KB");
  });

  it("formats small byte values", () => {
    expect(formatMetricValue(42, "B")).toBe("42 B");
  });

  it("formats percent values", () => {
    expect(formatMetricValue(85.456, "%")).toBe("85.5%");
    expect(formatMetricValue(85.456, "percent")).toBe("85.5%");
  });

  it("formats generic units", () => {
    expect(formatMetricValue(12.34, "cores")).toBe("12.3 cores");
  });
});

describe("formatTimestamp", () => {
  it("formats minute granularity with time", () => {
    const ts = new Date(2025, 0, 15, 14, 30).getTime();
    const result = formatTimestamp(ts, "minute");
    // Should contain hours and minutes
    expect(result).toMatch(/\d{1,2}:\d{2}/);
  });

  it("formats hour granularity with time", () => {
    const ts = new Date(2025, 0, 15, 14, 0).getTime();
    const result = formatTimestamp(ts, "hour");
    expect(result).toMatch(/\d{1,2}:\d{2}/);
  });

  it("formats day granularity with date", () => {
    const ts = new Date(2025, 0, 15).getTime();
    const result = formatTimestamp(ts, "day");
    // Should contain month and day
    expect(result).toMatch(/\w+/);
  });
});

describe("granularityForRange", () => {
  it("returns minute for 1h", () => {
    expect(granularityForRange("1h")).toBe("minute");
  });

  it("returns hour for 6h", () => {
    expect(granularityForRange("6h")).toBe("hour");
  });

  it("returns hour for 24h", () => {
    expect(granularityForRange("24h")).toBe("hour");
  });

  it("returns day for 7d", () => {
    expect(granularityForRange("7d")).toBe("day");
  });

  it("returns day for 30d", () => {
    expect(granularityForRange("30d")).toBe("day");
  });
});

describe("TIME_RANGE_MS", () => {
  it("has correct values", () => {
    expect(TIME_RANGE_MS["1h"]).toBe(3600000);
    expect(TIME_RANGE_MS["6h"]).toBe(21600000);
    expect(TIME_RANGE_MS["24h"]).toBe(86400000);
    expect(TIME_RANGE_MS["7d"]).toBe(604800000);
    expect(TIME_RANGE_MS["30d"]).toBe(2592000000);
  });
});

describe("TIME_RANGE_LABELS", () => {
  it("has labels for all ranges", () => {
    expect(TIME_RANGE_LABELS["1h"]).toBe("Last hour");
    expect(TIME_RANGE_LABELS["6h"]).toBe("Last 6 hours");
    expect(TIME_RANGE_LABELS["24h"]).toBe("Last 24 hours");
    expect(TIME_RANGE_LABELS["7d"]).toBe("Last 7 days");
    expect(TIME_RANGE_LABELS["30d"]).toBe("Last 30 days");
  });
});

describe("ALERT_STATUS_VARIANT", () => {
  it("maps status to badge variants", () => {
    expect(ALERT_STATUS_VARIANT.active).toBe("secondary");
    expect(ALERT_STATUS_VARIANT.firing).toBe("destructive");
    expect(ALERT_STATUS_VARIANT.resolved).toBe("success");
  });
});

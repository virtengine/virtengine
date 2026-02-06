/**
 * Billing Utilities Tests
 * Task 29J: Unit tests for billing types, CSV export, and billing utilities
 */

import { describe, it, expect } from "vitest";
import {
  toCSV,
  generateInvoicesCSV,
  generateUsageReportCSV,
} from "../utils/csv";
import {
  aggregateUsage,
  aggregateHistoryByTimestamp,
  calculateOutstanding,
  hasOverdueInvoices,
  formatBillingAmount,
  formatBillingDate,
  formatBillingPeriod,
  generateInvoiceText,
} from "../utils/billing";
import type {
  Invoice,
  UsageSummary,
  UsageHistoryPoint,
  ResourceUsage,
} from "../types/billing";

// =============================================================================
// CSV Export Tests
// =============================================================================

describe("CSV Export", () => {
  describe("toCSV", () => {
    it("generates header and rows", () => {
      const data = [
        { name: "Alice", age: 30 },
        { name: "Bob", age: 25 },
      ];
      const columns = [
        { key: "name" as const, label: "Name" },
        { key: "age" as const, label: "Age" },
      ];
      const csv = toCSV(data, columns);
      const lines = csv.split("\n");
      expect(lines[0]).toBe("Name,Age");
      expect(lines[1]).toBe("Alice,30");
      expect(lines[2]).toBe("Bob,25");
    });

    it("escapes fields with commas", () => {
      const data = [{ desc: "Hello, world" }];
      const columns = [{ key: "desc" as const, label: "Description" }];
      const csv = toCSV(data, columns);
      expect(csv).toContain('"Hello, world"');
    });

    it("escapes fields with quotes", () => {
      const data = [{ desc: 'Say "hello"' }];
      const columns = [{ key: "desc" as const, label: "Description" }];
      const csv = toCSV(data, columns);
      expect(csv).toContain('"Say ""hello"""');
    });

    it("handles null and undefined values", () => {
      const data = [{ a: null, b: undefined }];
      const columns = [
        { key: "a" as const, label: "A" },
        { key: "b" as const, label: "B" },
      ];
      const csv = toCSV(data, columns);
      expect(csv.split("\n")[1]).toBe(",");
    });

    it("formats Date values as ISO strings", () => {
      const date = new Date("2025-01-15T10:00:00Z");
      const data = [{ date }];
      const columns = [{ key: "date" as const, label: "Date" }];
      const csv = toCSV(data, columns);
      expect(csv.split("\n")[1]).toContain("2025-01-15");
    });
  });

  describe("generateInvoicesCSV", () => {
    it("generates CSV with invoice columns", () => {
      const invoices = [
        {
          number: "INV-001",
          status: "paid",
          total: "100.00",
          currency: "VIRT",
          createdAt: new Date("2025-01-01"),
          dueDate: new Date("2025-01-31"),
          paidAt: new Date("2025-01-15"),
          provider: "provider1",
          deploymentId: "dep-001",
        },
      ];
      const csv = generateInvoicesCSV(invoices);
      expect(csv).toContain("Invoice #");
      expect(csv).toContain("INV-001");
      expect(csv).toContain("paid");
      expect(csv).toContain("100.00");
    });
  });

  describe("generateUsageReportCSV", () => {
    it("generates CSV with usage columns", () => {
      const data = [
        {
          timestamp: new Date("2025-01-01"),
          cpu: 4,
          memory: 16,
          storage: 100,
          bandwidth: 50,
          gpu: 1,
          cost: "25.50",
        },
      ];
      const csv = generateUsageReportCSV(data);
      expect(csv).toContain("CPU (cores)");
      expect(csv).toContain("4");
      expect(csv).toContain("25.50");
    });
  });
});

// =============================================================================
// Billing Utility Tests
// =============================================================================

function makeResources(overrides?: Partial<ResourceUsage>): ResourceUsage {
  return {
    cpu: { used: 2, limit: 4, unit: "cores", cost: "10" },
    memory: { used: 8, limit: 16, unit: "GB", cost: "20" },
    storage: { used: 50, limit: 100, unit: "GB", cost: "5" },
    bandwidth: { used: 10, limit: 100, unit: "Mbps", cost: "3" },
    ...overrides,
  };
}

function makeUsageSummary(overrides?: Partial<UsageSummary>): UsageSummary {
  return {
    period: { start: new Date("2025-01-01"), end: new Date("2025-01-31") },
    totalCost: "38",
    currency: "VIRT",
    resources: makeResources(),
    byDeployment: [],
    byProvider: [],
    ...overrides,
  };
}

function makeInvoice(overrides?: Partial<Invoice>): Invoice {
  return {
    id: "inv-1",
    number: "INV-001",
    leaseId: "lease-1",
    deploymentId: "dep-1",
    provider: "provider1",
    period: { start: new Date("2025-01-01"), end: new Date("2025-01-31") },
    status: "pending",
    currency: "VIRT",
    subtotal: "90.00",
    fees: { platformFee: "5.00", providerFee: "3.00", networkFee: "2.00" },
    total: "100.00",
    dueDate: new Date("2025-02-01"),
    createdAt: new Date("2025-01-01"),
    lineItems: [
      {
        id: "li-1",
        description: "Compute",
        resourceType: "cpu",
        quantity: "720",
        unit: "hours",
        unitPrice: "0.125",
        total: "90.00",
      },
    ],
    payments: [],
    ...overrides,
  };
}

describe("Billing Utilities", () => {
  describe("aggregateUsage", () => {
    it("returns empty summary for no input", () => {
      const result = aggregateUsage([]);
      expect(result.totalCost).toBe("0");
      expect(result.currency).toBe("VIRT");
      expect(result.byDeployment).toHaveLength(0);
    });

    it("aggregates single usage", () => {
      const usage = makeUsageSummary();
      const result = aggregateUsage([usage]);
      expect(result.totalCost).toBe("38.00");
      expect(result.resources.cpu.used).toBe(2);
    });

    it("aggregates multiple usages", () => {
      const u1 = makeUsageSummary({ totalCost: "10" });
      const u2 = makeUsageSummary({ totalCost: "20" });
      const result = aggregateUsage([u1, u2]);
      expect(result.totalCost).toBe("30.00");
      expect(result.resources.cpu.used).toBe(4);
      expect(result.resources.memory.used).toBe(16);
    });

    it("merges deployments and providers", () => {
      const u1 = makeUsageSummary({
        byDeployment: [
          {
            deploymentId: "d1",
            provider: "p1",
            resources: makeResources(),
            cost: "10",
          },
        ],
      });
      const u2 = makeUsageSummary({
        byDeployment: [
          {
            deploymentId: "d2",
            provider: "p2",
            resources: makeResources(),
            cost: "20",
          },
        ],
      });
      const result = aggregateUsage([u1, u2]);
      expect(result.byDeployment).toHaveLength(2);
    });

    it("handles GPU resources", () => {
      const u1 = makeUsageSummary();
      u1.resources.gpu = { used: 1, limit: 2, unit: "units", cost: "50" };
      const u2 = makeUsageSummary();
      u2.resources.gpu = { used: 2, limit: 4, unit: "units", cost: "100" };
      const result = aggregateUsage([u1, u2]);
      expect(result.resources.gpu?.used).toBe(3);
      expect(result.resources.gpu?.cost).toBe("150.00");
    });
  });

  describe("aggregateHistoryByTimestamp", () => {
    it("returns empty for no input", () => {
      expect(aggregateHistoryByTimestamp([])).toHaveLength(0);
    });

    it("merges same-timestamp points", () => {
      const ts = new Date("2025-01-01T12:00:00Z");
      const points: UsageHistoryPoint[] = [
        {
          timestamp: ts,
          cpu: 1,
          memory: 2,
          storage: 3,
          bandwidth: 4,
          gpu: 0,
          cost: "10",
        },
        {
          timestamp: ts,
          cpu: 2,
          memory: 3,
          storage: 4,
          bandwidth: 5,
          gpu: 1,
          cost: "20",
        },
      ];
      const result = aggregateHistoryByTimestamp(points);
      expect(result).toHaveLength(1);
      expect(result[0].cpu).toBe(3);
      expect(result[0].cost).toBe("30.00");
    });

    it("sorts by timestamp", () => {
      const points: UsageHistoryPoint[] = [
        {
          timestamp: new Date("2025-01-02"),
          cpu: 1,
          memory: 1,
          storage: 1,
          bandwidth: 1,
          gpu: 0,
          cost: "5",
        },
        {
          timestamp: new Date("2025-01-01"),
          cpu: 1,
          memory: 1,
          storage: 1,
          bandwidth: 1,
          gpu: 0,
          cost: "3",
        },
      ];
      const result = aggregateHistoryByTimestamp(points);
      expect(result[0].timestamp.getDate()).toBe(1);
      expect(result[1].timestamp.getDate()).toBe(2);
    });
  });

  describe("calculateOutstanding", () => {
    it("returns 0 for undefined", () => {
      expect(calculateOutstanding(undefined)).toBe("0");
    });

    it("returns 0 for no outstanding invoices", () => {
      const invoices = [makeInvoice({ status: "paid", total: "100.00" })];
      expect(calculateOutstanding(invoices)).toBe("0.00");
    });

    it("sums pending and overdue", () => {
      const invoices = [
        makeInvoice({ status: "pending", total: "50.00" }),
        makeInvoice({ status: "overdue", total: "30.00" }),
        makeInvoice({ status: "paid", total: "100.00" }),
      ];
      expect(calculateOutstanding(invoices)).toBe("80.00");
    });
  });

  describe("hasOverdueInvoices", () => {
    it("returns false for undefined", () => {
      expect(hasOverdueInvoices(undefined)).toBe(false);
    });

    it("returns false when no overdue", () => {
      expect(hasOverdueInvoices([makeInvoice({ status: "paid" })])).toBe(false);
    });

    it("returns true when overdue exists", () => {
      expect(hasOverdueInvoices([makeInvoice({ status: "overdue" })])).toBe(
        true,
      );
    });
  });

  describe("formatBillingAmount", () => {
    it("formats amount with currency", () => {
      expect(formatBillingAmount("100.50", "VIRT")).toContain("100");
      expect(formatBillingAmount("100.50", "VIRT")).toContain("VIRT");
    });

    it("handles NaN", () => {
      expect(formatBillingAmount("abc")).toBe("0 VIRT");
    });

    it("uses default VIRT currency", () => {
      expect(formatBillingAmount("50")).toContain("VIRT");
    });
  });

  describe("formatBillingDate", () => {
    it("formats date", () => {
      const date = new Date("2025-06-15");
      const result = formatBillingDate(date);
      expect(result).toContain("2025");
      expect(result).toContain("15");
    });
  });

  describe("formatBillingPeriod", () => {
    it("formats date range", () => {
      const start = new Date("2025-01-01");
      const end = new Date("2025-01-31");
      const result = formatBillingPeriod(start, end);
      expect(result).toContain("Jan");
      expect(result).toContain("2025");
      expect(result).toContain("-");
    });
  });

  describe("generateInvoiceText", () => {
    it("generates text representation", () => {
      const invoice = makeInvoice();
      const text = generateInvoiceText(invoice);
      expect(text).toContain("INVOICE #INV-001");
      expect(text).toContain("PENDING");
      expect(text).toContain("Compute");
      expect(text).toContain("90.00");
      expect(text).toContain("TOTAL: 100.00 VIRT");
    });

    it("includes payments if present", () => {
      const invoice = makeInvoice({
        payments: [
          {
            id: "pay-1",
            invoiceId: "inv-1",
            amount: "100.00",
            currency: "VIRT",
            status: "confirmed",
            txHash: "0xabc123",
            paidAt: new Date("2025-01-15"),
          },
        ],
      });
      const text = generateInvoiceText(invoice);
      expect(text).toContain("PAYMENTS");
      expect(text).toContain("0xabc123");
    });
  });
});

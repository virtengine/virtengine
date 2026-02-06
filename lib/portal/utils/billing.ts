/**
 * Billing Utilities
 * Task 29J: Aggregation, formatting, and PDF helpers for billing
 */

import type {
  UsageSummary,
  ResourceUsage,
  UsageMetric,
  UsageHistoryPoint,
  Invoice,
} from "../types/billing";

/**
 * Create an empty usage metric
 */
function emptyMetric(unit: string): UsageMetric {
  return { used: 0, limit: 0, unit, cost: "0" };
}

/**
 * Create empty resource usage
 */
function emptyResources(): ResourceUsage {
  return {
    cpu: emptyMetric("cores"),
    memory: emptyMetric("GB"),
    storage: emptyMetric("GB"),
    bandwidth: emptyMetric("Mbps"),
  };
}

/**
 * Add two usage metrics together
 */
function addMetrics(a: UsageMetric, b: UsageMetric): UsageMetric {
  return {
    used: a.used + b.used,
    limit: a.limit + b.limit,
    unit: a.unit,
    cost: (parseFloat(a.cost) + parseFloat(b.cost)).toFixed(2),
  };
}

/**
 * Add two resource usages together
 */
function addResources(a: ResourceUsage, b: ResourceUsage): ResourceUsage {
  const result: ResourceUsage = {
    cpu: addMetrics(a.cpu, b.cpu),
    memory: addMetrics(a.memory, b.memory),
    storage: addMetrics(a.storage, b.storage),
    bandwidth: addMetrics(a.bandwidth, b.bandwidth),
  };
  if (a.gpu || b.gpu) {
    result.gpu = addMetrics(
      a.gpu ?? emptyMetric("units"),
      b.gpu ?? emptyMetric("units"),
    );
  }
  return result;
}

/**
 * Aggregate multiple provider usage summaries into one
 */
export function aggregateUsage(usages: UsageSummary[]): UsageSummary {
  if (usages.length === 0) {
    return {
      period: { start: new Date(), end: new Date() },
      totalCost: "0",
      currency: "VIRT",
      resources: emptyResources(),
      byDeployment: [],
      byProvider: [],
    };
  }

  let resources = emptyResources();
  let totalCost = 0;
  const allDeployments: UsageSummary["byDeployment"] = [];
  const allProviders: UsageSummary["byProvider"] = [];

  let earliest = usages[0].period.start;
  let latest = usages[0].period.end;

  for (const usage of usages) {
    resources = addResources(resources, usage.resources);
    totalCost += parseFloat(usage.totalCost);
    allDeployments.push(...usage.byDeployment);
    allProviders.push(...usage.byProvider);

    if (usage.period.start < earliest) earliest = usage.period.start;
    if (usage.period.end > latest) latest = usage.period.end;
  }

  return {
    period: { start: earliest, end: latest },
    totalCost: totalCost.toFixed(2),
    currency: usages[0].currency,
    resources,
    byDeployment: allDeployments,
    byProvider: allProviders,
  };
}

/**
 * Aggregate usage history points by timestamp
 */
export function aggregateHistoryByTimestamp(
  points: UsageHistoryPoint[],
): UsageHistoryPoint[] {
  const map = new Map<string, UsageHistoryPoint>();

  for (const point of points) {
    const key = point.timestamp.toISOString();
    const existing = map.get(key);
    if (existing) {
      existing.cpu += point.cpu;
      existing.memory += point.memory;
      existing.storage += point.storage;
      existing.bandwidth += point.bandwidth;
      existing.gpu += point.gpu;
      existing.cost = (
        parseFloat(existing.cost) + parseFloat(point.cost)
      ).toFixed(2);
    } else {
      map.set(key, { ...point });
    }
  }

  return Array.from(map.values()).sort(
    (a, b) => a.timestamp.getTime() - b.timestamp.getTime(),
  );
}

/**
 * Calculate outstanding balance from invoices
 */
export function calculateOutstanding(invoices: Invoice[] | undefined): string {
  if (!invoices) return "0";
  let total = 0;
  for (const inv of invoices) {
    if (inv.status === "pending" || inv.status === "overdue") {
      total += parseFloat(inv.total);
    }
  }
  return total.toFixed(2);
}

/**
 * Check if any invoices are overdue
 */
export function hasOverdueInvoices(invoices: Invoice[] | undefined): boolean {
  if (!invoices) return false;
  return invoices.some((inv) => inv.status === "overdue");
}

/**
 * Format billing amount for display
 */
export function formatBillingAmount(
  amount: string,
  currency: string = "VIRT",
): string {
  const num = parseFloat(amount);
  if (isNaN(num)) return `0 ${currency}`;
  return `${num.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 6 })} ${currency}`;
}

/**
 * Format billing date for display
 */
export function formatBillingDate(date: Date): string {
  return date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

/**
 * Format billing date range
 */
export function formatBillingPeriod(start: Date, end: Date): string {
  const startStr = start.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
  });
  const endStr = end.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
  return `${startStr} - ${endStr}`;
}

/**
 * Generate a simple text-based invoice representation for PDF export.
 * Returns the invoice data as a formatted text blob.
 */
export function generateInvoiceText(invoice: Invoice): string {
  const lines: string[] = [
    `INVOICE #${invoice.number}`,
    `Status: ${invoice.status.toUpperCase()}`,
    `Period: ${formatBillingPeriod(invoice.period.start, invoice.period.end)}`,
    `Due: ${formatBillingDate(invoice.dueDate)}`,
    "",
    "LINE ITEMS",
    "-".repeat(60),
  ];

  for (const item of invoice.lineItems) {
    lines.push(
      `${item.description} (${item.resourceType})  ${item.quantity} ${item.unit} x ${item.unitPrice} = ${item.total} ${invoice.currency}`,
    );
  }

  lines.push("-".repeat(60));
  lines.push(`Subtotal: ${invoice.subtotal} ${invoice.currency}`);
  lines.push(`Platform Fee: ${invoice.fees.platformFee} ${invoice.currency}`);
  lines.push(`Provider Fee: ${invoice.fees.providerFee} ${invoice.currency}`);
  lines.push(`Network Fee: ${invoice.fees.networkFee} ${invoice.currency}`);
  lines.push(`TOTAL: ${invoice.total} ${invoice.currency}`);

  if (invoice.payments.length > 0) {
    lines.push("");
    lines.push("PAYMENTS");
    lines.push("-".repeat(60));
    for (const payment of invoice.payments) {
      lines.push(
        `${payment.amount} ${payment.currency} - ${payment.status} - TX: ${payment.txHash}`,
      );
    }
  }

  return lines.join("\n");
}

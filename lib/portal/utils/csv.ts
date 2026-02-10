/**
 * CSV Export Utilities
 * Task 29J: Billing data export helpers
 */

/**
 * Escape a CSV field value
 */
function escapeCSVField(value: string): string {
  if (value.includes(",") || value.includes('"') || value.includes("\n")) {
    return `"${value.replace(/"/g, '""')}"`;
  }
  return value;
}

/**
 * Convert an array of objects to CSV string
 */
export function toCSV<T extends Record<string, unknown>>(
  data: T[],
  columns: { key: keyof T; label: string }[],
): string {
  const header = columns.map((c) => escapeCSVField(c.label)).join(",");
  const rows = data.map((row) =>
    columns
      .map((c) => {
        const val = row[c.key];
        if (val === null || val === undefined) return "";
        if (val instanceof Date) return escapeCSVField(val.toISOString());
        return escapeCSVField(String(val));
      })
      .join(","),
  );
  return [header, ...rows].join("\n");
}

/**
 * Trigger a file download in the browser
 */
export function downloadFile(
  content: string | Blob,
  filename: string,
  mimeType: string,
): void {
  const blob =
    content instanceof Blob ? content : new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

/**
 * Generate invoice CSV export data
 */
export function generateInvoicesCSV(
  invoices: Array<{
    number: string;
    status: string;
    total: string;
    currency: string;
    createdAt: Date;
    dueDate: Date;
    paidAt?: Date;
    provider: string;
    deploymentId: string;
  }>,
): string {
  return toCSV(invoices, [
    { key: "number", label: "Invoice #" },
    { key: "status", label: "Status" },
    { key: "total", label: "Total" },
    { key: "currency", label: "Currency" },
    { key: "createdAt", label: "Created" },
    { key: "dueDate", label: "Due Date" },
    { key: "paidAt", label: "Paid At" },
    { key: "provider", label: "Provider" },
    { key: "deploymentId", label: "Deployment" },
  ]);
}

/**
 * Generate usage report CSV
 */
export function generateUsageReportCSV(
  data: Array<{
    timestamp: Date;
    cpu: number;
    memory: number;
    storage: number;
    bandwidth: number;
    gpu: number;
    cost: string;
  }>,
): string {
  return toCSV(data, [
    { key: "timestamp", label: "Timestamp" },
    { key: "cpu", label: "CPU (cores)" },
    { key: "memory", label: "Memory (GB)" },
    { key: "storage", label: "Storage (GB)" },
    { key: "bandwidth", label: "Bandwidth (Mbps)" },
    { key: "gpu", label: "GPU (units)" },
    { key: "cost", label: "Cost" },
  ]);
}

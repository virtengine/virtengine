/**
 * Billing Hooks
 * Task 29J: React hooks for billing data fetching and management
 *
 * @packageDocumentation
 */

import { useCallback, useEffect, useState } from "react";
import type {
  Invoice,
  InvoiceFilterOptions,
  UsageSummary,
  CostProjection,
  UsageHistoryPoint,
  UsageHistoryOptions,
  InvoiceStatus,
} from "../types/billing";
import { aggregateUsage, aggregateHistoryByTimestamp } from "../utils/billing";

const API_BASE_URL =
  typeof process !== "undefined" && process.env?.NEXT_PUBLIC_BILLING_API_URL
    ? process.env.NEXT_PUBLIC_BILLING_API_URL
    : "/api/billing";

interface ApiOptions {
  method?: "GET" | "POST";
  body?: unknown;
}

async function billingApi<T>(
  endpoint: string,
  options: ApiOptions = {},
): Promise<T> {
  const { method = "GET", body } = options;
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    method,
    headers: { "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    throw new Error(
      `Billing API error: ${response.status} ${response.statusText}`,
    );
  }

  return response.json() as Promise<T>;
}

/**
 * Hook for fetching invoice list with filters
 */
export function useInvoices(options?: InvoiceFilterOptions) {
  const [data, setData] = useState<Invoice[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetchInvoices = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams();
      if (options?.status) params.set("status", options.status);
      if (options?.startDate)
        params.set("start", options.startDate.toISOString());
      if (options?.endDate) params.set("end", options.endDate.toISOString());
      if (options?.search) params.set("search", options.search);
      if (options?.limit) params.set("limit", options.limit.toString());
      if (options?.cursor) params.set("cursor", options.cursor);

      const query = params.toString();
      const result = await billingApi<{ invoices: Invoice[] }>(
        `/invoices${query ? `?${query}` : ""}`,
      );

      const invoices = result.invoices.map((inv) => ({
        ...inv,
        period: {
          start: new Date(inv.period.start),
          end: new Date(inv.period.end),
        },
        dueDate: new Date(inv.dueDate),
        createdAt: new Date(inv.createdAt),
        paidAt: inv.paidAt ? new Date(inv.paidAt) : undefined,
        payments: inv.payments.map((p) => ({
          ...p,
          paidAt: new Date(p.paidAt),
        })),
      }));

      setData(
        invoices.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()),
      );
    } catch (err) {
      setError(
        err instanceof Error ? err : new Error("Failed to fetch invoices"),
      );
    } finally {
      setIsLoading(false);
    }
  }, [
    options?.status,
    options?.startDate,
    options?.endDate,
    options?.search,
    options?.limit,
    options?.cursor,
  ]);

  useEffect(() => {
    void fetchInvoices();
  }, [fetchInvoices]);

  return { data, isLoading, error, refetch: fetchInvoices };
}

/**
 * Hook for fetching a single invoice by ID
 */
export function useInvoice(invoiceId: string | undefined) {
  const [data, setData] = useState<Invoice | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!invoiceId) return;

    let cancelled = false;
    setIsLoading(true);
    setError(null);

    billingApi<Invoice>(`/invoices/${invoiceId}`)
      .then((inv) => {
        if (cancelled) return;
        setData({
          ...inv,
          period: {
            start: new Date(inv.period.start),
            end: new Date(inv.period.end),
          },
          dueDate: new Date(inv.dueDate),
          createdAt: new Date(inv.createdAt),
          paidAt: inv.paidAt ? new Date(inv.paidAt) : undefined,
          payments: inv.payments.map((p) => ({
            ...p,
            paidAt: new Date(p.paidAt),
          })),
        });
      })
      .catch((err) => {
        if (cancelled) return;
        setError(
          err instanceof Error ? err : new Error("Failed to fetch invoice"),
        );
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [invoiceId]);

  return { data, isLoading, error };
}

/**
 * Hook for fetching current usage summary
 */
export function useCurrentUsage() {
  const [data, setData] = useState<UsageSummary | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetchUsage = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const result = await billingApi<{ usages: UsageSummary[] }>("/usage");
      const usages = result.usages.map((u) => ({
        ...u,
        period: {
          start: new Date(u.period.start),
          end: new Date(u.period.end),
        },
      }));
      setData(aggregateUsage(usages));
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Failed to fetch usage"));
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchUsage();
    const interval = setInterval(() => void fetchUsage(), 60_000);
    return () => clearInterval(interval);
  }, [fetchUsage]);

  return { data, isLoading, error, refetch: fetchUsage };
}

/**
 * Hook for fetching usage history
 */
export function useUsageHistory(options: UsageHistoryOptions) {
  const [data, setData] = useState<UsageHistoryPoint[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    setIsLoading(true);
    setError(null);

    const params = new URLSearchParams({
      start: options.startDate.toISOString(),
      end: options.endDate.toISOString(),
      granularity: options.granularity,
    });

    billingApi<{ data: UsageHistoryPoint[] }>(`/usage/history?${params}`)
      .then((result) => {
        if (cancelled) return;
        const points = result.data.map((p) => ({
          ...p,
          timestamp: new Date(p.timestamp),
        }));
        setData(aggregateHistoryByTimestamp(points));
      })
      .catch((err) => {
        if (cancelled) return;
        setError(
          err instanceof Error
            ? err
            : new Error("Failed to fetch usage history"),
        );
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [options.startDate, options.endDate, options.granularity]);

  return { data, isLoading, error };
}

/**
 * Hook for fetching cost projection
 */
export function useCostProjection() {
  const [data, setData] = useState<CostProjection | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    setIsLoading(true);
    setError(null);

    billingApi<CostProjection>("/usage/projection")
      .then((result) => {
        if (!cancelled) setData(result);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(
          err instanceof Error
            ? err
            : new Error("Failed to fetch cost projection"),
        );
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  return { data, isLoading, error };
}

// Re-export types for convenience
export type {
  Invoice,
  InvoiceStatus,
  InvoiceFilterOptions,
  UsageSummary,
  CostProjection,
  UsageHistoryPoint,
  UsageHistoryOptions,
};

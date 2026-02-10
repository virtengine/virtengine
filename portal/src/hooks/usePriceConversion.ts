/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';

const RATE_STORAGE_KEY = 've_uve_usd_rate';
const REFRESH_INTERVAL_MS = 5 * 60 * 1000;

type StoredRate = {
  rate: number;
  fetchedAt: string;
};

type PriceResponse = {
  rate: number;
  fetchedAt: string;
  source: string;
  ageSeconds: number;
  stale: boolean;
  fallback: boolean;
  ttlSeconds: number;
};

function readStoredRate(): StoredRate | null {
  if (typeof window === 'undefined') return null;
  try {
    const stored = window.localStorage.getItem(RATE_STORAGE_KEY);
    if (!stored) return null;
    return JSON.parse(stored) as StoredRate;
  } catch {
    return null;
  }
}

function writeStoredRate(rate: StoredRate) {
  if (typeof window === 'undefined') return;
  try {
    window.localStorage.setItem(RATE_STORAGE_KEY, JSON.stringify(rate));
  } catch {
    // Ignore storage errors.
  }
}

export function usePriceConversion() {
  const stored = readStoredRate();
  const [rate, setRate] = useState<number | null>(stored?.rate ?? null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(
    stored?.fetchedAt ? new Date(stored.fetchedAt) : null
  );
  const [stale, setStale] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setIsLoading(true);
    try {
      const response = await fetch('/api/price/uve-usd');
      if (!response.ok) {
        throw new Error(`Rate fetch failed: ${response.status}`);
      }

      const data = (await response.json()) as PriceResponse;
      if (!Number.isFinite(data.rate)) {
        throw new Error('Rate missing from response');
      }

      setRate(data.rate);
      setLastUpdated(new Date(data.fetchedAt));
      setStale(data.stale);
      setError(null);
      writeStoredRate({ rate: data.rate, fetchedAt: data.fetchedAt });
    } catch (err) {
      const fallback = readStoredRate();
      if (fallback) {
        setRate(fallback.rate);
        setLastUpdated(new Date(fallback.fetchedAt));
        setStale(true);
      }
      setError(err instanceof Error ? err.message : 'Failed to load rate');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
    const interval = setInterval(() => {
      void refresh();
    }, REFRESH_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [refresh]);

  const uveToUsd = useMemo(() => {
    return (uve: number) => (rate ? uve * rate : null);
  }, [rate]);

  const usdToUve = useMemo(() => {
    return (usd: number) => (rate ? usd / rate : null);
  }, [rate]);

  return { uveToUsd, usdToUve, rate, stale, lastUpdated, isLoading, error, refresh };
}

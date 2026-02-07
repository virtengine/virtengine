/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { NextResponse } from 'next/server';
import { getRedisClient } from '@/lib/redis';
import {
  recordPriceFetchError,
  recordPriceFetchSuccess,
  recordPriceFreshness,
} from '@/lib/price-metrics';

export const runtime = 'nodejs';

const CACHE_TTL_SECONDS = 5 * 60;
const LAST_RATE_TTL_SECONDS = 30 * 24 * 60 * 60;
const MEMORY_CACHE_TTL_MS = CACHE_TTL_SECONDS * 1000;
const PRICE_KEY_CURRENT = 'portal:price:uve-usd:current';
const PRICE_KEY_LAST = 'portal:price:uve-usd:last';
const BUILD_PHASES = new Set(['phase-production-build', 'phase-production-export']);

type CachedRate = {
  rate: number;
  fetchedAt: string;
  source: string;
};

type CacheEntry = {
  value: CachedRate;
  expiresAt: number;
};

const globalForCache = globalThis as unknown as {
  vePriceCache?: {
    current?: CacheEntry;
    last?: CachedRate;
  };
};

function getMemoryCache() {
  if (!globalForCache.vePriceCache) {
    globalForCache.vePriceCache = {};
  }
  return globalForCache.vePriceCache;
}

function getFreshMemoryRate(now: number): CachedRate | null {
  const cache = getMemoryCache();
  if (cache.current && cache.current.expiresAt > now) {
    return cache.current.value;
  }
  return null;
}

function getLastMemoryRate(): CachedRate | null {
  const cache = getMemoryCache();
  return cache.last ?? null;
}

function setMemoryRate(rate: CachedRate, now: number) {
  const cache = getMemoryCache();
  cache.current = {
    value: rate,
    expiresAt: now + MEMORY_CACHE_TTL_MS,
  };
  cache.last = rate;
}

async function getRedisRate(key: string): Promise<CachedRate | null> {
  const client = await getRedisClient();
  if (!client) return null;

  try {
    const value = await client.get(key);
    if (!value) return null;
    return JSON.parse(value) as CachedRate;
  } catch (err) {
    console.error('Redis get failed:', err);
    return null;
  }
}

async function setRedisRate(key: string, rate: CachedRate, ttlSeconds?: number) {
  const client = await getRedisClient();
  if (!client) return;

  try {
    const payload = JSON.stringify(rate);
    if (ttlSeconds) {
      await client.set(key, payload, { EX: ttlSeconds });
    } else {
      await client.set(key, payload);
    }
  } catch (err) {
    console.error('Redis set failed:', err);
  }
}

function getPriceSource(): 'coingecko' | 'coinmarketcap' {
  const source = process.env.UVE_PRICE_SOURCE?.toLowerCase();
  if (source === 'coinmarketcap') return 'coinmarketcap';
  return 'coingecko';
}

async function fetchFromCoinGecko(): Promise<CachedRate> {
  const apiUrl = process.env.COINGECKO_API_URL ?? 'https://api.coingecko.com/api/v3';
  const coinId = process.env.COINGECKO_COIN_ID ?? 'virtengine';
  const vsCurrency = process.env.COINGECKO_VS_CURRENCY ?? 'usd';
  const url = `${apiUrl}/simple/price?ids=${encodeURIComponent(coinId)}&vs_currencies=${encodeURIComponent(vsCurrency)}`;

  const response = await fetch(url, {
    headers: {
      Accept: 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error(`CoinGecko response ${response.status}`);
  }

  const data = (await response.json()) as Record<string, Record<string, number>>;
  const rate = data?.[coinId]?.[vsCurrency];
  if (!rate || !Number.isFinite(rate)) {
    throw new Error('CoinGecko rate missing');
  }

  return {
    rate,
    fetchedAt: new Date().toISOString(),
    source: 'coingecko',
  };
}

async function fetchFromCoinMarketCap(): Promise<CachedRate> {
  const apiUrl =
    process.env.COINMARKETCAP_API_URL ??
    'https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest';
  const symbol = process.env.COINMARKETCAP_SYMBOL ?? 'UVE';
  const apiKey = process.env.COINMARKETCAP_API_KEY ?? process.env.CMC_API_KEY;
  if (!apiKey) {
    throw new Error('CoinMarketCap API key missing');
  }

  const url = new URL(apiUrl);
  url.searchParams.set('symbol', symbol);
  url.searchParams.set('convert', 'USD');

  const response = await fetch(url.toString(), {
    headers: {
      Accept: 'application/json',
      'X-CMC_PRO_API_KEY': apiKey,
    },
  });

  if (!response.ok) {
    throw new Error(`CoinMarketCap response ${response.status}`);
  }

  const data = (await response.json()) as {
    data?: Record<string, { quote?: { USD?: { price?: number } } }>;
  };
  const rate = data?.data?.[symbol]?.quote?.USD?.price;
  if (!rate || !Number.isFinite(rate)) {
    throw new Error('CoinMarketCap rate missing');
  }

  return {
    rate,
    fetchedAt: new Date().toISOString(),
    source: 'coinmarketcap',
  };
}

async function fetchExternalRate(): Promise<CachedRate> {
  const source = getPriceSource();
  const rate =
    source === 'coinmarketcap' ? await fetchFromCoinMarketCap() : await fetchFromCoinGecko();
  recordPriceFetchSuccess(rate.source);
  return rate;
}

function buildResponse(rate: CachedRate, fallback: boolean) {
  const fetchedAt = new Date(rate.fetchedAt);
  const ageSeconds = Math.max(0, Math.floor((Date.now() - fetchedAt.getTime()) / 1000));
  const stale = ageSeconds > CACHE_TTL_SECONDS;

  recordPriceFreshness(ageSeconds, stale, fetchedAt);

  return NextResponse.json(
    {
      rate: rate.rate,
      fetchedAt: rate.fetchedAt,
      source: rate.source,
      ageSeconds,
      stale,
      fallback,
      ttlSeconds: CACHE_TTL_SECONDS,
    },
    {
      headers: {
        'Cache-Control': `s-maxage=${CACHE_TTL_SECONDS}, stale-while-revalidate=60`,
      },
    }
  );
}

export async function GET() {
  const now = Date.now();
  const isBuildTime = BUILD_PHASES.has(process.env.NEXT_PHASE ?? '');

  const memoryRate = getFreshMemoryRate(now);
  if (memoryRate) {
    return buildResponse(memoryRate, false);
  }

  const redisRate = await getRedisRate(PRICE_KEY_CURRENT);
  if (redisRate) {
    setMemoryRate(redisRate, now);
    return buildResponse(redisRate, false);
  }

  if (isBuildTime) {
    const fallbackRate = getLastMemoryRate() ?? (await getRedisRate(PRICE_KEY_LAST));
    if (fallbackRate) {
      setMemoryRate(fallbackRate, now);
      return buildResponse(fallbackRate, true);
    }

    return NextResponse.json(
      {
        error: 'Rate unavailable',
      },
      { status: 503 }
    );
  }

  try {
    const rate = await fetchExternalRate();
    setMemoryRate(rate, now);
    await setRedisRate(PRICE_KEY_CURRENT, rate, CACHE_TTL_SECONDS);
    await setRedisRate(PRICE_KEY_LAST, rate, LAST_RATE_TTL_SECONDS);
    return buildResponse(rate, false);
  } catch (err) {
    recordPriceFetchError(getPriceSource());
    console.error('Failed to fetch UVE/USD rate:', err);
  }

  const fallbackRate = getLastMemoryRate() ?? (await getRedisRate(PRICE_KEY_LAST));
  if (fallbackRate) {
    setMemoryRate(fallbackRate, now);
    return buildResponse(fallbackRate, true);
  }

  return NextResponse.json(
    {
      error: 'Rate unavailable',
    },
    { status: 503 }
  );
}

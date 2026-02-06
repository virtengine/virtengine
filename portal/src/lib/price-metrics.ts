/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import client from 'prom-client';

type PriceMetrics = {
  registry: client.Registry;
  rateAgeSeconds: client.Gauge;
  rateStale: client.Gauge;
  rateFetchErrors: client.Counter;
  rateFetchSuccess: client.Counter;
  rateLastUpdated: client.Gauge;
};

const globalForMetrics = globalThis as unknown as { vePriceMetrics?: PriceMetrics };

function initMetrics(): PriceMetrics {
  const registry = new client.Registry();

  client.collectDefaultMetrics({ register: registry });

  const rateAgeSeconds = new client.Gauge({
    name: 'portal_uve_usd_rate_age_seconds',
    help: 'Age in seconds of the most recently fetched UVE/USD rate.',
    registers: [registry],
  });

  const rateStale = new client.Gauge({
    name: 'portal_uve_usd_rate_stale',
    help: 'Whether the UVE/USD rate used by the portal is stale (1) or fresh (0).',
    registers: [registry],
  });

  const rateLastUpdated = new client.Gauge({
    name: 'portal_uve_usd_rate_last_updated',
    help: 'Unix timestamp of the most recent UVE/USD rate update.',
    registers: [registry],
  });

  const rateFetchErrors = new client.Counter({
    name: 'portal_uve_usd_rate_fetch_errors_total',
    help: 'Count of errors when fetching the UVE/USD rate.',
    labelNames: ['source'],
    registers: [registry],
  });

  const rateFetchSuccess = new client.Counter({
    name: 'portal_uve_usd_rate_fetch_success_total',
    help: 'Count of successful UVE/USD rate fetches.',
    labelNames: ['source'],
    registers: [registry],
  });

  return {
    registry,
    rateAgeSeconds,
    rateStale,
    rateFetchErrors,
    rateFetchSuccess,
    rateLastUpdated,
  };
}

export function getPriceMetrics(): PriceMetrics {
  if (!globalForMetrics.vePriceMetrics) {
    globalForMetrics.vePriceMetrics = initMetrics();
  }

  return globalForMetrics.vePriceMetrics;
}

export function recordPriceFreshness(ageSeconds: number, stale: boolean, fetchedAt: Date) {
  const metrics = getPriceMetrics();
  metrics.rateAgeSeconds.set(ageSeconds);
  metrics.rateStale.set(stale ? 1 : 0);
  metrics.rateLastUpdated.set(Math.floor(fetchedAt.getTime() / 1000));
}

export function recordPriceFetchError(source: string) {
  const metrics = getPriceMetrics();
  metrics.rateFetchErrors.inc({ source });
}

export function recordPriceFetchSuccess(source: string) {
  const metrics = getPriceMetrics();
  metrics.rateFetchSuccess.inc({ source });
}

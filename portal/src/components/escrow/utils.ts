/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { formatCurrency } from '@/lib/utils';
import type { FiatRates } from './data';

export function formatToken(amount: number, currency: string): string {
  return `${amount.toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })} ${currency}`;
}

export function formatFiat(amount: number, currency: 'USD' | 'EUR'): string {
  return formatCurrency(amount, currency);
}

export function formatFiatEstimates(amount: number, rates: FiatRates): string[] {
  const estimates: string[] = [];
  if (typeof rates.usd === 'number' && Number.isFinite(rates.usd) && rates.usd > 0) {
    estimates.push(`${formatFiat(amount * rates.usd, 'USD')} USD`);
  }
  if (typeof rates.eur === 'number' && Number.isFinite(rates.eur) && rates.eur > 0) {
    estimates.push(`${formatFiat(amount * rates.eur, 'EUR')} EUR`);
  }
  return estimates;
}

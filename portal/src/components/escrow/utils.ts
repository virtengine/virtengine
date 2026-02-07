/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { formatCurrency } from '@/lib/utils';

export function formatToken(amount: number, currency: string): string {
  return `${amount.toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })} ${currency}`;
}

export function formatFiat(amount: number, currency: 'USD' | 'EUR'): string {
  return formatCurrency(amount, currency);
}

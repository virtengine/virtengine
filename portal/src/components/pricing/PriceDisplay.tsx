/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useMemo } from 'react';
import { formatTokenAmount } from '@/features/orders';
import { usePriceConversion } from '@/hooks/usePriceConversion';

const MICRO_UVE = 1_000_000;

export interface PriceDisplayProps {
  amount: number;
  denom: string;
  showUsd?: boolean;
  decimals?: number;
  className?: string;
}

function formatUsd(value: number): string {
  const precision = value < 0.01 ? 4 : 2;
  return `$${value.toFixed(precision)} USD`;
}

export function PriceDisplay({
  amount,
  denom,
  showUsd = false,
  decimals = 2,
  className,
}: PriceDisplayProps) {
  const { uveToUsd, rate } = usePriceConversion();

  const displayValue = useMemo(() => {
    if (denom === 'uve') {
      return amount / MICRO_UVE;
    }
    return amount;
  }, [amount, denom]);

  const usdValue = useMemo(() => {
    if (!showUsd || denom !== 'uve' || !rate) return null;
    return uveToUsd(displayValue);
  }, [showUsd, denom, rate, uveToUsd, displayValue]);

  return (
    <span className={className}>
      {formatTokenAmount(displayValue, decimals)} {denom === 'uve' ? 'VE' : denom.toUpperCase()}
      {showUsd && usdValue !== null && (
        <span className="text-muted-foreground"> (~{formatUsd(usdValue)})</span>
      )}
      {showUsd && usdValue === null && <span className="text-muted-foreground"> (~—)</span>}
    </span>
  );
}

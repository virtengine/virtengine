/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';

interface UsageSummaryCardProps {
  title: string;
  value: string;
  currency?: string;
  subtitle?: string;
  loading?: boolean;
  status?: 'normal' | 'warning';
}

export function UsageSummaryCard({
  title,
  value,
  currency,
  subtitle,
  loading,
  status = 'normal',
}: UsageSummaryCardProps) {
  return (
    <Card className={status === 'warning' ? 'border-warning' : undefined}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <Skeleton className="h-8 w-24" />
        ) : (
          <>
            <div className="text-2xl font-bold">
              {value}
              {currency && (
                <span className="ml-1 text-sm font-normal text-muted-foreground">{currency}</span>
              )}
            </div>
            {subtitle && <p className="text-xs text-muted-foreground">{subtitle}</p>}
          </>
        )}
      </CardContent>
    </Card>
  );
}

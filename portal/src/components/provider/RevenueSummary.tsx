/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import Link from 'next/link';
import { useProviderStore } from '@/stores/providerStore';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { formatCurrency } from '@/lib/utils';
import type { RevenueSummaryData } from '@/types/provider';

function ChangeIndicator({ value }: { value: number }) {
  const isPositive = value >= 0;
  return (
    <span
      className={`text-sm font-medium ${isPositive ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}
    >
      {isPositive ? '↑' : '↓'} {Math.abs(value).toFixed(1)}%
    </span>
  );
}

function RevenueBreakdown({ byOffering }: { byOffering: RevenueSummaryData['byOffering'] }) {
  return (
    <div className="space-y-3">
      {byOffering.map((item) => (
        <div key={item.offeringName}>
          <div className="flex justify-between text-sm">
            <span className="truncate pr-4">{item.offeringName}</span>
            <span className="font-medium">{formatCurrency(item.revenue)}</span>
          </div>
          <div className="mt-1 h-2 rounded-full bg-muted">
            <div
              className="h-full rounded-full bg-primary"
              style={{ width: `${item.percentage}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}

function RevenueHistory({ history }: { history: RevenueSummaryData['history'] }) {
  const maxRevenue = Math.max(...history.map((h) => h.revenue));

  return (
    <div className="flex items-end gap-2">
      {history.map((h) => {
        const height = maxRevenue > 0 ? (h.revenue / maxRevenue) * 100 : 0;
        return (
          <div key={h.period} className="flex flex-1 flex-col items-center gap-1">
            <span className="text-xs font-medium">{formatCurrency(h.revenue)}</span>
            <div
              className="w-full rounded-t bg-primary/80 transition-all hover:bg-primary"
              style={{ height: `${Math.max(height, 4)}px`, minHeight: '4px', maxHeight: '80px' }}
              title={`${h.period}: ${formatCurrency(h.revenue)} (${h.orders} orders)`}
            />
            <span className="text-2xs text-muted-foreground">
              {h.period.split(' ')[0]?.substring(0, 3)}
            </span>
          </div>
        );
      })}
    </div>
  );
}

export default function RevenueSummary() {
  const revenue = useProviderStore((s) => s.revenue);

  return (
    <div className="space-y-6">
      {/* Revenue Stats */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Current Month
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(revenue.currentMonth)}</div>
            <ChangeIndicator value={revenue.changePercent} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Previous Month
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(revenue.previousMonth)}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Lifetime Revenue
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(revenue.totalLifetime)}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Pending Payouts
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(revenue.pendingPayouts)}</div>
            <Badge variant="warning" size="sm">
              Pending
            </Badge>
          </CardContent>
        </Card>
      </div>

      {/* Revenue Chart and Breakdown */}
      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-lg">Revenue History</CardTitle>
              <Link href="/provider/orders" className="text-sm text-primary hover:underline">
                View orders
              </Link>
            </div>
          </CardHeader>
          <CardContent>
            <RevenueHistory history={revenue.history} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Revenue by Offering</CardTitle>
          </CardHeader>
          <CardContent>
            <RevenueBreakdown byOffering={revenue.byOffering} />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { formatCurrency } from '@/lib/utils';
import type { BillingSummaryData } from '@/types/customer';
import { useTranslation } from 'react-i18next';

interface BillingSummaryProps {
  billing: BillingSummaryData;
}

export function BillingSummary({ billing }: BillingSummaryProps) {
  const { t } = useTranslation();
  const changeIsPositive = billing.changePercent > 0;

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">{t('Billing Summary')}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Current period */}
        <div>
          <p className="text-sm text-muted-foreground">{t('Current period')}</p>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold">{formatCurrency(billing.currentPeriodCost)}</span>
            <Badge variant={changeIsPositive ? 'destructive' : 'success'} size="sm">
              {changeIsPositive ? '+' : ''}
              {billing.changePercent.toFixed(1)}%
            </Badge>
          </div>
        </div>

        {/* Outstanding */}
        {billing.outstandingBalance > 0 && (
          <div className="flex items-center justify-between rounded-md bg-warning/10 px-3 py-2">
            <span className="text-sm font-medium">{t('Outstanding')}</span>
            <span className="text-sm font-semibold">
              {formatCurrency(billing.outstandingBalance)}
            </span>
          </div>
        )}

        {/* Lifetime */}
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">{t('Lifetime spend')}</span>
          <span className="font-medium">{formatCurrency(billing.totalLifetimeSpend)}</span>
        </div>

        {/* By provider breakdown */}
        <div className="space-y-2">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            {t('By Provider')}
          </p>
          {billing.byProvider.map((bp) => (
            <div key={bp.providerName} className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">{bp.providerName}</span>
              <span>
                {formatCurrency(bp.amount)}{' '}
                <span className="text-xs text-muted-foreground">
                  {t('({{percentage}}%)', { percentage: bp.percentage })}
                </span>
              </span>
            </div>
          ))}
        </div>

        {/* Spend history (last 3 periods) */}
        <div className="space-y-2">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            {t('Recent History')}
          </p>
          {billing.history.slice(-3).map((h) => (
            <div key={h.period} className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">{h.period}</span>
              <span>
                {formatCurrency(h.amount)}{' '}
                <span className="text-xs text-muted-foreground">
                  {t('({{count}} orders)', { count: h.orders })}
                </span>
              </span>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

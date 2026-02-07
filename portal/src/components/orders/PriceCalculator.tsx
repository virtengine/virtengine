/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Separator } from '@/components/ui/Separator';
import type { ResourceConfig, PriceBreakdown } from '@/features/orders';
import { formatTokenAmount, durationToHours } from '@/features/orders';
import { usePriceConversion } from '@/hooks/usePriceConversion';
import { useTranslation } from 'react-i18next';
import { formatCurrency } from '@/lib/utils';

interface PriceCalculatorProps {
  resources: ResourceConfig;
  priceBreakdown: PriceBreakdown;
}

/**
 * Step 2: Price Calculator
 * Shows real-time price calculation with line-item breakdown.
 */
function formatUsd(value: number): string {
  const precision = value < 0.01 ? 4 : 2;
  return formatCurrency(Number(value.toFixed(precision)), 'USD');
}

export function PriceCalculator({ resources, priceBreakdown }: PriceCalculatorProps) {
  const { t } = useTranslation();
  const totalHours = durationToHours(resources.duration, resources.durationUnit);
  const { uveToUsd, isLoading: rateLoading } = usePriceConversion();

  return (
    <div className="space-y-6">
      {/* Resource Summary */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">{t('Resource Summary')}</CardTitle>
          <CardDescription>{t('Your configured deployment resources')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <ResourceSummaryItem
              label={t('CPU')}
              value={t('{{count}} vCPU', { count: resources.cpu })}
            />
            <ResourceSummaryItem
              label={t('Memory')}
              value={t('{{count}} GB', { count: resources.memory })}
            />
            <ResourceSummaryItem
              label={t('Storage')}
              value={t('{{count}} GB', { count: resources.storage })}
            />
            {resources.gpu > 0 && (
              <ResourceSummaryItem
                label={t('GPU')}
                value={t('{{count}} GPU', { count: resources.gpu })}
              />
            )}
          </div>
          <div className="mt-4 rounded-md bg-muted/50 px-4 py-2 text-sm text-muted-foreground">
            {t('Duration: {{duration}} {{unit}} ({{hours}} hours total)', {
              duration: resources.duration,
              unit: t(resources.durationUnit),
              hours: totalHours,
            })}
            {resources.region && ` • ${t('Region')}: ${resources.region}`}
          </div>
        </CardContent>
      </Card>

      {/* Price Breakdown */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">{t('Price Breakdown')}</CardTitle>
          <CardDescription>{t('Estimated cost for your deployment')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {priceBreakdown.items.map((item) => (
              <div key={item.resourceType} className="flex items-center justify-between text-sm">
                <div>
                  <span className="font-medium">{item.label}</span>
                  <span className="ml-2 text-muted-foreground">
                    {t('@ {{price}} {{currency}}/{{unit}}', {
                      price: formatTokenAmount(item.unitPrice, 4),
                      currency: priceBreakdown.currency,
                      unit: item.unit,
                    })}
                  </span>
                </div>
                <span className="font-medium">
                  {t('{{amount}} {{currency}}', {
                    amount: formatTokenAmount(item.total),
                    currency: priceBreakdown.currency,
                  })}
                  {!rateLoading && uveToUsd(item.total) !== null && (
                    <span className="ml-1 text-muted-foreground">
                      {t('(~{{amount}})', { amount: formatUsd(uveToUsd(item.total)!) })}
                    </span>
                  )}
                </span>
              </div>
            ))}

            <Separator />

            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">{t('Subtotal')}</span>
              <span className="font-medium">
                {t('{{amount}} {{currency}}', {
                  amount: formatTokenAmount(priceBreakdown.subtotal),
                  currency: priceBreakdown.currency,
                })}
                {!rateLoading && uveToUsd(priceBreakdown.subtotal) !== null && (
                  <span className="ml-1 text-muted-foreground">
                    {t('(~{{amount}})', { amount: formatUsd(uveToUsd(priceBreakdown.subtotal)!) })}
                  </span>
                )}
              </span>
            </div>

            <div className="flex items-center justify-between text-sm">
              <div>
                <span className="text-muted-foreground">{t('Escrow Deposit')}</span>
                <span className="ml-1 text-xs text-muted-foreground">{t('(refundable)')}</span>
              </div>
              <span className="font-medium text-primary">
                {t('{{amount}} {{currency}}', {
                  amount: formatTokenAmount(priceBreakdown.escrowDeposit),
                  currency: priceBreakdown.currency,
                })}
                {!rateLoading && uveToUsd(priceBreakdown.escrowDeposit) !== null && (
                  <span className="ml-1 text-muted-foreground">
                    {t('(~{{amount}})', {
                      amount: formatUsd(uveToUsd(priceBreakdown.escrowDeposit)!),
                    })}
                  </span>
                )}
              </span>
            </div>

            <Separator />

            <div className="flex items-center justify-between">
              <span className="text-base font-semibold">{t('Estimated Total')}</span>
              <div className="text-right">
                <span className="text-lg font-bold text-primary">
                  {t('{{amount}} {{currency}}', {
                    amount: formatTokenAmount(priceBreakdown.estimatedTotal),
                    currency: priceBreakdown.currency,
                  })}
                </span>
                {!rateLoading && uveToUsd(priceBreakdown.estimatedTotal) !== null && (
                  <p className="text-sm text-muted-foreground">
                    {t('~{{amount}} USD', {
                      amount: formatUsd(uveToUsd(priceBreakdown.estimatedTotal)!),
                    })}
                  </p>
                )}
              </div>
            </div>
          </div>

          <div className="mt-4 rounded-md border border-blue-200 bg-blue-50 p-3 text-sm text-blue-800 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-200">
            <strong>{t('Note:')}</strong>{' '}
            {t(
              'The escrow deposit is held on-chain and refunded when the order completes. Actual costs are settled based on usage.'
            )}
          </div>
        </CardContent>
      </Card>

      {/* Hourly Rate Card */}
      <Card>
        <CardContent className="py-4">
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">{t('Effective Hourly Rate')}</span>
            <div className="text-right">
              <span className="text-lg font-semibold">
                {totalHours > 0
                  ? t('{{amount}} {{currency}}/hr', {
                      amount: formatTokenAmount(priceBreakdown.subtotal / totalHours, 4),
                      currency: priceBreakdown.currency,
                    })
                  : t('—')}
              </span>
              {totalHours > 0 &&
                !rateLoading &&
                uveToUsd(priceBreakdown.subtotal / totalHours) !== null && (
                  <p className="text-sm text-muted-foreground">
                    {t('~{{amount}}/hr', {
                      amount: formatUsd(uveToUsd(priceBreakdown.subtotal / totalHours)!),
                    })}
                  </p>
                )}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function ResourceSummaryItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg bg-muted/50 p-3 text-center">
      <div className="text-lg font-bold">{value}</div>
      <div className="text-xs text-muted-foreground">{label}</div>
    </div>
  );
}

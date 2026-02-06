/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Separator } from '@/components/ui/Separator';
import type { ResourceConfig, PriceBreakdown } from '@/features/orders';
import { formatTokenAmount, durationToHours } from '@/features/orders';

interface PriceCalculatorProps {
  resources: ResourceConfig;
  priceBreakdown: PriceBreakdown;
}

/**
 * Step 2: Price Calculator
 * Shows real-time price calculation with line-item breakdown.
 */
export function PriceCalculator({ resources, priceBreakdown }: PriceCalculatorProps) {
  const totalHours = durationToHours(resources.duration, resources.durationUnit);

  return (
    <div className="space-y-6">
      {/* Resource Summary */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Resource Summary</CardTitle>
          <CardDescription>Your configured deployment resources</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <ResourceSummaryItem label="CPU" value={`${resources.cpu} vCPU`} />
            <ResourceSummaryItem label="Memory" value={`${resources.memory} GB`} />
            <ResourceSummaryItem label="Storage" value={`${resources.storage} GB`} />
            {resources.gpu > 0 && (
              <ResourceSummaryItem label="GPU" value={`${resources.gpu} GPU`} />
            )}
          </div>
          <div className="mt-4 rounded-md bg-muted/50 px-4 py-2 text-sm text-muted-foreground">
            Duration: {resources.duration} {resources.durationUnit} ({totalHours} hours total)
            {resources.region && ` • Region: ${resources.region}`}
          </div>
        </CardContent>
      </Card>

      {/* Price Breakdown */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Price Breakdown</CardTitle>
          <CardDescription>Estimated cost for your deployment</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {priceBreakdown.items.map((item) => (
              <div key={item.resourceType} className="flex items-center justify-between text-sm">
                <div>
                  <span className="font-medium">{item.label}</span>
                  <span className="ml-2 text-muted-foreground">
                    @ {formatTokenAmount(item.unitPrice, 4)} {priceBreakdown.currency}/{item.unit}
                  </span>
                </div>
                <span className="font-medium">
                  {formatTokenAmount(item.total)} {priceBreakdown.currency}
                </span>
              </div>
            ))}

            <Separator />

            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Subtotal</span>
              <span className="font-medium">
                {formatTokenAmount(priceBreakdown.subtotal)} {priceBreakdown.currency}
              </span>
            </div>

            <div className="flex items-center justify-between text-sm">
              <div>
                <span className="text-muted-foreground">Escrow Deposit</span>
                <span className="ml-1 text-xs text-muted-foreground">(refundable)</span>
              </div>
              <span className="font-medium text-primary">
                {formatTokenAmount(priceBreakdown.escrowDeposit)} {priceBreakdown.currency}
              </span>
            </div>

            <Separator />

            <div className="flex items-center justify-between">
              <span className="text-base font-semibold">Estimated Total</span>
              <span className="text-lg font-bold text-primary">
                {formatTokenAmount(priceBreakdown.estimatedTotal)} {priceBreakdown.currency}
              </span>
            </div>
          </div>

          <div className="mt-4 rounded-md border border-blue-200 bg-blue-50 p-3 text-sm text-blue-800 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-200">
            <strong>Note:</strong> The escrow deposit is held on-chain and refunded when the order
            completes. Actual costs are settled based on usage.
          </div>
        </CardContent>
      </Card>

      {/* Hourly Rate Card */}
      <Card>
        <CardContent className="py-4">
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Effective Hourly Rate</span>
            <span className="text-lg font-semibold">
              {totalHours > 0
                ? `${formatTokenAmount(priceBreakdown.subtotal / totalHours, 4)} ${priceBreakdown.currency}/hr`
                : '—'}
            </span>
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

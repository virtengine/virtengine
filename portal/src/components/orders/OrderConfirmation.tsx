/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import type { OrderCreateResult, PriceBreakdown, ResourceConfig } from '@/features/orders';
import { formatTokenAmount, durationToHours } from '@/features/orders';

interface OrderConfirmationProps {
  orderResult: OrderCreateResult;
  resources: ResourceConfig;
  priceBreakdown: PriceBreakdown;
  offeringName: string;
}

/**
 * Step 4: Order Confirmation
 * Shows order result and provides links to tracking.
 */
export function OrderConfirmation({
  orderResult,
  resources,
  priceBreakdown,
  offeringName,
}: OrderConfirmationProps) {
  const totalHours = durationToHours(resources.duration, resources.durationUnit);
  const statusColor =
    orderResult.status === 'matched'
      ? 'bg-green-500'
      : orderResult.status === 'pending'
        ? 'bg-yellow-500'
        : 'bg-red-500';

  return (
    <div className="space-y-6">
      {/* Success Banner */}
      <div className="rounded-lg border border-green-200 bg-green-50 p-6 text-center dark:border-green-800 dark:bg-green-950">
        <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-green-100 dark:bg-green-900">
          <svg
            className="h-6 w-6 text-green-600 dark:text-green-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <h2 className="text-xl font-bold text-green-800 dark:text-green-200">
          Order Created Successfully
        </h2>
        <p className="mt-1 text-sm text-green-700 dark:text-green-300">
          Your order has been submitted to the marketplace
        </p>
      </div>

      {/* Order Details */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Order Details</CardTitle>
          <CardDescription>Reference information for your order</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Order ID</span>
              <code className="rounded bg-muted px-2 py-0.5 font-mono text-xs">
                {orderResult.orderId}
              </code>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Transaction Hash</span>
              <code className="max-w-[200px] truncate rounded bg-muted px-2 py-0.5 font-mono text-xs">
                {orderResult.txHash}
              </code>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Status</span>
              <Badge className={`${statusColor} text-white`}>{orderResult.status}</Badge>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Offering</span>
              <span className="font-medium">{offeringName}</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Created At</span>
              <span className="font-medium">
                {new Date(orderResult.createdAt).toLocaleString()}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Resource & Cost Summary */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Resource &amp; Cost Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-2">
            <SummaryItem label="CPU" value={`${resources.cpu} vCPU`} />
            <SummaryItem label="Memory" value={`${resources.memory} GB`} />
            <SummaryItem label="Storage" value={`${resources.storage} GB`} />
            {resources.gpu > 0 && <SummaryItem label="GPU" value={`${resources.gpu} GPU`} />}
            <SummaryItem
              label="Duration"
              value={`${resources.duration} ${resources.durationUnit} (${totalHours}h)`}
            />
            {resources.region && <SummaryItem label="Region" value={resources.region} />}
          </div>
          <div className="mt-4 flex items-center justify-between rounded-lg bg-primary/5 p-3">
            <span className="font-medium">Escrow Deposit</span>
            <span className="text-lg font-bold text-primary">
              {formatTokenAmount(priceBreakdown.escrowDeposit)} {priceBreakdown.currency}
            </span>
          </div>
        </CardContent>
      </Card>

      {/* Actions */}
      <div className="flex flex-col gap-3 sm:flex-row">
        <Button asChild className="flex-1">
          <Link href={`/orders/${orderResult.orderId}`}>View Order Details</Link>
        </Button>
        <Button variant="outline" asChild className="flex-1">
          <Link href="/orders">Back to Orders</Link>
        </Button>
        <Button variant="outline" asChild className="flex-1">
          <Link href="/marketplace">Browse Marketplace</Link>
        </Button>
      </div>
    </div>
  );
}

function SummaryItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between rounded-md bg-muted/50 px-3 py-2 text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium">{value}</span>
    </div>
  );
}

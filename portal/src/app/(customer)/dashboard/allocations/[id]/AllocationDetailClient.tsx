/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, useState, useCallback } from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Progress } from '@/components/ui/Progress';
import { Skeleton } from '@/components/ui/Skeleton';
import { TerminateAllocationDialog } from '@/components/dashboard/TerminateAllocationDialog';
import { useCustomerDashboardStore, selectAllocationById } from '@/stores/customerDashboardStore';
import { formatCurrency, formatDate, truncateAddress } from '@/lib/utils';
import { CUSTOMER_ALLOCATION_STATUS_VARIANT } from '@/types/customer';
import type { CustomerAllocationStatus } from '@/types/customer';

function statusLabel(status: CustomerAllocationStatus): string {
  const labels: Record<CustomerAllocationStatus, string> = {
    pending: 'Pending',
    deploying: 'Deploying',
    running: 'Running',
    paused: 'Paused',
    stopped: 'Stopped',
    failed: 'Failed',
    terminated: 'Terminated',
  };
  return labels[status];
}

function isTerminable(status: CustomerAllocationStatus): boolean {
  return status === 'running' || status === 'paused' || status === 'deploying';
}

export default function AllocationDetailClient() {
  const params = useParams();
  const id = params.id as string;

  const { fetchDashboard, terminateAllocation, isLoading, allocations } =
    useCustomerDashboardStore();
  const allocation = useCustomerDashboardStore((state) => selectAllocationById(state, id));

  const [showTerminate, setShowTerminate] = useState(false);

  useEffect(() => {
    if (allocations.length === 0) {
      void fetchDashboard();
    }
  }, [allocations.length, fetchDashboard]);

  const handleTerminate = useCallback(
    async (allocationId: string) => {
      await terminateAllocation(allocationId);
    },
    [terminateAllocation]
  );

  if (isLoading && !allocation) {
    return <AllocationDetailSkeleton />;
  }

  if (!allocation) {
    return (
      <div className="space-y-6">
        <Link
          href="/dashboard"
          className="text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          ← Back to Dashboard
        </Link>
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <h2 className="text-lg font-medium">Allocation not found</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            The allocation you are looking for does not exist.
          </p>
        </div>
      </div>
    );
  }

  const estimatedMonthly = allocation.costPerHour * 730;

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex items-center justify-between">
        <Link
          href="/dashboard"
          className="text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          ← Back to Dashboard
        </Link>
      </div>

      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{allocation.offeringName}</h1>
            <Badge variant={CUSTOMER_ALLOCATION_STATUS_VARIANT[allocation.status]} dot size="lg">
              {statusLabel(allocation.status)}
            </Badge>
          </div>
          <p className="mt-1 text-sm text-muted-foreground">
            {allocation.providerName} · Created {formatDate(allocation.createdAt)}
          </p>
        </div>
        {isTerminable(allocation.status) && (
          <Button variant="destructive" size="sm" onClick={() => setShowTerminate(true)}>
            Terminate
          </Button>
        )}
      </div>

      {/* Main grid */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Left column */}
        <div className="space-y-6 lg:col-span-2">
          {/* Resources */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Provisioned Resources</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <ResourceStat label="CPU" value={`${allocation.resources.cpu} vCPU`} />
                <ResourceStat label="Memory" value={`${allocation.resources.memory} GB`} />
                <ResourceStat label="Storage" value={`${allocation.resources.storage} GB`} />
                {allocation.resources.gpu !== null &&
                  allocation.resources.gpu !== undefined &&
                  allocation.resources.gpu > 0 && (
                    <ResourceStat label="GPU" value={`${allocation.resources.gpu} GPU`} />
                  )}
              </div>
            </CardContent>
          </Card>

          {/* Usage metrics */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Usage Metrics</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {allocation.resources.cpu > 0 && (
                <UsageRow
                  label="CPU"
                  used={Math.round(allocation.resources.cpu * 0.67)}
                  total={allocation.resources.cpu}
                  unit="cores"
                />
              )}
              {allocation.resources.memory > 0 && (
                <UsageRow
                  label="Memory"
                  used={Math.round(allocation.resources.memory * 0.62)}
                  total={allocation.resources.memory}
                  unit="GB"
                />
              )}
              {allocation.resources.storage > 0 && (
                <UsageRow
                  label="Storage"
                  used={Math.round(allocation.resources.storage * 0.45)}
                  total={allocation.resources.storage}
                  unit="GB"
                />
              )}
              {allocation.resources.gpu !== null &&
                allocation.resources.gpu !== undefined &&
                allocation.resources.gpu > 0 && (
                  <UsageRow
                    label="GPU"
                    used={Math.round(allocation.resources.gpu * 0.8)}
                    total={allocation.resources.gpu}
                    unit="units"
                  />
                )}
            </CardContent>
          </Card>

          {/* Allocation details */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Details</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 sm:grid-cols-2">
                <InfoItem label="Allocation ID" value={allocation.id} />
                <InfoItem label="Order ID" value={allocation.orderId} />
                <InfoItem label="Provider" value={allocation.providerName} />
                <InfoItem
                  label="Provider Address"
                  value={truncateAddress(allocation.providerAddress)}
                />
                <InfoItem label="Created" value={formatDate(allocation.createdAt)} />
                <InfoItem label="Last Updated" value={formatDate(allocation.updatedAt)} />
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Right column */}
        <div className="space-y-6">
          {/* Cost summary */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Cost Summary</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Hourly Rate</span>
                <span>{formatCurrency(allocation.costPerHour)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Total Spent</span>
                <span className="font-medium">{formatCurrency(allocation.totalSpent)}</span>
              </div>
              {allocation.status === 'running' && (
                <div className="border-t border-border pt-3">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Projected Monthly</span>
                    <span>{formatCurrency(estimatedMonthly)}</span>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Quick actions */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Actions</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-2">
              <Button variant="outline" size="sm" asChild>
                <Link href={`/orders/${allocation.orderId}`}>View Order</Link>
              </Button>
              <Button variant="outline" size="sm" asChild>
                <Link href="/support">Contact Support</Link>
              </Button>
              {isTerminable(allocation.status) && (
                <Button variant="destructive" size="sm" onClick={() => setShowTerminate(true)}>
                  Terminate Allocation
                </Button>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Terminate dialog */}
      <TerminateAllocationDialog
        allocation={allocation}
        open={showTerminate}
        onOpenChange={setShowTerminate}
        onConfirm={handleTerminate}
      />
    </div>
  );
}

// =============================================================================
// Helper Components
// =============================================================================

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-sm text-muted-foreground">{label}</dt>
      <dd className="mt-1 font-medium">{value}</dd>
    </div>
  );
}

function ResourceStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg bg-muted/50 p-4 text-center">
      <div className="text-xl font-bold">{value}</div>
      <div className="mt-1 text-sm text-muted-foreground">{label}</div>
    </div>
  );
}

function utilizationVariant(pct: number): 'default' | 'success' | 'warning' | 'destructive' {
  if (pct >= 90) return 'destructive';
  if (pct >= 75) return 'warning';
  if (pct >= 40) return 'success';
  return 'default';
}

function UsageRow({
  label,
  used,
  total,
  unit,
}: {
  label: string;
  used: number;
  total: number;
  unit: string;
}) {
  const pct = total > 0 ? Math.round((used / total) * 100) : 0;
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-sm">
        <span className="font-medium">{label}</span>
        <span className="text-muted-foreground">
          {used} / {total} {unit}
        </span>
      </div>
      <Progress value={pct} size="sm" variant={utilizationVariant(pct)} />
    </div>
  );
}

function AllocationDetailSkeleton() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-5 w-32" />
      <div className="space-y-2">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-5 w-96" />
      </div>
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <Skeleton className="h-48 rounded-lg" />
          <Skeleton className="h-64 rounded-lg" />
        </div>
        <div className="space-y-6">
          <Skeleton className="h-40 rounded-lg" />
          <Skeleton className="h-48 rounded-lg" />
        </div>
      </div>
    </div>
  );
}

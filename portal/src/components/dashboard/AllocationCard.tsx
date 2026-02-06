/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import Link from 'next/link';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { formatCurrency, formatRelativeTime } from '@/lib/utils';
import type { CustomerAllocation, CustomerAllocationStatus } from '@/types/customer';
import { CUSTOMER_ALLOCATION_STATUS_VARIANT } from '@/types/customer';

interface AllocationCardProps {
  allocation: CustomerAllocation;
}

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

function ResourceChip({ label, value, unit }: { label: string; value: number; unit: string }) {
  if (value === 0) return null;
  return (
    <span className="inline-flex items-center gap-1 rounded bg-muted px-2 py-0.5 text-xs text-muted-foreground">
      {label}: {value} {unit}
    </span>
  );
}

export function AllocationCard({ allocation }: AllocationCardProps) {
  return (
    <Link href={`/orders/${allocation.orderId}`} className="block">
      <Card className="transition-all hover:shadow-md">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">{allocation.offeringName}</CardTitle>
          <Badge variant={CUSTOMER_ALLOCATION_STATUS_VARIANT[allocation.status]} dot size="sm">
            {statusLabel(allocation.status)}
          </Badge>
        </CardHeader>
        <CardContent>
          <div className="mb-3 flex items-center justify-between">
            <span className="text-xs text-muted-foreground">{allocation.providerName}</span>
            <span className="text-sm font-semibold">
              {formatCurrency(allocation.costPerHour)}/hr
            </span>
          </div>
          <div className="mb-3 flex flex-wrap gap-1.5">
            <ResourceChip label="CPU" value={allocation.resources.cpu} unit="cores" />
            <ResourceChip label="Mem" value={allocation.resources.memory} unit="GB" />
            <ResourceChip label="Disk" value={allocation.resources.storage} unit="GB" />
            {allocation.resources.gpu && (
              <ResourceChip label="GPU" value={allocation.resources.gpu} unit="" />
            )}
          </div>
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <span>Total: {formatCurrency(allocation.totalSpent)}</span>
            <span>{formatRelativeTime(allocation.updatedAt)}</span>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

interface AllocationListProps {
  allocations: CustomerAllocation[];
}

export function AllocationList({ allocations }: AllocationListProps) {
  if (allocations.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12 text-center">
        <p className="text-sm text-muted-foreground">No allocations match the current filter.</p>
      </div>
    );
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {allocations.map((alloc) => (
        <AllocationCard key={alloc.id} allocation={alloc} />
      ))}
    </div>
  );
}

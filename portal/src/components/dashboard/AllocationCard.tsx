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
import { useTranslation } from 'react-i18next';

interface AllocationCardProps {
  allocation: CustomerAllocation;
}

function statusLabel(status: CustomerAllocationStatus, t: (key: string) => string): string {
  const labels: Record<CustomerAllocationStatus, string> = {
    pending: t('Pending'),
    deploying: t('Deploying'),
    running: t('Running'),
    paused: t('Paused'),
    stopped: t('Stopped'),
    failed: t('Failed'),
    terminated: t('Terminated'),
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
  const { t } = useTranslation();

  return (
    <Link href={`/dashboard/allocations/${allocation.id}`} className="block">
      <Card className="transition-all hover:shadow-md">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">{allocation.offeringName}</CardTitle>
          <Badge variant={CUSTOMER_ALLOCATION_STATUS_VARIANT[allocation.status]} dot size="sm">
            {statusLabel(allocation.status, t)}
          </Badge>
        </CardHeader>
        <CardContent>
          <div className="mb-3 flex items-center justify-between">
            <span className="text-xs text-muted-foreground">{allocation.providerName}</span>
            <span className="text-sm font-semibold">
              {t('{{amount}}/hr', { amount: formatCurrency(allocation.costPerHour) })}
            </span>
          </div>
          <div className="mb-3 flex flex-wrap gap-1.5">
            <ResourceChip label={t('CPU')} value={allocation.resources.cpu} unit={t('cores')} />
            <ResourceChip label={t('Mem')} value={allocation.resources.memory} unit={t('GB')} />
            <ResourceChip label={t('Disk')} value={allocation.resources.storage} unit={t('GB')} />
            {allocation.resources.gpu && (
              <ResourceChip label={t('GPU')} value={allocation.resources.gpu} unit="" />
            )}
          </div>
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <span>{t('Total: {{amount}}', { amount: formatCurrency(allocation.totalSpent) })}</span>
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
  const { t } = useTranslation();

  if (allocations.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12 text-center">
        <p className="text-sm text-muted-foreground">
          {t('No allocations match the current filter.')}
        </p>
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

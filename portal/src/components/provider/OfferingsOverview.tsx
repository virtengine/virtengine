/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import Link from 'next/link';
import { useProviderStore } from '@/stores/providerStore';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { formatCurrency, formatRelativeTime } from '@/lib/utils';
import { getCategoryIcon } from '@/hooks/useOfferingSync';
import type { OfferingSyncStatus, ProviderOfferingSummary } from '@/types/provider';

const STATUS_VARIANT: Record<
  ProviderOfferingSummary['status'],
  'default' | 'success' | 'warning' | 'destructive' | 'secondary' | 'outline'
> = {
  pending: 'warning',
  published: 'success',
  failed: 'destructive',
  paused: 'secondary',
  deprecated: 'outline',
  draft: 'default',
};

const SYNC_VARIANT: Record<OfferingSyncStatus, 'default' | 'success' | 'warning' | 'destructive'> =
  {
    synced: 'success',
    syncing: 'warning',
    pending: 'default',
    failed: 'destructive',
  };

const SYNC_LABEL: Record<OfferingSyncStatus, string> = {
  synced: 'Synced',
  syncing: 'Syncing',
  pending: 'Pending',
  failed: 'Failed',
};

function OfferingRow({ offering }: { offering: ProviderOfferingSummary }) {
  return (
    <div className="flex flex-col gap-3 rounded-lg border border-border p-4 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex items-start gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-muted text-lg">
          {getCategoryIcon(offering.category)}
        </div>
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-sm font-semibold">{offering.name}</span>
            <Badge variant={STATUS_VARIANT[offering.status]} size="sm">
              {offering.status}
            </Badge>
            <Badge variant={SYNC_VARIANT[offering.syncStatus]} size="sm" dot>
              {SYNC_LABEL[offering.syncStatus]}
            </Badge>
          </div>
          <div className="mt-1 text-xs text-muted-foreground">
            {offering.activeOrders} active · {offering.totalOrders} total orders · Updated{' '}
            {formatRelativeTime(offering.updatedAt)}
          </div>
        </div>
      </div>
      <div className="text-right text-sm">
        <div className="font-semibold">
          {formatCurrency(offering.basePrice, offering.currency)}
          <span className="text-xs text-muted-foreground"> / unit</span>
        </div>
        <div className="text-xs text-muted-foreground">
          Sync: {formatRelativeTime(offering.lastSyncedAt)}
        </div>
      </div>
    </div>
  );
}

export default function OfferingsOverview() {
  const offerings = useProviderStore((s) => s.offerings);

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <CardTitle className="text-lg">My Offerings</CardTitle>
          <div className="flex items-center gap-3 text-sm">
            <span className="text-muted-foreground">{offerings.length} total</span>
            <Link href="/provider/offerings" className="text-primary hover:underline">
              Manage offerings
            </Link>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {offerings.length === 0 ? (
          <div className="py-6 text-center text-sm text-muted-foreground">
            No offerings synced yet.
          </div>
        ) : (
          offerings.map((offering) => <OfferingRow key={offering.id} offering={offering} />)
        )}
      </CardContent>
    </Card>
  );
}

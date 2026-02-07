/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useProviderStore } from '@/stores/providerStore';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { formatRelativeTime } from '@/lib/utils';
import type { ProviderSyncStatus, SyncHealthStatus } from '@/types/provider';

const STATUS_VARIANT: Record<
  SyncHealthStatus,
  'default' | 'success' | 'warning' | 'destructive' | 'secondary'
> = {
  synced: 'success',
  syncing: 'warning',
  degraded: 'warning',
  failed: 'destructive',
  offline: 'secondary',
};

function EndpointCard({ status }: { status: ProviderSyncStatus['waldur'] }) {
  return (
    <div className="rounded-lg border border-border p-4">
      <div className="flex items-center justify-between">
        <div className="text-sm font-medium">{status.name}</div>
        <Badge variant={STATUS_VARIANT[status.status]} size="sm" dot>
          {status.status}
        </Badge>
      </div>
      <div className="mt-2 text-xs text-muted-foreground">
        Last success {formatRelativeTime(status.lastSuccessAt)}
      </div>
      <div className="mt-1 text-xs text-muted-foreground">Lag: {status.lagSeconds}s</div>
      {status.message && <div className="mt-2 text-xs">{status.message}</div>}
    </div>
  );
}

export default function ProviderSyncStatus() {
  const syncStatus = useProviderStore((s) => s.syncStatus);

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <CardTitle className="text-lg">Waldur ↔ Chain Sync Status</CardTitle>
          <Badge variant={syncStatus.isRunning ? 'warning' : 'success'} size="sm">
            {syncStatus.isRunning ? 'Sync running' : 'Idle'}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div className="rounded-lg border border-border p-4">
            <div className="text-xs text-muted-foreground">Last sync</div>
            <div className="mt-1 text-sm font-semibold">
              {new Date(syncStatus.lastSyncAt).toLocaleString()}
            </div>
          </div>
          <div className="rounded-lg border border-border p-4">
            <div className="text-xs text-muted-foreground">Next sync</div>
            <div className="mt-1 text-sm font-semibold">
              {new Date(syncStatus.nextSyncAt).toLocaleString()}
            </div>
          </div>
          <div className="rounded-lg border border-border p-4">
            <div className="text-xs text-muted-foreground">Pending queue</div>
            <div className="mt-1 text-sm font-semibold">
              {syncStatus.pendingOfferings} offerings · {syncStatus.pendingAllocations} allocations
            </div>
          </div>
          <div className="rounded-lg border border-border p-4">
            <div className="text-xs text-muted-foreground">Errors</div>
            <div className="mt-1 text-sm font-semibold">{syncStatus.errorCount}</div>
            {syncStatus.lastError && (
              <div className="mt-1 text-xs text-muted-foreground">{syncStatus.lastError}</div>
            )}
          </div>
        </div>

        <div className="grid gap-4 lg:grid-cols-3">
          <EndpointCard status={syncStatus.waldur} />
          <EndpointCard status={syncStatus.chain} />
          <EndpointCard status={syncStatus.providerDaemon} />
        </div>
      </CardContent>
    </Card>
  );
}

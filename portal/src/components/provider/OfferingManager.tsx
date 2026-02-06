/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import Link from 'next/link';
import { useProviderStore } from '@/stores/providerStore';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { formatDate, truncateAddress } from '@/lib/utils';
import type { QueuedAllocation } from '@/types/provider';

function QueueItem({ item }: { item: QueuedAllocation }) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-border p-3">
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{item.offeringName}</span>
          <Badge variant="warning" size="sm">
            Queued
          </Badge>
        </div>
        <div className="mt-1 text-xs text-muted-foreground">
          Customer: {truncateAddress(item.customerAddress, 14, 4)}
        </div>
        <div className="mt-0.5 text-xs text-muted-foreground">
          {item.resources.cpu > 0 && <span>{item.resources.cpu} CPU · </span>}
          {item.resources.memory > 0 && <span>{item.resources.memory} GB RAM</span>}
          {item.resources.gpu && item.resources.gpu > 0 && <span> · {item.resources.gpu} GPU</span>}
        </div>
      </div>
      <div className="text-right">
        <div className="text-xs text-muted-foreground">ETA: {item.estimatedProvisionTime}</div>
        <div className="text-xs text-muted-foreground">{formatDate(item.requestedAt)}</div>
      </div>
    </div>
  );
}

export default function OfferingManager() {
  const stats = useProviderStore((s) => s.stats);
  const queue = useProviderStore((s) => s.queue);

  return (
    <div className="space-y-6">
      {/* Offering Stats */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="pt-6">
            <div className="text-sm text-muted-foreground">Active Allocations</div>
            <div className="mt-1 text-3xl font-bold">{stats.activeAllocations}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-sm text-muted-foreground">Published Offerings</div>
            <div className="mt-1 text-3xl font-bold">
              {stats.publishedOfferings}
              <span className="text-base font-normal text-muted-foreground">
                /{stats.totalOfferings}
              </span>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-sm text-muted-foreground">Uptime</div>
            <div className="mt-1 text-3xl font-bold">{stats.uptime}%</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-sm text-muted-foreground">Pending Orders</div>
            <div className="mt-1 text-3xl font-bold">{stats.pendingOrders}</div>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions + Queue */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Quick Actions */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Quick Actions</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-3 sm:grid-cols-2">
            <Link
              href="/provider/offerings"
              className="card-hover flex items-center gap-3 rounded-lg border border-border p-3 transition-all hover:bg-accent"
            >
              <span className="text-xl">📦</span>
              <div>
                <div className="text-sm font-medium">Manage Offerings</div>
                <div className="text-xs text-muted-foreground">Create, update, disable</div>
              </div>
            </Link>
            <Link
              href="/provider/pricing"
              className="card-hover flex items-center gap-3 rounded-lg border border-border p-3 transition-all hover:bg-accent"
            >
              <span className="text-xl">💰</span>
              <div>
                <div className="text-sm font-medium">Update Pricing</div>
                <div className="text-xs text-muted-foreground">Adjust rates and strategy</div>
              </div>
            </Link>
            <Link
              href="/provider/orders"
              className="card-hover flex items-center gap-3 rounded-lg border border-border p-3 transition-all hover:bg-accent"
            >
              <span className="text-xl">📋</span>
              <div>
                <div className="text-sm font-medium">View Orders</div>
                <div className="text-xs text-muted-foreground">Manage deployments</div>
              </div>
            </Link>
            <Link
              href="/provider/offerings"
              className="card-hover flex items-center gap-3 rounded-lg border border-border p-3 transition-all hover:bg-accent"
            >
              <span className="text-xl">⚙️</span>
              <div>
                <div className="text-sm font-medium">Configure Capacity</div>
                <div className="text-xs text-muted-foreground">Set resource limits</div>
              </div>
            </Link>
          </CardContent>
        </Card>

        {/* Allocation Queue */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-lg">Allocation Queue</CardTitle>
              <Badge variant="secondary">{queue.length} pending</Badge>
            </div>
          </CardHeader>
          <CardContent>
            {queue.length === 0 ? (
              <div className="py-4 text-center text-sm text-muted-foreground">
                No pending provisions
              </div>
            ) : (
              <div className="space-y-3">
                {queue.map((item) => (
                  <QueueItem key={item.id} item={item} />
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useProviderStore } from '@/stores/providerStore';
import { useWallet } from '@/lib/portal-adapter';
import { formatDate, truncateAddress } from '@/lib/utils';
import { formatToken } from '@/components/escrow/utils';
import type { AllocationStatus } from '@/types/provider';

interface ProviderOrderRow {
  id: string;
  customer: string;
  offering: string;
  status: AllocationStatus;
  updatedAt: string;
  revenue: number;
  currency: string;
}

function getStatusColor(status: AllocationStatus) {
  switch (status) {
    case 'ok':
      return 'bg-green-500/10 text-green-600 dark:text-green-400';
    case 'pending':
    case 'creating':
    case 'updating':
      return 'bg-yellow-500/10 text-yellow-600 dark:text-yellow-400';
    case 'terminated':
      return 'bg-blue-500/10 text-blue-600 dark:text-blue-400';
    case 'erred':
    case 'terminating':
      return 'bg-red-500/10 text-red-600 dark:text-red-400';
    default:
      return 'bg-gray-500/10 text-gray-600 dark:text-gray-400';
  }
}

function toTitleCase(value: string) {
  return value.replace(/_/g, ' ').replace(/\b\w/g, (char) => char.toUpperCase());
}

export default function ProviderOrdersPage() {
  const [filter, setFilter] = useState<AllocationStatus | 'all'>('all');
  const wallet = useWallet();
  const account = wallet.accounts[wallet.activeAccountIndex];
  const allocations = useProviderStore((s) => s.allocations);
  const isLoading = useProviderStore((s) => s.isLoading);
  const error = useProviderStore((s) => s.error);
  const fetchDashboard = useProviderStore((s) => s.fetchDashboard);

  useEffect(() => {
    if (!account?.address) return;
    void fetchDashboard(account.address);
  }, [account?.address, fetchDashboard]);

  const orders = useMemo<ProviderOrderRow[]>(
    () =>
      allocations.map((allocation) => ({
        id: allocation.id,
        customer: allocation.customerAddress,
        offering: allocation.offeringName,
        status: allocation.status,
        updatedAt: allocation.updatedAt,
        revenue: allocation.monthlyRevenue,
        currency: 'UVE',
      })),
    [allocations]
  );

  const filteredOrders =
    filter === 'all' ? orders : orders.filter((order) => order.status === filter);

  if (!account?.address) {
    return (
      <div className="container py-8">
        <div className="rounded-lg border border-border bg-card p-6">
          <h1 className="text-2xl font-bold">Provider Orders</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Connect the provider wallet to load live allocations and deployment state.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="container py-8">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Provider Orders</h1>
          <p className="mt-1 text-muted-foreground">
            Manage live customer deployments and provider-side allocation state.
          </p>
        </div>
      </div>

      {error && (
        <div className="mb-6 rounded-lg border border-destructive bg-destructive/10 p-4">
          <p className="font-medium text-destructive">Failed to load provider orders</p>
          <p className="mt-1 text-sm text-muted-foreground">{error}</p>
        </div>
      )}

      <div className="mb-6 flex gap-2">
        {(
          ['all', 'ok', 'pending', 'creating', 'updating', 'terminated', 'erred', 'terminating'] as const
        ).map((status) => (
          <button
            key={status}
            onClick={() => setFilter(status)}
            className={`rounded-full px-4 py-2 text-sm font-medium transition-colors ${
              filter === status
                ? 'bg-primary text-primary-foreground'
                : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
            }`}
          >
            {status === 'all' ? 'All' : toTitleCase(status)}
          </button>
        ))}
      </div>

      <div className="rounded-lg border border-border">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="border-b border-border bg-muted/50">
              <tr>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">
                  Deployment ID
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">
                  Customer
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">
                  Offering
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">
                  Status
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">
                  Updated
                </th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted-foreground">
                  Accrued
                </th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody>
              {filteredOrders.map((order) => (
                <tr key={order.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-4">
                    <span className="font-mono text-sm">{order.id}</span>
                  </td>
                  <td className="px-4 py-4">
                    <span className="text-sm text-muted-foreground">
                      {truncateAddress(order.customer, 14, 4)}
                    </span>
                  </td>
                  <td className="px-4 py-4">
                    <span className="text-sm">{order.offering}</span>
                  </td>
                  <td className="px-4 py-4">
                    <span
                      className={`rounded-full px-2 py-1 text-xs font-medium ${getStatusColor(order.status)}`}
                    >
                      {toTitleCase(order.status)}
                    </span>
                  </td>
                  <td className="px-4 py-4">
                    <span className="text-sm text-muted-foreground">
                      {formatDate(order.updatedAt)}
                    </span>
                  </td>
                  <td className="px-4 py-4 text-right">
                    <span className="font-medium">{formatToken(order.revenue, order.currency)}</span>
                  </td>
                  <td className="px-4 py-4">
                    <Link
                      href={`/provider/orders/${order.id}` as '/provider/orders/[id]'}
                      className="text-sm text-primary hover:underline"
                    >
                      View
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {isLoading && filteredOrders.length === 0 && (
        <div className="py-12 text-center">
          <p className="text-muted-foreground">Loading live provider allocations…</p>
        </div>
      )}

      {!isLoading && filteredOrders.length === 0 && (
        <div className="py-12 text-center">
          <p className="text-muted-foreground">No live allocations match the current filter.</p>
        </div>
      )}
    </div>
  );
}

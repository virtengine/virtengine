/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * OrderList component with tab-based filtering, search, and sort.
 */

'use client';

import { useEffect, useMemo, useState, useCallback } from 'react';
import Link from 'next/link';
import { useOrderStore, selectFilteredOrders, type OrderStatus } from '@/stores/orderStore';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Skeleton } from '@/components/ui/Skeleton';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/Tabs';
import {
  type OrderTabFilter,
  ORDER_TAB_FILTERS,
  STATUS_TO_TAB,
  ORDER_STATUS_CONFIG,
  isOrderActive,
  formatDuration,
} from '@/features/orders/tracking-types';
import { formatCurrency, formatRelativeTime } from '@/lib/utils';
import { useWallet } from '@/lib/portal-adapter';

export function OrderList() {
  const { fetchOrders, isLoading, error, setFilter } = useOrderStore();
  const filteredOrders = useOrderStore(selectFilteredOrders);
  const [activeTab, setActiveTab] = useState<OrderTabFilter>('active');
  const [searchQuery, setSearchQuery] = useState('');
  const wallet = useWallet();
  const account = wallet.accounts[wallet.activeAccountIndex];

  useEffect(() => {
    if (!account?.address) return;
    void fetchOrders(account.address);
    const interval = setInterval(() => {
      void fetchOrders(account.address);
    }, 30000);
    return () => clearInterval(interval);
  }, [account?.address, fetchOrders]);

  const handleTabChange = useCallback(
    (value: string) => {
      const tab = value as OrderTabFilter;
      setActiveTab(tab);
      if (tab === 'all') {
        setFilter({ status: 'all' });
      } else {
        // The store filter works on individual statuses, so we filter in the component
        setFilter({ status: 'all' });
      }
    },
    [setFilter]
  );

  const displayedOrders = useMemo(() => {
    let orders = filteredOrders;

    // Tab-based filtering
    if (activeTab !== 'all') {
      orders = orders.filter((o) => STATUS_TO_TAB[o.status] === activeTab);
    }

    // Search
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      orders = orders.filter(
        (o) =>
          o.id.toLowerCase().includes(q) ||
          o.providerName.toLowerCase().includes(q) ||
          o.resourceType.toLowerCase().includes(q)
      );
    }

    return orders;
  }, [filteredOrders, activeTab, searchQuery]);

  const tabCounts = useMemo(() => {
    const counts: Record<OrderTabFilter, number> = {
      active: 0,
      pending: 0,
      completed: 0,
      all: filteredOrders.length,
    };
    for (const order of filteredOrders) {
      const tab = STATUS_TO_TAB[order.status];
      counts[tab]++;
    }
    return counts;
  }, [filteredOrders]);

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <div className="rounded-full bg-destructive/10 p-4">
          <span className="text-4xl" role="img" aria-label="Error">
            ⚠
          </span>
        </div>
        <h2 className="mt-4 text-lg font-medium">Failed to load orders</h2>
        <p className="mt-2 text-sm text-muted-foreground">{error}</p>
        <Button
          onClick={() => account?.address && fetchOrders(account.address)}
          className="mt-4"
          variant="outline"
        >
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold">Orders</h1>
          <p className="mt-1 text-muted-foreground">Manage your orders and deployments</p>
        </div>
        <Button asChild>
          <Link href="/marketplace">+ New Order</Link>
        </Button>
      </div>

      {/* Search */}
      <div className="flex gap-4">
        <Input
          placeholder="Search orders by ID, provider, or type..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="max-w-md"
        />
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={handleTabChange}>
        <TabsList>
          {ORDER_TAB_FILTERS.map((tab) => (
            <TabsTrigger key={tab.value} value={tab.value}>
              {tab.label}
              {tabCounts[tab.value] > 0 && (
                <Badge variant="secondary" size="sm" className="ml-2">
                  {tabCounts[tab.value]}
                </Badge>
              )}
            </TabsTrigger>
          ))}
        </TabsList>

        {ORDER_TAB_FILTERS.map((tab) => (
          <TabsContent key={tab.value} value={tab.value}>
            {isLoading ? (
              <OrderListSkeleton />
            ) : displayedOrders.length === 0 ? (
              <EmptyState tab={tab.value} searchQuery={searchQuery} />
            ) : (
              <div className="space-y-3">
                {displayedOrders.map((order) => (
                  <OrderCard
                    key={order.id}
                    id={order.id}
                    providerName={order.providerName}
                    resourceType={order.resourceType}
                    status={order.status}
                    createdAt={order.createdAt}
                    hourlyRate={order.cost.hourlyRate}
                    totalCost={order.cost.totalCost}
                    currency={order.cost.currency}
                    resources={order.resources}
                  />
                ))}
              </div>
            )}
          </TabsContent>
        ))}
      </Tabs>
    </div>
  );
}

// =============================================================================
// Order Card
// =============================================================================

interface OrderCardProps {
  id: string;
  providerName: string;
  resourceType: string;
  status: OrderStatus;
  createdAt: Date;
  hourlyRate: number;
  totalCost: number;
  currency: string;
  resources: {
    cpu: number;
    memory: number;
    storage: number;
    gpu?: number;
  };
}

function OrderCard({
  id,
  providerName,
  resourceType,
  status,
  createdAt,
  hourlyRate,
  totalCost,
  currency,
  resources,
}: OrderCardProps) {
  const config = ORDER_STATUS_CONFIG[status];
  const active = isOrderActive(status);

  return (
    <Link
      href={`/orders/${id}`}
      className="group flex flex-col gap-4 rounded-lg border border-border bg-card p-4 transition-all hover:border-primary/30 hover:shadow-md sm:flex-row sm:items-center sm:justify-between"
    >
      <div className="flex items-center gap-4">
        <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-muted">
          <span className="text-lg font-bold text-muted-foreground">
            {resourceType.charAt(0).toUpperCase()}
          </span>
        </div>
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="truncate font-medium">Order #{id}</h3>
            <Badge variant={config.variant} dot size="sm">
              {config.label}
            </Badge>
          </div>
          <p className="mt-0.5 text-sm text-muted-foreground">
            {resourceType} · {providerName} · {formatRelativeTime(createdAt)}
          </p>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-4 text-sm sm:gap-6">
        {/* Resource summary */}
        <div className="hidden text-muted-foreground lg:flex lg:gap-3">
          <span>{resources.cpu} vCPU</span>
          <span>·</span>
          <span>{resources.memory} GB</span>
          {resources.gpu && resources.gpu > 0 && (
            <>
              <span>·</span>
              <span>{resources.gpu} GPU</span>
            </>
          )}
        </div>

        {/* Cost */}
        <div className="text-right">
          <div className="font-medium">{formatCurrency(totalCost, currency)}</div>
          {active && (
            <div className="text-xs text-muted-foreground">
              {formatCurrency(hourlyRate, currency)}/hr
            </div>
          )}
        </div>

        {/* Duration */}
        <div className="hidden text-muted-foreground sm:block">
          {formatDuration(createdAt.toISOString())}
        </div>

        {/* Arrow */}
        <span className="text-muted-foreground transition-transform group-hover:translate-x-1">
          →
        </span>
      </div>
    </Link>
  );
}

// =============================================================================
// Empty State
// =============================================================================

function EmptyState({ tab, searchQuery }: { tab: OrderTabFilter; searchQuery: string }) {
  if (searchQuery.trim()) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <h2 className="text-lg font-medium">No matching orders</h2>
        <p className="mt-2 text-sm text-muted-foreground">
          No orders match &quot;{searchQuery}&quot;. Try a different search.
        </p>
      </div>
    );
  }

  const messages: Record<OrderTabFilter, { title: string; description: string }> = {
    active: {
      title: 'No active orders',
      description: 'Your running and deploying orders will appear here',
    },
    pending: {
      title: 'No pending orders',
      description: 'Orders waiting to be matched with providers will appear here',
    },
    completed: {
      title: 'No completed orders',
      description: 'Your finished orders will appear here',
    },
    all: {
      title: 'No orders yet',
      description: 'Start by browsing the marketplace for compute resources',
    },
  };

  const msg = messages[tab];

  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <h2 className="text-lg font-medium">{msg.title}</h2>
      <p className="mt-2 text-sm text-muted-foreground">{msg.description}</p>
      {tab === 'all' && (
        <Button asChild className="mt-4">
          <Link href="/marketplace">Browse Marketplace</Link>
        </Button>
      )}
    </div>
  );
}

// =============================================================================
// Loading Skeleton
// =============================================================================

function OrderListSkeleton() {
  const keys = ['skel-1', 'skel-2', 'skel-3', 'skel-4'];
  return (
    <div className="space-y-3">
      {keys.map((key) => (
        <div key={key} className="flex items-center gap-4 rounded-lg border border-border p-4">
          <Skeleton className="h-12 w-12 rounded-lg" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-5 w-48" />
            <Skeleton className="h-4 w-72" />
          </div>
          <Skeleton className="h-6 w-20" />
        </div>
      ))}
    </div>
  );
}

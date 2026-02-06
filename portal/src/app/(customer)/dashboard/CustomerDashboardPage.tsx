/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import { SkeletonCard } from '@/components/ui/Skeleton';
import {
  AllocationList,
  BillingSummary,
  NotificationsFeed,
  QuickActions,
  UsageSummary,
} from '@/components/dashboard';
import {
  useCustomerDashboardStore,
  selectFilteredCustomerAllocations,
  selectUnreadNotificationCount,
} from '@/stores/customerDashboardStore';
import { formatCurrency } from '@/lib/utils';
import type { CustomerAllocationStatus } from '@/types/customer';

const ALLOCATION_TABS: { label: string; value: CustomerAllocationStatus | 'all' }[] = [
  { label: 'All', value: 'all' },
  { label: 'Running', value: 'running' },
  { label: 'Deploying', value: 'deploying' },
  { label: 'Failed', value: 'failed' },
  { label: 'Terminated', value: 'terminated' },
];

function StatCard({
  title,
  value,
  subtitle,
  change,
}: {
  title: string;
  value: string;
  subtitle?: string;
  change?: number;
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        {change !== undefined && (
          <Badge variant={change >= 0 ? 'destructive' : 'success'} size="sm">
            {change >= 0 ? '+' : ''}
            {change.toFixed(1)}%
          </Badge>
        )}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {subtitle && <p className="text-xs text-muted-foreground">{subtitle}</p>}
      </CardContent>
    </Card>
  );
}

export function CustomerDashboardPage() {
  const {
    stats,
    usage,
    billing,
    notifications,
    isLoading,
    error,
    allocationFilter,
    fetchDashboard,
    setAllocationFilter,
    markNotificationRead,
    dismissNotification,
  } = useCustomerDashboardStore();

  const filteredAllocations = useCustomerDashboardStore(selectFilteredCustomerAllocations);
  const unreadCount = useCustomerDashboardStore(selectUnreadNotificationCount);

  useEffect(() => {
    void fetchDashboard();
  }, [fetchDashboard]);

  if (error) {
    return (
      <div className="rounded-lg border border-destructive bg-destructive/10 p-4">
        <p className="font-medium text-destructive">Error loading dashboard</p>
        <p className="text-sm text-muted-foreground">{error}</p>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold">Dashboard</h1>
          <p className="mt-1 text-muted-foreground">Loading your dashboard…</p>
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }, (_, i) => `skel-${i}`).map((key) => (
            <SkeletonCard key={key} />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div>
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <p className="mt-1 text-muted-foreground">
          Overview of your allocations, usage, and billing
        </p>
      </div>

      {/* Summary stat cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Active Allocations"
          value={String(stats.activeAllocations)}
          subtitle={`${stats.totalOrders} total orders`}
        />
        <StatCard
          title="Monthly Spend"
          value={formatCurrency(stats.monthlySpend)}
          change={stats.spendChange}
        />
        <StatCard
          title="Pending Orders"
          value={String(stats.pendingOrders)}
          subtitle="Awaiting deployment"
        />
        <StatCard
          title="Notifications"
          value={String(unreadCount)}
          subtitle={`${notifications.length} total`}
        />
      </div>

      {/* Main content grid */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Left column: allocations + usage */}
        <div className="space-y-6 lg:col-span-2">
          {/* Allocations */}
          <div>
            <h2 className="mb-4 text-xl font-semibold">Allocations</h2>
            <Tabs
              value={allocationFilter}
              onValueChange={(v) => setAllocationFilter(v as CustomerAllocationStatus | 'all')}
            >
              <TabsList>
                {ALLOCATION_TABS.map((tab) => (
                  <TabsTrigger key={tab.value} value={tab.value}>
                    {tab.label}
                  </TabsTrigger>
                ))}
              </TabsList>
              {ALLOCATION_TABS.map((tab) => (
                <TabsContent key={tab.value} value={tab.value}>
                  <AllocationList allocations={filteredAllocations} />
                </TabsContent>
              ))}
            </Tabs>
          </div>

          {/* Usage */}
          <UsageSummary usage={usage} />
        </div>

        {/* Right column: billing, quick actions, notifications */}
        <div className="space-y-6">
          <BillingSummary billing={billing} />
          <QuickActions />
          <NotificationsFeed
            notifications={notifications}
            onMarkRead={markNotificationRead}
            onDismiss={dismissNotification}
          />
        </div>
      </div>
    </div>
  );
}

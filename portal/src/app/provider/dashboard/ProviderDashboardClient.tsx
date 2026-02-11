/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect } from 'react';
import { useProviderStore } from '@/stores/providerStore';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import OfferingManager from '@/components/provider/OfferingManager';
import OfferingsOverview from '@/components/provider/OfferingsOverview';
import RevenueSummary from '@/components/provider/RevenueSummary';
import CapacityView from '@/components/provider/CapacityView';
import AllocationTable from '@/components/provider/AllocationTable';
import PendingBidsTable from '@/components/provider/PendingBidsTable';
import PayoutHistory from '@/components/provider/PayoutHistory';
import ProviderTickets from '@/components/provider/ProviderTickets';
import ProviderSyncStatus from '@/components/provider/ProviderSyncStatus';
import { formatCurrency } from '@/lib/utils';
import { useWallet } from '@/lib/portal-adapter';

function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="animate-pulse rounded-lg border border-border bg-card p-6">
            <div className="h-4 w-24 rounded bg-muted" />
            <div className="mt-3 h-8 w-16 rounded bg-muted" />
          </div>
        ))}
      </div>
      <div className="animate-pulse rounded-lg border border-border bg-card p-6">
        <div className="h-6 w-40 rounded bg-muted" />
        <div className="mt-4 space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-12 rounded bg-muted" />
          ))}
        </div>
      </div>
    </div>
  );
}

export default function ProviderDashboardClient() {
  const isLoading = useProviderStore((s) => s.isLoading);
  const error = useProviderStore((s) => s.error);
  const stats = useProviderStore((s) => s.stats);
  const fetchDashboard = useProviderStore((s) => s.fetchDashboard);
  const clearError = useProviderStore((s) => s.clearError);
  const wallet = useWallet();
  const account = wallet.accounts[wallet.activeAccountIndex];

  useEffect(() => {
    if (!account?.address) return;
    void fetchDashboard(account.address);
    const interval = setInterval(() => {
      void fetchDashboard(account.address);
    }, 30000);
    return () => clearInterval(interval);
  }, [account?.address, fetchDashboard]);

  if (error) {
    return (
      <div className="rounded-lg border border-destructive bg-destructive/10 p-4">
        <h2 className="font-semibold text-destructive">Error loading dashboard</h2>
        <p className="mt-1 text-sm">{error}</p>
        <button
          type="button"
          onClick={() => {
            clearError();
            if (account?.address) {
              void fetchDashboard(account.address);
            }
          }}
          className="mt-4 rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground"
        >
          Retry
        </button>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div>
        <div className="mb-8">
          <h1 className="text-3xl font-bold">Provider Dashboard</h1>
          <p className="mt-1 text-muted-foreground">
            Manage your infrastructure and monitor performance
          </p>
        </div>
        <DashboardSkeleton />
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Provider Dashboard</h1>
          <p className="mt-1 text-muted-foreground">
            Manage your infrastructure and monitor performance
          </p>
        </div>
        <div className="text-right">
          <div className="text-sm text-muted-foreground">Monthly Revenue</div>
          <div className="text-2xl font-bold">{formatCurrency(stats.monthlyRevenue)}</div>
          <div
            className={`text-sm ${stats.revenueChange >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}
          >
            {stats.revenueChange >= 0 ? '↑' : '↓'} {Math.abs(stats.revenueChange).toFixed(1)}% vs
            last month
          </div>
        </div>
      </div>

      <Tabs defaultValue="overview" className="space-y-6">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="offerings">Offerings</TabsTrigger>
          <TabsTrigger value="allocations">Allocations</TabsTrigger>
          <TabsTrigger value="bids">Pending Bids</TabsTrigger>
          <TabsTrigger value="revenue">Revenue</TabsTrigger>
          <TabsTrigger value="sync">Sync Status</TabsTrigger>
          <TabsTrigger value="payouts">Payouts</TabsTrigger>
          <TabsTrigger value="support">Support</TabsTrigger>
        </TabsList>

        <TabsContent value="overview">
          <div className="space-y-6">
            <OfferingsOverview />
            <OfferingManager />
            <div className="grid gap-6 lg:grid-cols-2">
              <CapacityView />
              <ProviderTickets />
            </div>
          </div>
        </TabsContent>

        <TabsContent value="offerings">
          <OfferingsOverview />
        </TabsContent>

        <TabsContent value="allocations">
          <AllocationTable />
        </TabsContent>

        <TabsContent value="bids">
          <PendingBidsTable />
        </TabsContent>

        <TabsContent value="revenue">
          <RevenueSummary />
        </TabsContent>

        <TabsContent value="sync">
          <ProviderSyncStatus />
        </TabsContent>

        <TabsContent value="payouts">
          <PayoutHistory />
        </TabsContent>

        <TabsContent value="support">
          <ProviderTickets />
        </TabsContent>
      </Tabs>
    </div>
  );
}

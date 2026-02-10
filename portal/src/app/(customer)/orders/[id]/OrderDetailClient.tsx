'use client';

import Link from 'next/link';
import { useParams } from 'next/navigation';
import { useCallback, type ReactNode } from 'react';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/Tabs';
import {
  OrderStatusHeader,
  OrderStatusTimelineView,
  ConnectionStatusIndicator,
} from '@/components/orders/OrderStatusTracker';
import { ResourceAccess } from '@/components/orders/ResourceAccess';
import { UsageMonitor } from '@/components/orders/UsageMonitor';
import { OrderActions } from '@/components/orders/OrderActions';
import { useOrderTracking } from '@/features/orders/useOrderTracking';
import {
  ORDER_STATUS_CONFIG,
  formatDuration,
  isOrderActive,
} from '@/features/orders/tracking-types';
import type {
  ExtendOrderRequest,
  CancelOrderRequest,
  SupportTicketRequest,
  OrderActionResult,
} from '@/features/orders/tracking-types';
import { formatCurrency, formatDate, truncateAddress } from '@/lib/utils';
import { accountLink, txLink } from '@/lib/explorer';
import { useOrderStore } from '@/stores/orderStore';
import { useWallet } from '@/lib/portal-adapter';

export default function OrderDetailClient() {
  const params = useParams();
  const id = params.id as string;

  const { order, connectionStatus, isLoading, error, refresh } = useOrderTracking({
    orderId: id,
    enabled: true,
    pollingInterval: 30000,
  });
  const wallet = useWallet();
  const closeOrder = useOrderStore((s) => s.closeOrder);

  const handleExtend = useCallback(async (req: ExtendOrderRequest): Promise<OrderActionResult> => {
    // In production: apiClient.post('/orders/extend', req)
    await new Promise((resolve) => setTimeout(resolve, 800));
    return {
      success: true,
      message: `Order extended by ${req.additionalDuration} ${req.durationUnit}`,
    };
  }, []);

  const handleCancel = useCallback(
    async (req: CancelOrderRequest): Promise<OrderActionResult> => {
      if (wallet.status !== 'connected') {
        return { success: false, message: 'Connect your wallet to cancel this order.' };
      }
      const account = wallet.accounts[wallet.activeAccountIndex];
      if (!account?.address) {
        return { success: false, message: 'No active wallet account.' };
      }
      try {
        await closeOrder(
          req.orderId,
          account.address,
          wallet as unknown as Parameters<typeof closeOrder>[2]
        );
        return { success: true, message: 'Order cancellation submitted.' };
      } catch (error) {
        return {
          success: false,
          message: error instanceof Error ? error.message : 'Failed to cancel order.',
        };
      }
    },
    [closeOrder, wallet]
  );

  const handleSupport = useCallback(
    async (_req: SupportTicketRequest): Promise<OrderActionResult> => {
      await new Promise((resolve) => setTimeout(resolve, 800));
      return { success: true, message: 'Support ticket created', ticketId: 'TKT-1234' };
    },
    []
  );

  if (isLoading && !order) {
    return <OrderDetailSkeleton />;
  }

  if (error && !order) {
    return (
      <div className="container py-8">
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <h2 className="text-lg font-medium">Failed to load order</h2>
          <p className="mt-2 text-sm text-muted-foreground">{error}</p>
          <Button onClick={refresh} className="mt-4" variant="outline">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  if (!order) return null;

  const config = ORDER_STATUS_CONFIG[order.status];
  const active = isOrderActive(order.status);

  return (
    <div className="container py-8">
      {/* Breadcrumb */}
      <div className="mb-6 flex items-center justify-between">
        <Link
          href="/orders"
          className="text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          ← Back to Orders
        </Link>
        <ConnectionStatusIndicator status={connectionStatus} />
      </div>

      {/* Header */}
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">Order #{order.id}</h1>
            <Badge variant={config.variant} dot size="lg">
              {config.label}
            </Badge>
          </div>
          <p className="mt-1 text-sm text-muted-foreground">
            {order.offeringName} · {order.providerName} · Created {formatDate(order.createdAt)}
          </p>
        </div>
        {active && (
          <Button variant="outline" onClick={refresh} size="sm">
            ↻ Refresh
          </Button>
        )}
      </div>

      {/* Main Grid */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Left Column - Main Content */}
        <div className="space-y-6 lg:col-span-2">
          {/* Status & Progress */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Status</CardTitle>
            </CardHeader>
            <CardContent>
              <OrderStatusHeader
                status={order.status}
                estimatedCompletion={order.timeline.estimatedCompletion}
              />
            </CardContent>
          </Card>

          {/* Tabbed Content */}
          <Tabs defaultValue="overview">
            <TabsList>
              <TabsTrigger value="overview">Overview</TabsTrigger>
              <TabsTrigger value="usage">Usage</TabsTrigger>
              <TabsTrigger value="access">Access</TabsTrigger>
              <TabsTrigger value="timeline">Timeline</TabsTrigger>
            </TabsList>

            {/* Overview Tab */}
            <TabsContent value="overview">
              <div className="space-y-6">
                {/* Order Info */}
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">Order Details</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="grid gap-4 sm:grid-cols-2">
                      <InfoItem label="Resource Type" value={order.resourceType} />
                      <InfoItem label="Provider" value={order.providerName} />
                      <InfoItem label="Region" value={order.region} />
                      <InfoItem label="Duration" value={formatDuration(order.createdAt)} />
                      <InfoItem
                        label="Provider Address"
                        value={
                          <a
                            className="font-medium text-primary hover:underline"
                            href={accountLink(order.providerAddress)}
                            rel="noopener noreferrer"
                            target="_blank"
                          >
                            {truncateAddress(order.providerAddress)}
                          </a>
                        }
                      />
                      {order.txHash && (
                        <InfoItem
                          label="Tx Hash"
                          value={
                            <a
                              className="font-medium text-primary hover:underline"
                              href={txLink(order.txHash)}
                              rel="noopener noreferrer"
                              target="_blank"
                            >
                              {truncateAddress(order.txHash)}
                            </a>
                          }
                        />
                      )}
                    </div>
                  </CardContent>
                </Card>

                {/* Resource Details */}
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">Resources</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                      <ResourceStat label="CPU" value={`${order.resources.cpu} vCPU`} />
                      <ResourceStat label="Memory" value={`${order.resources.memory} GB`} />
                      <ResourceStat label="Storage" value={`${order.resources.storage} GB`} />
                      {order.resources.gpu && order.resources.gpu > 0 && (
                        <ResourceStat label="GPU" value={`${order.resources.gpu} GPU`} />
                      )}
                    </div>
                  </CardContent>
                </Card>
              </div>
            </TabsContent>

            {/* Usage Tab */}
            <TabsContent value="usage">
              <UsageMonitor usage={order.usage} />
            </TabsContent>

            {/* Access Tab */}
            <TabsContent value="access">
              <ResourceAccess access={order.access} />
            </TabsContent>

            {/* Timeline Tab */}
            <TabsContent value="timeline">
              <Card>
                <CardContent className="pt-6">
                  <OrderStatusTimelineView timeline={order.timeline} />
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>

        {/* Right Column - Sidebar */}
        <div className="space-y-6">
          {/* Cost Summary */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Cost Summary</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Hourly Rate</span>
                <span>{formatCurrency(order.cost.hourlyRate, order.cost.currency)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Total Cost</span>
                <span className="font-medium">
                  {formatCurrency(order.cost.totalCost, order.cost.currency)}
                </span>
              </div>
              {active && (
                <>
                  <div className="border-t border-border pt-3">
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Escrow Balance</span>
                      <span>
                        {formatCurrency(order.usage.cost.escrowBalance, order.cost.currency)}
                      </span>
                    </div>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Projected Monthly</span>
                    <span>
                      {formatCurrency(order.usage.cost.projectedMonthlyCost, order.cost.currency)}
                    </span>
                  </div>
                </>
              )}
            </CardContent>
          </Card>

          {/* Actions */}
          <OrderActions
            orderId={order.id}
            status={order.status}
            providerName={order.providerName}
            onExtend={handleExtend}
            onCancel={handleCancel}
            onSupport={handleSupport}
          />
        </div>
      </div>
    </div>
  );
}

// =============================================================================
// Helper Components
// =============================================================================

function InfoItem({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div>
      <dt className="text-sm text-muted-foreground">{label}</dt>
      <dd className="mt-1">{value}</dd>
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

function OrderDetailSkeleton() {
  return (
    <div className="container py-8">
      <Skeleton className="mb-6 h-5 w-32" />
      <div className="mb-6 space-y-2">
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

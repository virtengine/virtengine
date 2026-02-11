/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Alert, AlertDescription } from '@/components/ui/Alert';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/Table';
import { Eye, ArrowRight } from 'lucide-react';
import {
  useInvoices,
  useCurrentUsage,
  useCostProjection,
} from '@virtengine/portal/hooks/useBilling';
import type { Invoice, InvoiceStatus } from '@virtengine/portal/types/billing';
import {
  calculateOutstanding,
  hasOverdueInvoices,
  formatBillingAmount,
  formatBillingPeriod,
} from '@virtengine/portal/utils/billing';
import { UsageSummaryCard } from './UsageSummaryCard';
import { CostProjectionCard } from './CostProjectionCard';
import { CostTrendChart } from './CostTrendChart';
import { useUsageHistory } from '@virtengine/portal/hooks/useBilling';
import { MFAChallenge } from '@/components/mfa';
import { useMFAGate } from '@/features/mfa';
import { useChainQuery } from '@/hooks/useChainQuery';
import { useWallet } from '@/lib/portal-adapter';
import { formatTokenAmount } from '@/lib/utils';

function thirtyDaysAgo(): Date {
  const d = new Date();
  d.setDate(d.getDate() - 30);
  return d;
}

const STATUS_VARIANT: Record<
  InvoiceStatus,
  'default' | 'success' | 'warning' | 'destructive' | 'secondary'
> = {
  draft: 'secondary',
  pending: 'warning',
  paid: 'success',
  overdue: 'destructive',
  cancelled: 'secondary',
};

interface BillingDashboardProps {
  onViewInvoice?: (id: string) => void;
  onViewAllInvoices?: () => void;
  onViewUsage?: () => void;
}

export function BillingDashboard({
  onViewInvoice,
  onViewAllInvoices,
  onViewUsage,
}: BillingDashboardProps) {
  const [withdrawalNotice, setWithdrawalNotice] = useState(false);
  const { gateAction, challengeProps } = useMFAGate();
  const wallet = useWallet();
  const activeAddress = wallet.accounts?.[wallet.activeAccountIndex ?? 0]?.address ?? null;

  const { data: usage, isLoading: usageLoading } = useCurrentUsage();
  const { data: projection, isLoading: projectionLoading } = useCostProjection();
  const { data: invoices, isLoading: invoicesLoading } = useInvoices({ limit: 5 });
  const { data: trendData, isLoading: trendLoading } = useUsageHistory({
    startDate: thirtyDaysAgo(),
    endDate: new Date(),
    granularity: 'day',
  });
  const { data: claimableRewards, isLoading: rewardsLoading } = useChainQuery(
    async (client) => {
      if (!activeAddress) return null;
      return client.settlement.estimateRewards(activeAddress);
    },
    [activeAddress]
  );

  const rewardCoin = claimableRewards?.totalClaimable?.[0];
  const claimableValue = rewardCoin ? formatTokenAmount(rewardCoin.amount) : '0';
  const claimableDenom = rewardCoin?.denom
    ? rewardCoin.denom === 'uve'
      ? 'VE'
      : rewardCoin.denom.toUpperCase()
    : 'VE';

  const outstanding = calculateOutstanding(invoices);
  const hasOverdue = hasOverdueInvoices(invoices);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Billing</h1>
        <Button
          variant="outline"
          onClick={() =>
            gateAction({
              transactionType: 'withdrawal',
              actionDescription: 'Request a withdrawal',
              onAuthorized: () => setWithdrawalNotice(true),
            })
          }
        >
          Request Withdrawal
        </Button>
      </div>

      {withdrawalNotice && (
        <Alert variant="success">
          <AlertDescription>
            Withdrawal request submitted. You&apos;ll receive a confirmation once it&apos;s
            processed.
          </AlertDescription>
        </Alert>
      )}

      {/* Summary cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <UsageSummaryCard
          title="Current Period"
          value={usage?.totalCost ?? '0'}
          currency="VIRT"
          loading={usageLoading}
        />
        <UsageSummaryCard
          title="Projected"
          value={projection?.currentPeriod.projected ?? '0'}
          currency="VIRT"
          subtitle={
            projection ? `${projection.currentPeriod.daysRemaining} days remaining` : undefined
          }
          loading={projectionLoading}
        />
        <UsageSummaryCard
          title="Outstanding"
          value={outstanding}
          currency="VIRT"
          status={hasOverdue ? 'warning' : 'normal'}
          loading={invoicesLoading}
        />
        <UsageSummaryCard
          title="Deployments"
          value={usage?.byDeployment.length.toString() ?? '0'}
          subtitle="Active deployments"
          loading={usageLoading}
        />
        <UsageSummaryCard
          title="Claimable Rewards"
          value={claimableValue}
          currency={claimableDenom}
          subtitle={activeAddress ? undefined : 'Connect wallet'}
          loading={rewardsLoading}
        />
      </div>

      {/* Charts row */}
      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0">
            <CardTitle className="text-base">Cost Trend (30 Days)</CardTitle>
            {onViewUsage && (
              <Button variant="ghost" size="sm" onClick={onViewUsage}>
                View Details
                <ArrowRight className="ml-1 h-4 w-4" />
              </Button>
            )}
          </CardHeader>
          <CardContent>
            <CostTrendChart data={trendData} loading={trendLoading} />
          </CardContent>
        </Card>

        <CostProjectionCard projection={projection} loading={projectionLoading} />
      </div>

      {/* Recent invoices */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <CardTitle className="text-base">Recent Invoices</CardTitle>
          {onViewAllInvoices && (
            <Button variant="ghost" size="sm" onClick={onViewAllInvoices}>
              View All
              <ArrowRight className="ml-1 h-4 w-4" />
            </Button>
          )}
        </CardHeader>
        <CardContent className="p-0">
          {invoicesLoading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 3 }, (_, i) => (
                <Skeleton key={`invoice-skeleton-${i}`} className="h-12 w-full" />
              ))}
            </div>
          ) : !invoices || invoices.length === 0 ? (
            <p className="p-6 text-center text-sm text-muted-foreground">No invoices yet</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Invoice #</TableHead>
                  <TableHead>Period</TableHead>
                  <TableHead className="text-right">Amount</TableHead>
                  <TableHead className="text-center">Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {invoices.slice(0, 5).map((invoice: Invoice) => (
                  <TableRow key={invoice.id}>
                    <TableCell className="font-medium">{invoice.number}</TableCell>
                    <TableCell>
                      {formatBillingPeriod(invoice.period.start, invoice.period.end)}
                    </TableCell>
                    <TableCell className="text-right">
                      {formatBillingAmount(invoice.total, invoice.currency)}
                    </TableCell>
                    <TableCell className="text-center">
                      <Badge variant={STATUS_VARIANT[invoice.status]} size="sm">
                        {invoice.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      {onViewInvoice && (
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() => onViewInvoice(invoice.id)}
                          aria-label={`View invoice ${invoice.number}`}
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <MFAChallenge {...challengeProps} />
    </div>
  );
}

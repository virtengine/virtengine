/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, useRef, useState } from 'react';
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
import {
  billingWithdrawalRequestsEqual,
  buildBillingWithdrawalRequest,
  submitBillingWithdrawal,
  type BillingWithdrawalAdapter,
  type BillingWithdrawalContext,
  type BillingWithdrawalMutationResult,
} from './withdrawal-mutation';

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
  withdrawalAdapter?: BillingWithdrawalAdapter;
  withdrawalContext?: BillingWithdrawalContext;
}

export function BillingDashboard({
  onViewInvoice,
  onViewAllInvoices,
  onViewUsage,
  withdrawalAdapter,
  withdrawalContext,
}: BillingDashboardProps) {
  const [withdrawalStatus, setWithdrawalStatus] = useState<
    'idle' | 'submitting' | 'committed' | 'error'
  >('idle');
  const [withdrawalResult, setWithdrawalResult] =
    useState<Readonly<BillingWithdrawalMutationResult> | null>(null);
  const withdrawalRequest = buildBillingWithdrawalRequest(withdrawalContext);
  const requestRef = useRef(withdrawalRequest);
  requestRef.current = withdrawalRequest;
  const inFlightRef = useRef(false);
  const generationRef = useRef(0);
  const authorizationRef = useRef(0);
  const mountedRef = useRef(true);
  const activeControllerRef = useRef<AbortController | null>(null);
  const authorityRef = useRef({ withdrawalAdapter, withdrawalRequest });
  authorityRef.current = { withdrawalAdapter, withdrawalRequest };
  const { gateAction, challengeProps } = useMFAGate();

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  useEffect(() => {
    activeControllerRef.current?.abort();
    activeControllerRef.current = null;
    generationRef.current += 1;
    authorizationRef.current += 1;
    inFlightRef.current = false;
    setWithdrawalResult(null);
    setWithdrawalStatus('idle');
    return () => {
      activeControllerRef.current?.abort();
      activeControllerRef.current = null;
      generationRef.current += 1;
      authorizationRef.current += 1;
      inFlightRef.current = false;
    };
  }, [withdrawalAdapter, withdrawalContext?.accountAddress, withdrawalContext?.chainId]);

  const requestWithdrawal = async (
    adapter: BillingWithdrawalAdapter,
    request: Readonly<BillingWithdrawalMutationResult['request']>,
    generation: number,
    authorization: number
  ) => {
    if (
      !mountedRef.current ||
      generationRef.current !== generation ||
      authorityRef.current.withdrawalAdapter !== adapter ||
      !billingWithdrawalRequestsEqual(authorityRef.current.withdrawalRequest, request) ||
      authorizationRef.current !== authorization ||
      inFlightRef.current
    ) {
      return;
    }
    authorizationRef.current += 1;
    inFlightRef.current = true;
    const controller = new AbortController();
    activeControllerRef.current = controller;
    setWithdrawalResult(null);
    setWithdrawalStatus('submitting');
    try {
      const result = await submitBillingWithdrawal({
        adapter,
        request,
        signal: controller.signal,
        getCurrentRequest: () =>
          mountedRef.current &&
          generationRef.current === generation &&
          authorityRef.current.withdrawalAdapter === adapter
            ? requestRef.current
            : null,
      });
      if (
        !mountedRef.current ||
        generationRef.current !== generation ||
        authorityRef.current.withdrawalAdapter !== adapter ||
        !billingWithdrawalRequestsEqual(authorityRef.current.withdrawalRequest, request) ||
        controller.signal.aborted
      ) {
        return;
      }
      setWithdrawalResult(result);
      setWithdrawalStatus('committed');
    } catch {
      if (generationRef.current === generation) setWithdrawalStatus('error');
    } finally {
      if (generationRef.current === generation) {
        activeControllerRef.current = null;
        inFlightRef.current = false;
      }
    }
  };

  const authorizeWithdrawal = () => {
    if (!withdrawalAdapter || !withdrawalRequest) return;
    const adapter = withdrawalAdapter;
    const request = withdrawalRequest;
    const generation = generationRef.current;
    const authorization = authorizationRef.current + 1;
    authorizationRef.current = authorization;
    void gateAction({
      transactionType: 'withdrawal',
      actionDescription: 'Request a withdrawal',
      onAuthorized: () => requestWithdrawal(adapter, request, generation, authorization),
    });
  };

  const { data: usage, isLoading: usageLoading } = useCurrentUsage();
  const { data: projection, isLoading: projectionLoading } = useCostProjection();
  const { data: invoices, isLoading: invoicesLoading } = useInvoices({ limit: 5 });
  const { data: trendData, isLoading: trendLoading } = useUsageHistory({
    startDate: thirtyDaysAgo(),
    endDate: new Date(),
    granularity: 'day',
  });

  const outstanding = calculateOutstanding(invoices);
  const hasOverdue = hasOverdueInvoices(invoices);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Billing</h1>
        <Button
          variant="outline"
          disabled={!withdrawalAdapter || !withdrawalRequest || withdrawalStatus === 'submitting'}
          onClick={authorizeWithdrawal}
        >
          {withdrawalStatus === 'submitting' ? 'Submitting Withdrawal' : 'Request Withdrawal'}
        </Button>
      </div>

      {!withdrawalAdapter || !withdrawalRequest ? (
        <Alert>
          <AlertDescription>
            Withdrawals are unavailable because no authoritative billing mutation is configured.
          </AlertDescription>
        </Alert>
      ) : withdrawalStatus === 'error' ? (
        <Alert variant="destructive">
          <AlertDescription>
            Withdrawal was not committed. The request failed or returned invalid evidence.
          </AlertDescription>
        </Alert>
      ) : withdrawalStatus === 'committed' &&
        withdrawalResult &&
        billingWithdrawalRequestsEqual(withdrawalResult.request, withdrawalRequest) ? (
        <Alert variant="success">
          <AlertDescription>
            Withdrawal committed in transaction {withdrawalResult.txHash} at height{' '}
            {withdrawalResult.blockHeight}.
          </AlertDescription>
        </Alert>
      ) : null}

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

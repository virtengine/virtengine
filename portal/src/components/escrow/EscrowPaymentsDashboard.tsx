/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, useMemo, useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import { env } from '@/config';
import { useWallet } from '@/lib/portal-adapter';
import { usePriceConversion } from '@/hooks/usePriceConversion';
import { useCustomerDashboardStore } from '@/stores/customerDashboardStore';
import { EscrowBalance } from './EscrowBalance';
import { DepositModal } from './DepositModal';
import { PayoutHistory } from './PayoutHistory';
import { SettlementLog } from './SettlementLog';
import { TransactionHistory } from './TransactionHistory';
import { WithdrawForm } from './WithdrawForm';
import type {
  EscrowAccount,
  EscrowTransaction,
  FiatRates,
  PayoutRecord,
  SettlementEvent,
} from './data';
import type {
  EscrowMutationAdapter,
  EscrowMutationContext,
  EscrowMutationResultProjector,
} from './mutations';
import { isValidEscrowMutationContext } from './mutations';

interface EscrowPaymentsDashboardProps {
  mutationAdapter?: EscrowMutationAdapter;
  mutationContext?: EscrowMutationContext;
  resultProjector?: EscrowMutationResultProjector;
}

export function EscrowPaymentsDashboard({
  mutationAdapter,
  mutationContext,
  resultProjector,
}: EscrowPaymentsDashboardProps = {}) {
  const [depositOpen, setDepositOpen] = useState(false);
  const wallet = useWallet();
  const account = wallet.accounts[wallet.activeAccountIndex];
  const fetchDashboard = useCustomerDashboardStore((s) => s.fetchDashboard);
  const allocations = useCustomerDashboardStore((s) => s.allocations);
  const billing = useCustomerDashboardStore((s) => s.billing);
  const escrowAccounts = useCustomerDashboardStore((s) => s.escrowAccounts);
  const escrowPayments = useCustomerDashboardStore((s) => s.escrowPayments);
  const isLoading = useCustomerDashboardStore((s) => s.isLoading);
  const error = useCustomerDashboardStore((s) => s.error);
  const { rate, stale, isLoading: rateLoading } = usePriceConversion();
  const boundMutationContext =
    account?.address &&
    wallet.chainId &&
    isValidEscrowMutationContext(mutationContext) &&
    mutationContext.accountAddress === account.address &&
    mutationContext.chainId === wallet.chainId
      ? mutationContext
      : undefined;
  const mutationsAvailable = Boolean(mutationAdapter && boundMutationContext && resultProjector);

  useEffect(() => {
    if (!account?.address) return;
    void fetchDashboard(account.address);
  }, [account?.address, fetchDashboard]);

  const fiatRates = useMemo<FiatRates>(
    () => ({
      usd: typeof rate === 'number' && Number.isFinite(rate) ? rate : undefined,
    }),
    [rate]
  );

  const escrowAccount = useMemo<EscrowAccount>(() => {
    const pendingSettlement = escrowPayments.reduce(
      (sum, payment) => sum + payment.balanceAmount,
      0
    );
    const latestUpdate = escrowAccounts
      .map((entry) => entry.settledAt)
      .filter((value): value is string => Boolean(value))
      .sort()
      .at(-1);

    return {
      accountId: account?.address ?? 'wallet-not-connected',
      currency: 'UVE',
      lockedBalance: billing.outstandingBalance,
      availableBalance: Math.max(billing.outstandingBalance - pendingSettlement, 0),
      pendingSettlement,
      walletBalance: 0,
      lastUpdated: latestUpdate ?? new Date().toISOString(),
    };
  }, [account?.address, billing.outstandingBalance, escrowAccounts, escrowPayments]);

  const transactions = useMemo<EscrowTransaction[]>(() => {
    const deposits = escrowAccounts
      .filter((entry) => entry.transferred > 0)
      .map((entry) => ({
        id: `deposit-${entry.scope}-${entry.xid}`,
        type: 'Deposit' as const,
        direction: 'credit' as const,
        status: entry.state === 'closed' ? ('completed' as const) : ('processing' as const),
        amount: entry.transferred,
        currency: 'UVE',
        reference: `${entry.scope} escrow account`,
        allocation: entry.xid || undefined,
        occurredAt: entry.settledAt ?? new Date().toISOString(),
      }));

    const spendEntries = allocations
      .filter((allocation) => allocation.totalSpent > 0)
      .map((allocation) => ({
        id: `settlement-${allocation.id}`,
        type: 'Settlement' as const,
        direction: 'debit' as const,
        status:
          allocation.status === 'failed'
            ? ('failed' as const)
            : allocation.status === 'terminated'
              ? ('completed' as const)
              : ('processing' as const),
        amount: allocation.totalSpent,
        currency: allocation.currency || 'UVE',
        reference: `Order ${allocation.orderId}`,
        allocation: allocation.id,
        occurredAt: allocation.updatedAt,
      }));

    return [...deposits, ...spendEntries].sort(
      (a, b) => new Date(b.occurredAt).getTime() - new Date(a.occurredAt).getTime()
    );
  }, [allocations, escrowAccounts]);

  const settlements = useMemo<SettlementEvent[]>(
    () =>
      allocations
        .filter((allocation) => allocation.totalSpent > 0)
        .map((allocation) => ({
          id: `settlement-${allocation.id}`,
          allocation: allocation.id,
          provider: allocation.providerName,
          period: `Last updated ${new Date(allocation.updatedAt).toLocaleString()}`,
          status:
            allocation.status === 'failed'
              ? ('disputed' as const)
              : allocation.status === 'terminated'
                ? ('posted' as const)
                : ('pending' as const),
          usageSummary: `${allocation.resources.cpu} CPU · ${allocation.resources.memory} GB RAM${allocation.resources.gpu ? ` · ${allocation.resources.gpu} GPU` : ''}`,
          amount: allocation.totalSpent,
          currency: allocation.currency || 'UVE',
          postedAt: allocation.updatedAt,
          breakdown: [
            {
              label: allocation.offeringName,
              units: `${allocation.resources.cpu} CPU / ${allocation.resources.memory} GB`,
              rate: `${allocation.costPerHour.toFixed(2)} ${(allocation.currency || 'UVE').toUpperCase()}/hr`,
              amount: allocation.totalSpent,
            },
          ],
        }))
        .sort((a, b) => new Date(b.postedAt).getTime() - new Date(a.postedAt).getTime()),
    [allocations]
  );

  const payoutHistory = useMemo<PayoutRecord[]>(
    () =>
      escrowPayments
        .filter((payment) => payment.withdrawnAmount > 0)
        .map((payment) => ({
          id: payment.paymentId,
          provider:
            allocations.find((allocation) => allocation.id === payment.xid)?.providerName ??
            payment.xid,
          status:
            payment.state === 'closed'
              ? ('completed' as const)
              : payment.state === 'overdrawn'
                ? ('failed' as const)
                : ('processing' as const),
          method: 'On-chain',
          amount: payment.withdrawnAmount,
          currency: 'UVE',
          txHash: undefined,
          requestedAt:
            allocations.find((allocation) => allocation.id === payment.xid)?.updatedAt ??
            new Date().toISOString(),
          completedAt:
            payment.state === 'closed'
              ? allocations.find((allocation) => allocation.id === payment.xid)?.updatedAt
              : undefined,
        }))
        .sort((a, b) => new Date(b.requestedAt).getTime() - new Date(a.requestedAt).getTime()),
    [allocations, escrowPayments]
  );

  const pricingFeedLabel = rateLoading
    ? 'Loading live fiat conversion…'
    : rate
      ? stale
        ? 'Using cached UVE/USD conversion'
        : 'Using live UVE/USD conversion'
      : 'Fiat conversion unavailable';

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">Escrow &amp; Payments</h1>
          <p className="text-sm text-muted-foreground">
            Track live escrow balances, derived usage settlements, and withdrawal state in one
            place.
          </p>
        </div>
        <div className="text-xs text-muted-foreground">{pricingFeedLabel}</div>
      </div>

      {!account?.address && (
        <div className="rounded-lg border border-border bg-card p-4 text-sm text-muted-foreground">
          Connect your wallet to load live escrow balances and billing state.
        </div>
      )}

      {!mutationsAvailable && (
        <div className="rounded-lg border border-border bg-card p-4 text-sm text-muted-foreground">
          Escrow deposits and wallet withdrawals are unavailable because no authoritative signing
          and broadcast adapter is configured.
        </div>
      )}

      {error && (
        <div className="rounded-lg border border-destructive bg-destructive/10 p-4">
          <p className="font-medium text-destructive">Failed to load escrow data</p>
          <p className="mt-1 text-sm text-muted-foreground">{error}</p>
        </div>
      )}

      <EscrowBalance
        account={escrowAccount}
        fiatRates={fiatRates}
        onDeposit={() => setDepositOpen(true)}
        onWithdraw={() => {
          document.getElementById('withdraw-form')?.scrollIntoView({ behavior: 'smooth' });
        }}
        actionsAvailable={mutationsAvailable}
      />

      <div className="grid gap-6 lg:grid-cols-[1.4fr_1fr]">
        <WithdrawForm
          account={escrowAccount}
          fiatRates={fiatRates}
          fiatOffRampUrl={env.fiatOffRampUrl || undefined}
          mutationAdapter={mutationAdapter}
          mutationContext={boundMutationContext}
          resultProjector={resultProjector}
        />
        <div className="rounded-lg border border-border/60 bg-muted/30 p-6">
          <h2 className="text-lg font-semibold">Deposit Guidance</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Escrow deposits cover projected usage for new allocations. On-chain account balances are
            authoritative and wallet confirmations are required for every top-up.
          </p>
          <div className="mt-4 space-y-3 text-sm">
            <div className="rounded-lg bg-background p-3">
              <p className="text-xs uppercase text-muted-foreground">Recommended buffer</p>
              <p className="mt-1 text-base font-semibold">
                {(billing.currentPeriodCost * 0.25).toFixed(2)} UVE
              </p>
            </div>
            <div className="rounded-lg bg-background p-3">
              <p className="text-xs uppercase text-muted-foreground">Top-up mode</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Recurring deposits are not exposed by the live portal APIs. Use signed wallet
                deposits for each escrow refill.
              </p>
            </div>
          </div>
        </div>
      </div>

      <Tabs defaultValue="transactions" className="space-y-4">
        <TabsList>
          <TabsTrigger value="transactions">Transactions</TabsTrigger>
          <TabsTrigger value="settlements">Settlements</TabsTrigger>
          <TabsTrigger value="payouts">Payouts</TabsTrigger>
        </TabsList>
        <TabsContent value="transactions">
          <TransactionHistory transactions={transactions} />
        </TabsContent>
        <TabsContent value="settlements">
          <SettlementLog settlements={settlements} />
        </TabsContent>
        <TabsContent value="payouts">
          <PayoutHistory payouts={payoutHistory} />
        </TabsContent>
      </Tabs>

      <DepositModal
        open={depositOpen}
        onOpenChange={setDepositOpen}
        account={escrowAccount}
        fiatRates={fiatRates}
        mutationAdapter={mutationAdapter}
        mutationContext={boundMutationContext}
        resultProjector={resultProjector}
      />

      {isLoading && !error && (
        <div className="text-sm text-muted-foreground">
          Refreshing live escrow and billing state…
        </div>
      )}
    </div>
  );
}

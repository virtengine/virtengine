/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import { env } from '@/config';
import { EscrowBalance } from './EscrowBalance';
import { DepositModal } from './DepositModal';
import { PayoutHistory } from './PayoutHistory';
import { SettlementLog } from './SettlementLog';
import { TransactionHistory } from './TransactionHistory';
import { WithdrawForm } from './WithdrawForm';
import {
  escrowAccount,
  escrowTransactions,
  fiatRates,
  payoutHistory,
  settlementEvents,
} from './data';

export function EscrowPaymentsDashboard() {
  const [depositOpen, setDepositOpen] = useState(false);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">Escrow &amp; Payments</h1>
          <p className="text-sm text-muted-foreground">
            Track escrow deposits, settlements, and provider payouts in one place.
          </p>
        </div>
        <div className="text-xs text-muted-foreground">
          Fiat estimates powered by internal pricing feed
        </div>
      </div>

      <EscrowBalance
        account={escrowAccount}
        fiatRates={fiatRates}
        onDeposit={() => setDepositOpen(true)}
        onWithdraw={() => {
          document.getElementById('withdraw-form')?.scrollIntoView({ behavior: 'smooth' });
        }}
      />

      <div className="grid gap-6 lg:grid-cols-[1.4fr_1fr]">
        <WithdrawForm
          account={escrowAccount}
          fiatRates={fiatRates}
          fiatOffRampUrl={env.fiatOffRampUrl || undefined}
        />
        <div className="rounded-lg border border-border/60 bg-muted/30 p-6">
          <h2 className="text-lg font-semibold">Deposit Guidance</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Escrow deposits cover projected usage for new allocations. Keep at least 7 days of
            runway to avoid interruptions.
          </p>
          <div className="mt-4 space-y-3 text-sm">
            <div className="rounded-lg bg-background p-3">
              <p className="text-xs uppercase text-muted-foreground">Recommended buffer</p>
              <p className="mt-1 text-base font-semibold">+10,000 VIRT</p>
            </div>
            <div className="rounded-lg bg-background p-3">
              <p className="text-xs uppercase text-muted-foreground">Auto-top up</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Coming soon · Enable recurring wallet transfers.
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
          <TransactionHistory transactions={escrowTransactions} />
        </TabsContent>
        <TabsContent value="settlements">
          <SettlementLog settlements={settlementEvents} />
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
      />
    </div>
  );
}

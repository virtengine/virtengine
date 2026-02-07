/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { ArrowUpRight, ArrowDownLeft } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { formatRelativeTime } from '@/lib/utils';
import type { EscrowAccount, FiatRates } from './data';
import { formatFiat, formatToken } from './utils';

interface EscrowBalanceProps {
  account: EscrowAccount;
  fiatRates: FiatRates;
  onDeposit: () => void;
  onWithdraw: () => void;
}

export function EscrowBalance({ account, fiatRates, onDeposit, onWithdraw }: EscrowBalanceProps) {
  return (
    <Card className="relative overflow-hidden">
      <div className="absolute right-0 top-0 h-32 w-32 -translate-y-1/3 translate-x-1/3 rounded-full bg-primary/10" />
      <CardHeader className="relative">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <CardTitle className="text-xl">Escrow Balance</CardTitle>
            <p className="mt-1 text-sm text-muted-foreground">
              Account {account.accountId} · Updated {formatRelativeTime(account.lastUpdated)}
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" onClick={onWithdraw}>
              <ArrowDownLeft className="h-4 w-4" />
              Withdraw
            </Button>
            <Button onClick={onDeposit}>
              <ArrowUpRight className="h-4 w-4" />
              Deposit
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="relative grid gap-6 md:grid-cols-3">
        <div className="space-y-2 rounded-lg border border-border/60 bg-background/60 p-4">
          <p className="text-sm text-muted-foreground">Locked in escrow</p>
          <p className="text-2xl font-semibold">
            {formatToken(account.lockedBalance, account.currency)}
          </p>
          <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
            <span>{formatFiat(account.lockedBalance * fiatRates.usd, 'USD')} USD</span>
            <span>·</span>
            <span>{formatFiat(account.lockedBalance * fiatRates.eur, 'EUR')} EUR</span>
          </div>
        </div>
        <div className="space-y-2 rounded-lg border border-border/60 bg-background/60 p-4">
          <p className="text-sm text-muted-foreground">Available balance</p>
          <p className="text-2xl font-semibold">
            {formatToken(account.availableBalance, account.currency)}
          </p>
          <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
            <span>{formatFiat(account.availableBalance * fiatRates.usd, 'USD')} USD</span>
            <span>·</span>
            <span>{formatFiat(account.availableBalance * fiatRates.eur, 'EUR')} EUR</span>
          </div>
        </div>
        <div className="space-y-2 rounded-lg border border-border/60 bg-background/60 p-4">
          <p className="text-sm text-muted-foreground">Pending settlement</p>
          <p className="text-2xl font-semibold">
            {formatToken(account.pendingSettlement, account.currency)}
          </p>
          <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
            <Badge variant="warning" size="sm">
              In review
            </Badge>
            <span>Next batch in ~3 hours</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

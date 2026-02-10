/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useMemo, useState } from 'react';
import Link from 'next/link';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import type { EscrowAccount, FiatRates } from './data';
import { formatFiat, formatToken } from './utils';

interface WithdrawFormProps {
  account: EscrowAccount;
  fiatRates: FiatRates;
  fiatOffRampUrl?: string;
}

export function WithdrawForm({ account, fiatRates, fiatOffRampUrl }: WithdrawFormProps) {
  const [amount, setAmount] = useState('250');
  const [destination, setDestination] = useState('wallet');
  const [submitted, setSubmitted] = useState(false);

  const numericAmount = useMemo(() => Number(amount), [amount]);
  const amountError =
    !amount || Number.isNaN(numericAmount) || numericAmount <= 0
      ? 'Enter a valid amount'
      : numericAmount > account.availableBalance
        ? 'Amount exceeds available balance'
        : '';

  const destinationError =
    destination === 'bank' && !fiatOffRampUrl
      ? 'Fiat off-ramp is not configured for this workspace'
      : '';

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    if (amountError || destinationError) return;
    setSubmitted(true);
  };

  return (
    <Card id="withdraw-form">
      <CardHeader>
        <CardTitle className="text-lg">Withdraw Funds</CardTitle>
        <p className="text-sm text-muted-foreground">
          Available: {formatToken(account.availableBalance, account.currency)}
        </p>
      </CardHeader>
      <CardContent>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="space-y-2">
            <Label htmlFor="withdraw-amount">Amount</Label>
            <Input
              id="withdraw-amount"
              type="number"
              inputMode="decimal"
              min={0}
              step="0.01"
              value={amount}
              onChange={(event) => setAmount(event.target.value)}
              error={Boolean(amountError)}
            />
            {amountError ? (
              <p className="text-xs text-destructive">{amountError}</p>
            ) : (
              <p className="text-xs text-muted-foreground">
                {formatFiat(numericAmount * fiatRates.usd, 'USD')} USD ·{' '}
                {formatFiat(numericAmount * fiatRates.eur, 'EUR')} EUR
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label>Destination</Label>
            <Select value={destination} onValueChange={setDestination}>
              <SelectTrigger error={Boolean(destinationError)}>
                <SelectValue placeholder="Select destination" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="wallet">Connected wallet</SelectItem>
                <SelectItem value="bank" disabled={!fiatOffRampUrl}>
                  Bank account (fiat off-ramp)
                </SelectItem>
                <SelectItem value="treasury">Organization treasury</SelectItem>
              </SelectContent>
            </Select>
            {destinationError && <p className="text-xs text-destructive">{destinationError}</p>}
          </div>

          {fiatOffRampUrl && (
            <div className="rounded-lg border border-border/60 bg-muted/40 p-3 text-xs">
              Need to cash out to fiat?{' '}
              <Link
                className="text-primary hover:underline"
                href={fiatOffRampUrl}
                target="_blank"
                rel="noreferrer"
              >
                Open off-ramp
              </Link>
            </div>
          )}

          {submitted && (
            <Alert variant="success">
              <AlertTitle>Withdrawal requested</AlertTitle>
              <AlertDescription>
                Your withdrawal is queued and pending on-chain confirmation.
              </AlertDescription>
            </Alert>
          )}

          <Button type="submit" disabled={Boolean(amountError || destinationError)}>
            Request withdrawal
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

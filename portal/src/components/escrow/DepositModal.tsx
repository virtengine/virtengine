/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useMemo, useState } from 'react';
import {
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Modal,
} from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import type { EscrowAccount, FiatRates } from './data';
import { formatFiat, formatToken } from './utils';

interface DepositModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  account: EscrowAccount;
  fiatRates: FiatRates;
}

const MIN_DEPOSIT = 50;

export function DepositModal({ open, onOpenChange, account, fiatRates }: DepositModalProps) {
  const [amount, setAmount] = useState('500');
  const [source, setSource] = useState('wallet');
  const [submitted, setSubmitted] = useState(false);

  const numericAmount = useMemo(() => Number(amount), [amount]);
  const amountError =
    !amount || Number.isNaN(numericAmount) || numericAmount < MIN_DEPOSIT
      ? `Minimum deposit is ${MIN_DEPOSIT} ${account.currency}`
      : numericAmount > account.walletBalance
        ? 'Amount exceeds available wallet balance'
        : '';

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    if (amountError) return;
    setSubmitted(true);
  };

  return (
    <Modal
      open={open}
      onOpenChange={(next) => {
        onOpenChange(next);
        if (!next) {
          setSubmitted(false);
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Deposit to Escrow</DialogTitle>
        </DialogHeader>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="space-y-2">
            <Label htmlFor="deposit-amount">Amount</Label>
            <Input
              id="deposit-amount"
              type="number"
              inputMode="decimal"
              min={MIN_DEPOSIT}
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
            <Label>Funding source</Label>
            <Select value={source} onValueChange={setSource}>
              <SelectTrigger>
                <SelectValue placeholder="Select funding source" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="wallet">
                  Wallet balance ({formatToken(account.walletBalance, account.currency)})
                </SelectItem>
                <SelectItem value="wire">Wire transfer (1-2 days)</SelectItem>
                <SelectItem value="card">Card top-up (instant)</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {submitted && (
            <Alert variant="success">
              <AlertTitle>Deposit queued</AlertTitle>
              <AlertDescription>
                Your deposit request has been created. Sign the on-chain transaction to finalize the
                escrow transfer.
              </AlertDescription>
            </Alert>
          )}

          <DialogFooter>
            <Button variant="outline" type="button" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={Boolean(amountError)}>
              Continue
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Modal>
  );
}

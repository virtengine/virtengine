/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import Link from 'next/link';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import type { EscrowAccount, FiatRates } from './data';
import {
  buildEscrowMutationRequest,
  escrowMutationRequestsEqual,
  isValidEscrowMutationAmount,
  isValidEscrowMutationContext,
  submitEscrowMutation,
  type EscrowCommittedResult,
  type EscrowMutationAdapter,
  type EscrowMutationContext,
  type EscrowMutationResultProjector,
  type EscrowMutationStatus,
} from './mutations';
import { formatFiatEstimates, formatToken } from './utils';

interface WithdrawFormProps {
  account: EscrowAccount;
  fiatRates: FiatRates;
  fiatOffRampUrl?: string;
  mutationAdapter?: EscrowMutationAdapter;
  mutationContext?: EscrowMutationContext;
  resultProjector?: EscrowMutationResultProjector;
  mutationTimeoutMs?: number;
}

export function WithdrawForm({
  account,
  fiatRates,
  fiatOffRampUrl,
  mutationAdapter,
  mutationContext,
  resultProjector,
  mutationTimeoutMs,
}: WithdrawFormProps) {
  const [amount, setAmount] = useState('250');
  const available = Boolean(
    mutationAdapter && isValidEscrowMutationContext(mutationContext) && resultProjector
  );
  const [status, setStatus] = useState<EscrowMutationStatus>('idle');
  const [result, setResult] = useState<EscrowCommittedResult | null>(null);
  const inFlight = useRef(false);
  const mutationSession = useRef(0);
  const activeSubmission = useRef<AbortController | null>(null);
  const authorityRef = useRef({ mutationAdapter, resultProjector });
  const committedAuthorityRef = useRef<{
    mutationAdapter: EscrowMutationAdapter;
    resultProjector: EscrowMutationResultProjector;
  } | null>(null);
  authorityRef.current = { mutationAdapter, resultProjector };
  const currentRequest = buildEscrowMutationRequest(
    'withdraw',
    mutationContext ?? { chainId: '', accountAddress: '', escrowAccountId: '' },
    amount,
    account.currency
  );
  const currentRequestRef = useRef(currentRequest);
  currentRequestRef.current = currentRequest;
  const resultIsCurrent = Boolean(
    result &&
    escrowMutationRequestsEqual(result.request, currentRequest) &&
    committedAuthorityRef.current?.mutationAdapter === mutationAdapter &&
    committedAuthorityRef.current?.resultProjector === resultProjector
  );
  const effectiveStatus = !available
    ? 'unavailable'
    : status === 'committed' && !resultIsCurrent
      ? 'idle'
      : status;

  useEffect(() => {
    activeSubmission.current?.abort();
    activeSubmission.current = null;
    mutationSession.current += 1;
    inFlight.current = false;
    setResult(null);
    setStatus('idle');
    return () => {
      activeSubmission.current?.abort();
      activeSubmission.current = null;
      mutationSession.current += 1;
      inFlight.current = false;
    };
  }, [
    mutationAdapter,
    mutationContext?.accountAddress,
    mutationContext?.chainId,
    mutationContext?.escrowAccountId,
    account.currency,
    resultProjector,
  ]);

  const numericAmount = useMemo(() => Number(amount), [amount]);
  const fiatEstimates = formatFiatEstimates(numericAmount, fiatRates);
  const amountError = !isValidEscrowMutationAmount(amount)
    ? 'Enter a valid amount'
    : numericAmount > account.availableBalance
      ? 'Amount exceeds available balance'
      : '';

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (amountError || inFlight.current) return;
    if (!mutationAdapter || !resultProjector || !currentRequest) {
      setStatus('unavailable');
      return;
    }

    inFlight.current = true;
    const submissionController = new AbortController();
    activeSubmission.current = submissionController;
    const submissionSession = mutationSession.current;
    const submittedAdapter = mutationAdapter;
    const submittedProjector = resultProjector;
    setStatus('submitting');
    setResult(null);
    try {
      const committed = await submitEscrowMutation({
        adapter: mutationAdapter,
        projector: resultProjector,
        request: currentRequest,
        signal: submissionController.signal,
        getCurrentRequest: () =>
          mutationSession.current === submissionSession &&
          authorityRef.current.mutationAdapter === submittedAdapter &&
          authorityRef.current.resultProjector === submittedProjector
            ? currentRequestRef.current
            : null,
        timeoutMs: mutationTimeoutMs,
      });
      if (mutationSession.current !== submissionSession) return;
      committedAuthorityRef.current = {
        mutationAdapter: submittedAdapter,
        resultProjector: submittedProjector,
      };
      setResult(committed);
      setStatus('committed');
    } catch {
      if (mutationSession.current === submissionSession) setStatus('error');
    } finally {
      if (mutationSession.current === submissionSession) {
        activeSubmission.current = null;
        inFlight.current = false;
      }
    }
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
              onChange={(event) => {
                setAmount(event.target.value);
                setResult(null);
                setStatus('idle');
              }}
              error={Boolean(amountError)}
              disabled={effectiveStatus === 'submitting'}
            />
            {amountError ? (
              <p className="text-xs text-destructive">{amountError}</p>
            ) : (
              <p className="text-xs text-muted-foreground">
                {fiatEstimates.length > 0
                  ? fiatEstimates.join(' · ')
                  : 'Live fiat conversion unavailable'}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label>Destination</Label>
            <p className="rounded-md border border-border px-3 py-2 text-sm">Connected wallet</p>
            <p className="text-xs text-muted-foreground">
              Organization treasury withdrawals are unavailable until checkpoint 89C.
            </p>
          </div>

          {fiatOffRampUrl && (
            <div className="rounded-lg border border-border/60 bg-muted/40 p-3 text-xs">
              Fiat off-ramp navigation only. No withdrawal is requested here.{' '}
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

          {effectiveStatus === 'unavailable' && (
            <Alert>
              <AlertTitle>Wallet withdrawals unavailable</AlertTitle>
              <AlertDescription>
                No authoritative signing and broadcast adapter is configured for escrow withdrawals.
              </AlertDescription>
            </Alert>
          )}
          {effectiveStatus === 'submitting' && (
            <Alert>
              <AlertTitle>Submitting withdrawal</AlertTitle>
              <AlertDescription>
                Waiting for committed on-chain transaction evidence.
              </AlertDescription>
            </Alert>
          )}
          {effectiveStatus === 'error' && (
            <Alert variant="destructive">
              <AlertTitle>Withdrawal not committed</AlertTitle>
              <AlertDescription>
                The transaction failed, timed out, changed, or returned invalid commit evidence.
              </AlertDescription>
            </Alert>
          )}
          {effectiveStatus === 'committed' && result && (
            <Alert variant="success">
              <AlertTitle>Withdrawal committed</AlertTitle>
              <AlertDescription>
                Transaction {result.txHash} committed at height {result.blockHeight}.
              </AlertDescription>
            </Alert>
          )}

          <Button
            type="submit"
            disabled={
              Boolean(amountError) ||
              effectiveStatus === 'submitting' ||
              effectiveStatus === 'unavailable'
            }
          >
            {effectiveStatus === 'submitting' ? 'Submitting…' : 'Withdraw to wallet'}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import {
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Modal,
} from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
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

interface DepositModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  account: EscrowAccount;
  fiatRates: FiatRates;
  mutationAdapter?: EscrowMutationAdapter;
  mutationContext?: EscrowMutationContext;
  resultProjector?: EscrowMutationResultProjector;
  mutationTimeoutMs?: number;
}

const MIN_DEPOSIT = 50;

export function DepositModal({
  open,
  onOpenChange,
  account,
  fiatRates,
  mutationAdapter,
  mutationContext,
  resultProjector,
  mutationTimeoutMs,
}: DepositModalProps) {
  const [amount, setAmount] = useState('500');
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
  const currentRequest = open
    ? buildEscrowMutationRequest(
        'deposit',
        mutationContext ?? { chainId: '', accountAddress: '', escrowAccountId: '' },
        amount,
        account.currency
      )
    : null;
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
    open,
    mutationAdapter,
    mutationContext?.accountAddress,
    mutationContext?.chainId,
    mutationContext?.escrowAccountId,
    account.currency,
    resultProjector,
  ]);

  const numericAmount = useMemo(() => Number(amount), [amount]);
  const fiatEstimates = formatFiatEstimates(numericAmount, fiatRates);
  const amountError =
    !isValidEscrowMutationAmount(amount) || numericAmount < MIN_DEPOSIT
      ? `Minimum deposit is ${MIN_DEPOSIT} ${account.currency}`
      : account.walletBalance > 0 && numericAmount > account.walletBalance
        ? 'Amount exceeds available wallet balance'
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

  const handleOpenChange = (next: boolean) => {
    if (!next) {
      activeSubmission.current?.abort();
      activeSubmission.current = null;
      mutationSession.current += 1;
      inFlight.current = false;
      setResult(null);
      setStatus('idle');
    }
    onOpenChange(next);
  };

  return (
    <Modal open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Deposit to Escrow</DialogTitle>
          <DialogDescription>
            Transfer funds into escrow to secure settlement before releasing compute resources.
          </DialogDescription>
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
            <Label>Funding source</Label>
            <p className="rounded-md border border-border px-3 py-2 text-sm">
              Connected wallet
              {account.walletBalance > 0
                ? ` (${formatToken(account.walletBalance, account.currency)})`
                : ' (balance confirmed during signing)'}
            </p>
          </div>

          {effectiveStatus === 'unavailable' && (
            <Alert>
              <AlertTitle>Wallet deposits unavailable</AlertTitle>
              <AlertDescription>
                No authoritative signing and broadcast adapter is configured for escrow deposits.
              </AlertDescription>
            </Alert>
          )}
          {effectiveStatus === 'submitting' && (
            <Alert>
              <AlertTitle>Submitting deposit</AlertTitle>
              <AlertDescription>
                Waiting for committed on-chain transaction evidence.
              </AlertDescription>
            </Alert>
          )}
          {effectiveStatus === 'error' && (
            <Alert variant="destructive">
              <AlertTitle>Deposit not committed</AlertTitle>
              <AlertDescription>
                The transaction failed, timed out, changed, or returned invalid commit evidence.
              </AlertDescription>
            </Alert>
          )}
          {effectiveStatus === 'committed' && result && (
            <Alert variant="success">
              <AlertTitle>Deposit committed</AlertTitle>
              <AlertDescription>
                Transaction {result.txHash} committed at height {result.blockHeight}.
              </AlertDescription>
            </Alert>
          )}

          <DialogFooter>
            <Button variant="outline" type="button" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={
                Boolean(amountError) ||
                effectiveStatus === 'submitting' ||
                effectiveStatus === 'unavailable'
              }
            >
              {effectiveStatus === 'submitting' ? 'Submitting…' : 'Deposit'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Modal>
  );
}

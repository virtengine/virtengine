/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useMemo, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/Modal';
import { Badge } from '@/components/ui/Badge';
import { useToast } from '@/hooks/use-toast';
import { useWallet } from '@/lib/portal-adapter';
import { useWalletTransaction } from '@/hooks/useWalletTransaction';
import { getChainInfo } from '@/config/chains';
import { formatTokenAmount, truncateAddress } from '@/lib/utils';
import { useGovernanceDelegations, useValidatorVoteHistory } from '@/hooks/useGovernance';
import { formatProposalStatus } from '@/lib/governance';
import type { GovernanceProposal } from '@/types/governance';

interface DelegationPanelProps {
  address?: string;
  upcomingProposals?: GovernanceProposal[];
}

function toMinimalAmount(amount: string, decimals: number): string {
  const numeric = Number(amount);
  if (Number.isNaN(numeric) || numeric <= 0) return '0';
  return Math.round(numeric * Math.pow(10, decimals)).toString();
}

export function DelegationPanel({ address, upcomingProposals = [] }: DelegationPanelProps) {
  const chainInfo = getChainInfo();
  const wallet = useWallet();
  const { toast } = useToast();
  const { data, isLoading, error } = useGovernanceDelegations(address);
  const { sendTransaction, estimateFee, isLoading: isSubmitting } = useWalletTransaction();
  const [selectedValidator, setSelectedValidator] = useState('');
  const [amount, setAmount] = useState('');
  const [confirmOpen, setConfirmOpen] = useState(false);
  const { data: history, isLoading: isHistoryLoading } = useValidatorVoteHistory(
    selectedValidator || undefined
  );

  const validators = useMemo(() => data?.validators ?? [], [data?.validators]);
  const delegations = useMemo(() => data?.delegations ?? [], [data?.delegations]);
  const totalDelegated = useMemo(() => {
    return delegations.reduce((sum, delegation) => sum + BigInt(delegation.balance.amount), 0n);
  }, [delegations]);

  const validatorMap = useMemo(() => {
    const map = new Map<string, string>();
    validators.forEach((validator) => {
      const label = validator.description?.moniker || truncateAddress(validator.operator_address);
      map.set(validator.operator_address, label);
    });
    return map;
  }, [validators]);

  const canDelegate = wallet.status === 'connected' && selectedValidator && Number(amount) > 0;
  const feeEstimate = estimateFee(240000);

  const handleDelegate = async () => {
    try {
      await sendTransaction([
        {
          typeUrl: '/cosmos.staking.v1beta1.MsgDelegate',
          value: {
            delegatorAddress: wallet.accounts?.[wallet.activeAccountIndex ?? 0]?.address,
            validatorAddress: selectedValidator,
            amount: {
              denom: chainInfo.stakeCurrency.coinMinimalDenom,
              amount: toMinimalAmount(amount, chainInfo.stakeCurrency.coinDecimals),
            },
          },
        },
      ]);
      toast({
        title: 'Delegation submitted',
        description: 'Your delegation transaction has been signed.',
      });
      setConfirmOpen(false);
      setAmount('');
    } catch (err) {
      toast({
        title: 'Delegation failed',
        description: err instanceof Error ? err.message : 'Unable to delegate',
        variant: 'destructive',
      });
    }
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Delegation management</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="rounded-lg border border-border bg-muted/30 p-4">
              <div className="text-sm text-muted-foreground">Delegated voting power</div>
              <div className="mt-1 text-2xl font-semibold">
                {formatTokenAmount(totalDelegated.toString())} {chainInfo.stakeCurrency.coinDenom}
              </div>
              <div className="text-xs text-muted-foreground">
                Across {delegations.length} validators
              </div>
            </div>
            <div className="rounded-lg border border-border bg-muted/30 p-4">
              <div className="text-sm text-muted-foreground">Wallet voting power</div>
              <div className="mt-1 text-2xl font-semibold">
                {formatTokenAmount(wallet.balance ?? '0')} {chainInfo.stakeCurrency.coinDenom}
              </div>
              <div className="text-xs text-muted-foreground">Available balance</div>
            </div>
          </div>

          {error ? (
            <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-3 text-xs text-destructive">
              {error}
            </div>
          ) : null}

          <div className="space-y-2">
            <div className="text-sm font-medium">Active delegations</div>
            {isLoading ? (
              <div className="text-xs text-muted-foreground">Loading delegations…</div>
            ) : delegations.length === 0 ? (
              <div className="text-xs text-muted-foreground">
                No delegations found. Delegate to a validator to activate voting power.
              </div>
            ) : (
              <div className="space-y-2">
                {delegations.slice(0, 4).map((delegation) => (
                  <div
                    key={delegation.delegation.validator_address}
                    className="flex items-center justify-between rounded-lg border border-border bg-card px-3 py-2 text-sm"
                  >
                    <span>
                      {validatorMap.get(delegation.delegation.validator_address) ??
                        truncateAddress(delegation.delegation.validator_address)}
                    </span>
                    <span className="text-muted-foreground">
                      {formatTokenAmount(delegation.balance.amount)}{' '}
                      {chainInfo.stakeCurrency.coinDenom}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="space-y-2 rounded-lg border border-border bg-background p-4">
            <div className="text-sm font-medium">Delegate voting power</div>
            <div className="grid gap-3 sm:grid-cols-2">
              <Select value={selectedValidator} onValueChange={setSelectedValidator}>
                <SelectTrigger>
                  <SelectValue placeholder="Select validator" />
                </SelectTrigger>
                <SelectContent>
                  {validators.map((validator) => (
                    <SelectItem key={validator.operator_address} value={validator.operator_address}>
                      {validator.description?.moniker ??
                        truncateAddress(validator.operator_address)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Input
                type="number"
                min={0}
                placeholder={`Amount (${chainInfo.stakeCurrency.coinDenom})`}
                value={amount}
                onChange={(event) => setAmount(event.target.value)}
              />
            </div>
            <Button
              className="mt-2 w-full"
              onClick={() => setConfirmOpen(true)}
              disabled={!canDelegate}
            >
              Delegate
            </Button>
            <p className="text-xs text-muted-foreground">
              Delegations follow validator voting power. Changes may take one block to reflect.
            </p>
          </div>

          <div className="space-y-2 rounded-lg border border-border bg-background p-4">
            <div className="text-sm font-medium">Validator voting history</div>
            {selectedValidator === '' ? (
              <div className="text-xs text-muted-foreground">
                Select a validator to view their recent governance votes.
              </div>
            ) : isHistoryLoading ? (
              <div className="text-xs text-muted-foreground">Loading vote history…</div>
            ) : (
              <div className="space-y-2">
                {history?.proposals.map((entry) => (
                  <div
                    key={entry.proposalId}
                    className="flex items-center justify-between rounded-lg border border-border px-3 py-2 text-xs"
                  >
                    <div>
                      <div className="font-medium">#{entry.proposalId}</div>
                      <div className="text-muted-foreground">{entry.title ?? 'Proposal'}</div>
                    </div>
                    <Badge variant="outline">
                      {entry.vote?.options?.[0]?.option
                        ? entry.vote.options[0].option.replace('VOTE_OPTION_', '')
                        : 'No vote'}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Upcoming proposal alerts</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {upcomingProposals.length === 0 ? (
            <div className="text-sm text-muted-foreground">No upcoming proposals right now.</div>
          ) : (
            upcomingProposals.slice(0, 4).map((proposal) => (
              <div
                key={proposal.id}
                className="flex items-center justify-between rounded-lg border border-border bg-card px-3 py-2 text-sm"
              >
                <div>
                  <div className="font-medium">#{proposal.id}</div>
                  <div className="text-xs text-muted-foreground">
                    {proposal.title ?? 'Proposal'}
                  </div>
                </div>
                <Badge variant="outline">{formatProposalStatus(proposal.status)}</Badge>
              </div>
            ))
          )}
        </CardContent>
      </Card>

      <Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm delegation</DialogTitle>
            <DialogDescription>
              Delegate {amount || 0} {chainInfo.stakeCurrency.coinDenom} to{' '}
              {validatorMap.get(selectedValidator) ?? selectedValidator}
            </DialogDescription>
          </DialogHeader>
          <div className="rounded-lg border border-border bg-muted/40 p-4 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Estimated fee</span>
              <span className="font-medium">
                {feeEstimate.amount?.[0]?.amount ?? '0'} {feeEstimate.amount?.[0]?.denom ?? ''}
              </span>
            </div>
          </div>
          <div className="mt-4 flex flex-col gap-3 sm:flex-row">
            <Button variant="outline" className="flex-1" onClick={() => setConfirmOpen(false)}>
              Cancel
            </Button>
            <Button className="flex-1" onClick={handleDelegate} loading={isSubmitting}>
              Confirm delegation
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

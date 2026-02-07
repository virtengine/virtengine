/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useMemo, useState } from 'react';
import { Button } from '@/components/ui/Button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/Modal';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import { useToast } from '@/hooks/use-toast';
import { useWallet } from '@/lib/portal-adapter';
import { useWalletTransaction } from '@/hooks/useWalletTransaction';
import { WeightedVoteSlider } from '@/components/governance/WeightedVoteSlider';
import { formatVoteOption, formatTokenBalance } from '@/lib/governance';
import { getChainInfo } from '@/config/chains';
import type { VoteOption, WeightedVoteOption } from '@/types/governance';

interface VoteModalProps {
  proposalId: string;
  open: boolean;
  onClose: () => void;
}

const VOTE_OPTIONS: VoteOption[] = [
  'VOTE_OPTION_YES',
  'VOTE_OPTION_NO',
  'VOTE_OPTION_NO_WITH_VETO',
  'VOTE_OPTION_ABSTAIN',
];

const DEFAULT_GAS_LIMIT = 220000;

export function VoteModal({ proposalId, open, onClose }: VoteModalProps) {
  const [mode, setMode] = useState<'simple' | 'weighted'>('simple');
  const [simpleVote, setSimpleVote] = useState<VoteOption | null>(null);
  const [weights, setWeights] = useState({
    yes: 0,
    no: 0,
    veto: 0,
    abstain: 0,
  });
  const [confirmOpen, setConfirmOpen] = useState(false);
  const { toast } = useToast();
  const wallet = useWallet();
  const { estimateFee, sendTransaction, isLoading } = useWalletTransaction();
  const chainInfo = getChainInfo();
  const voterAddress = wallet.accounts?.[wallet.activeAccountIndex ?? 0]?.address;

  const totalWeight = weights.yes + weights.no + weights.veto + weights.abstain;
  const feeEstimate = estimateFee(DEFAULT_GAS_LIMIT);

  const votePreview = useMemo(() => {
    if (mode === 'simple' && simpleVote) {
      return formatVoteOption(simpleVote);
    }
    if (mode === 'weighted') {
      return `Weighted (${totalWeight}%)`;
    }
    return 'Select a vote option';
  }, [mode, simpleVote, totalWeight]);

  const buildVoteMessage = () => {
    if (mode === 'simple' && simpleVote) {
      return [
        {
          typeUrl: '/cosmos.gov.v1.MsgVote',
          value: {
            proposalId,
            voter: voterAddress,
            option: simpleVote,
          },
        },
      ];
    }

    const options: WeightedVoteOption[] = [
      { option: 'VOTE_OPTION_YES' as VoteOption, weight: (weights.yes / 100).toString() },
      { option: 'VOTE_OPTION_NO' as VoteOption, weight: (weights.no / 100).toString() },
      {
        option: 'VOTE_OPTION_NO_WITH_VETO' as VoteOption,
        weight: (weights.veto / 100).toString(),
      },
      { option: 'VOTE_OPTION_ABSTAIN' as VoteOption, weight: (weights.abstain / 100).toString() },
    ].filter((entry) => parseFloat(entry.weight) > 0);

    return [
      {
        typeUrl: '/cosmos.gov.v1.MsgVoteWeighted',
        value: {
          proposalId,
          voter: voterAddress,
          options,
        },
      },
    ];
  };

  const handleConfirm = async () => {
    try {
      await sendTransaction(buildVoteMessage());
      toast({
        title: 'Vote submitted',
        description: 'Your vote has been signed and queued for broadcast.',
      });
      setConfirmOpen(false);
      onClose();
    } catch (err) {
      toast({
        title: 'Vote failed',
        description: err instanceof Error ? err.message : 'Unable to submit vote',
        variant: 'destructive',
      });
    }
  };

  const canSubmit =
    wallet.status === 'connected' &&
    Boolean(voterAddress) &&
    ((mode === 'simple' && simpleVote !== null) || (mode === 'weighted' && totalWeight === 100));

  const votingPower = formatTokenBalance(
    wallet.balance ?? undefined,
    chainInfo.stakeCurrency.coinDecimals
  );

  return (
    <>
      <Dialog open={open} onOpenChange={(value) => (!value ? onClose() : null)}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle>Cast your vote</DialogTitle>
            <DialogDescription>
              Choose how you want to vote on proposal #{proposalId}. Voting power follows your
              active delegation.
            </DialogDescription>
          </DialogHeader>

          <Tabs value={mode} onValueChange={(value) => setMode(value as 'simple' | 'weighted')}>
            <TabsList className="w-full">
              <TabsTrigger value="simple" className="flex-1">
                Simple vote
              </TabsTrigger>
              <TabsTrigger value="weighted" className="flex-1">
                Weighted vote
              </TabsTrigger>
            </TabsList>

            <TabsContent value="simple">
              <div className="mt-4 grid grid-cols-2 gap-3">
                {VOTE_OPTIONS.map((option) => (
                  <Button
                    key={option}
                    variant={simpleVote === option ? 'default' : 'outline'}
                    onClick={() => setSimpleVote(option)}
                  >
                    {formatVoteOption(option)}
                  </Button>
                ))}
              </div>
            </TabsContent>

            <TabsContent value="weighted">
              <div className="mt-4 space-y-3">
                <WeightedVoteSlider
                  label="Yes"
                  value={weights.yes}
                  onChange={(value) => setWeights((prev) => ({ ...prev, yes: value }))}
                  colorClass="bg-success"
                />
                <WeightedVoteSlider
                  label="No"
                  value={weights.no}
                  onChange={(value) => setWeights((prev) => ({ ...prev, no: value }))}
                  colorClass="bg-destructive"
                />
                <WeightedVoteSlider
                  label="No with Veto"
                  value={weights.veto}
                  onChange={(value) => setWeights((prev) => ({ ...prev, veto: value }))}
                  colorClass="bg-orange-500"
                />
                <WeightedVoteSlider
                  label="Abstain"
                  value={weights.abstain}
                  onChange={(value) => setWeights((prev) => ({ ...prev, abstain: value }))}
                  colorClass="bg-slate-400"
                />
                <p className="text-xs text-muted-foreground">
                  Total weight: {totalWeight}% {totalWeight !== 100 ? '(must equal 100%)' : ''}
                </p>
              </div>
            </TabsContent>
          </Tabs>

          <div className="mt-5 space-y-2 rounded-lg border border-border bg-muted/30 p-4 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Voting power</span>
              <span className="font-medium">
                {votingPower} {chainInfo.stakeCurrency.coinDenom}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Estimated fee</span>
              <span className="font-medium">
                {feeEstimate.amount?.[0]?.amount ?? '0'} {feeEstimate.amount?.[0]?.denom ?? ''}
              </span>
            </div>
          </div>

          <Button
            className="mt-6 w-full"
            onClick={() => setConfirmOpen(true)}
            disabled={!canSubmit}
          >
            Review vote
          </Button>
        </DialogContent>
      </Dialog>

      <Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm your vote</DialogTitle>
            <DialogDescription>
              Proposal #{proposalId} · {votePreview}
            </DialogDescription>
          </DialogHeader>
          <div className="rounded-lg border border-border bg-muted/40 p-4 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Gas limit</span>
              <span className="font-medium">{feeEstimate.gas}</span>
            </div>
            <div className="mt-2 flex items-center justify-between">
              <span className="text-muted-foreground">Estimated fee</span>
              <span className="font-medium">
                {feeEstimate.amount?.[0]?.amount ?? '0'} {feeEstimate.amount?.[0]?.denom ?? ''}
              </span>
            </div>
          </div>
          <div className="mt-4 flex flex-col gap-3 sm:flex-row">
            <Button variant="outline" className="flex-1" onClick={() => setConfirmOpen(false)}>
              Back
            </Button>
            <Button className="flex-1" onClick={handleConfirm} loading={isLoading}>
              Submit vote
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

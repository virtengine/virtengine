/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { calculateTallyPercentages, getTallyTotal } from '@/lib/governance';
import { formatTokenAmount } from '@/lib/utils';
import type { TallyResult } from '@/types/governance';

interface VoteTallyProps {
  tally?: TallyResult | null;
  bondedTokens?: string;
}

export function VoteTally({ tally, bondedTokens }: VoteTallyProps) {
  const percentages = calculateTallyPercentages(tally ?? undefined);
  const totalVotes = getTallyTotal(tally ?? undefined);

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between text-sm">
        <span className="text-success">Yes {percentages.yes}%</span>
        <span className="text-destructive">No {percentages.no}%</span>
        <span className="text-muted-foreground">Abstain {percentages.abstain}%</span>
        <span className="text-orange-500">Veto {percentages.veto}%</span>
      </div>
      <div className="flex h-2 overflow-hidden rounded-full bg-muted">
        <div className="bg-success" style={{ width: `${percentages.yes}%` }} />
        <div className="bg-destructive" style={{ width: `${percentages.no}%` }} />
        <div className="bg-slate-400" style={{ width: `${percentages.abstain}%` }} />
        <div className="bg-orange-500" style={{ width: `${percentages.veto}%` }} />
      </div>
      <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
        <span>Total votes: {formatTokenAmount(totalVotes.toString())} VE</span>
        {bondedTokens ? <span>Bonded supply: {formatTokenAmount(bondedTokens)} VE</span> : null}
      </div>
    </div>
  );
}

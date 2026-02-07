/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Progress } from '@/components/ui/Progress';
import { calculateQuorumProgress, formatQuorumTarget } from '@/lib/governance';
import type { TallyParams, TallyResult } from '@/types/governance';

interface QuorumProgressProps {
  tally?: TallyResult | null;
  bondedTokens?: string;
  tallyParams?: TallyParams;
}

export function QuorumProgress({ tally, bondedTokens, tallyParams }: QuorumProgressProps) {
  const quorum = calculateQuorumProgress(tally ?? undefined, bondedTokens);
  const quorumPercent = Math.round(quorum * 1000) / 10;

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between text-xs text-muted-foreground">
        <span>{formatQuorumTarget(tallyParams)}</span>
        <span>{quorumPercent}% quorum reached</span>
      </div>
      <Progress value={quorumPercent} />
    </div>
  );
}

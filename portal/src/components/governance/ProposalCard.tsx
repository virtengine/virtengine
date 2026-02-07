/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import Link from 'next/link';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent } from '@/components/ui/Card';
import { QuorumProgress } from '@/components/governance/QuorumProgress';
import { VoteTally } from '@/components/governance/VoteTally';
import {
  formatProposalStatus,
  getProposalStatusStyles,
  getProposalSummary,
  getProposalTitle,
  getProposalType,
  getVotingTimeRemaining,
} from '@/lib/governance';
import { formatDateTime } from '@/lib/utils';
import type { GovernanceProposalWithTally, TallyParams } from '@/types/governance';

interface ProposalCardProps {
  proposal: GovernanceProposalWithTally;
  tallyParams?: TallyParams;
  bondedTokens?: string;
}

export function ProposalCard({ proposal, tallyParams, bondedTokens }: ProposalCardProps) {
  const statusLabel = formatProposalStatus(proposal.status);
  const statusStyles = getProposalStatusStyles(proposal.status);
  const typeLabel = getProposalType(proposal.messages);
  const endTime = proposal.voting_end_time;
  const timeRemaining = endTime ? getVotingTimeRemaining(endTime) : 'N/A';
  const endLabel =
    proposal.status === 'PROPOSAL_STATUS_VOTING_PERIOD'
      ? `Ends ${timeRemaining}`
      : endTime
        ? `Ended ${formatDateTime(endTime)}`
        : 'Timeline unavailable';

  return (
    <Card className="card-hover transition-all">
      <CardContent className="p-6">
        <Link href={`/governance/proposals/${proposal.id}`} className="block">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex flex-wrap items-center gap-3">
              <span className="font-mono text-xs text-muted-foreground">#{proposal.id}</span>
              <Badge className={`${statusStyles.bg} ${statusStyles.text}`}>{statusLabel}</Badge>
              <Badge variant="outline" className="text-xs">
                {typeLabel}
              </Badge>
            </div>
            <span className="text-xs text-muted-foreground">{endLabel}</span>
          </div>

          <div className="mt-3">
            <h3 className="text-lg font-semibold">{getProposalTitle(proposal)}</h3>
            <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">
              {getProposalSummary(proposal)}
            </p>
          </div>

          <div className="mt-4 space-y-4">
            <VoteTally tally={proposal.tally} bondedTokens={bondedTokens} />
            <QuorumProgress
              tally={proposal.tally}
              bondedTokens={bondedTokens}
              tallyParams={tallyParams}
            />
          </div>
        </Link>
      </CardContent>
    </Card>
  );
}

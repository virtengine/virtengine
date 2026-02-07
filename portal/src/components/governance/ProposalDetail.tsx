/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import Link from 'next/link';
import { useMemo, useState } from 'react';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { ProposalMarkdown } from '@/components/governance/ProposalMarkdown';
import { VoteModal } from '@/components/governance/VoteModal';
import { VoteTally } from '@/components/governance/VoteTally';
import { QuorumProgress } from '@/components/governance/QuorumProgress';
import {
  formatProposalStatus,
  getProposalStatusStyles,
  getProposalBody,
  getProposalForumUrl,
  getProposalTitle,
  getProposalType,
} from '@/lib/governance';
import { formatDateTime, truncateAddress } from '@/lib/utils';
import type {
  GovernanceProposal,
  GovernanceVote,
  TallyParams,
  TallyResult,
  VotingParams,
} from '@/types/governance';

interface ProposalDetailProps {
  proposal: GovernanceProposal;
  tally?: TallyResult | null;
  votes?: GovernanceVote[];
  tallyParams?: TallyParams;
  votingParams?: VotingParams;
  bondedTokens?: string;
  relatedProposals?: GovernanceProposal[];
  voterVote?: GovernanceVote | null;
}

export function ProposalDetail({
  proposal,
  tally,
  votes,
  tallyParams,
  votingParams,
  bondedTokens,
  relatedProposals,
  voterVote,
}: ProposalDetailProps) {
  const [voteOpen, setVoteOpen] = useState(false);
  const statusLabel = formatProposalStatus(proposal.status);
  const statusStyles = getProposalStatusStyles(proposal.status);
  const typeLabel = getProposalType(proposal.messages);
  const forumUrl = getProposalForumUrl(proposal);

  const voteSummary = useMemo(() => {
    if (!voterVote?.options?.length) return null;
    return voterVote.options
      .map(
        (option) =>
          `${option.option.replace('VOTE_OPTION_', '')} ${(Number(option.weight) * 100).toFixed(0)}%`
      )
      .join(', ');
  }, [voterVote]);

  return (
    <div className="container py-8">
      <div className="mb-6">
        <Link
          href="/governance/proposals"
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to Proposals
        </Link>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <Card>
            <CardContent className="p-6">
              <div className="flex flex-wrap items-center gap-3">
                <span className="font-mono text-xs text-muted-foreground">#{proposal.id}</span>
                <Badge className={`${statusStyles.bg} ${statusStyles.text}`}>{statusLabel}</Badge>
                <Badge variant="outline">{typeLabel}</Badge>
              </div>

              <h1 className="mt-4 text-2xl font-bold">{getProposalTitle(proposal)}</h1>

              <div className="mt-4 flex flex-wrap gap-4 text-sm text-muted-foreground">
                <span>
                  Proposer: {proposal.proposer ? truncateAddress(proposal.proposer) : 'Unknown'}
                </span>
                <span>•</span>
                <span>
                  Submitted: {proposal.submit_time ? formatDateTime(proposal.submit_time) : 'N/A'}
                </span>
                <span>•</span>
                <span>
                  Voting ends:{' '}
                  {proposal.voting_end_time ? formatDateTime(proposal.voting_end_time) : 'N/A'}
                </span>
              </div>

              <div className="mt-6">
                <ProposalMarkdown content={getProposalBody(proposal)} />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Discussion</CardTitle>
            </CardHeader>
            <CardContent>
              {forumUrl ? (
                <a
                  href={forumUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-primary hover:underline"
                >
                  View on forum →
                </a>
              ) : (
                <p className="text-sm text-muted-foreground">
                  No official discussion link has been provided for this proposal.
                </p>
              )}
            </CardContent>
          </Card>

          {relatedProposals && relatedProposals.length > 0 ? (
            <Card>
              <CardHeader>
                <CardTitle>More from this proposer</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                {relatedProposals.map((item) => (
                  <Link
                    key={item.id}
                    href={`/governance/proposals/${item.id}`}
                    className="flex items-center justify-between rounded-lg border border-border bg-card px-3 py-2 text-sm hover:bg-accent"
                  >
                    <span className="font-medium">#{item.id}</span>
                    <span className="text-muted-foreground">{item.title ?? 'Proposal'}</span>
                  </Link>
                ))}
              </CardContent>
            </Card>
          ) : null}
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Cast your vote</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {proposal.status === 'PROPOSAL_STATUS_VOTING_PERIOD' ? (
                <>
                  <Button className="w-full" onClick={() => setVoteOpen(true)}>
                    Vote on proposal
                  </Button>
                  {voterVote ? (
                    <div className="rounded-lg border border-border bg-muted/40 p-3 text-xs text-muted-foreground">
                      Your current vote:{' '}
                      <span className="font-medium text-foreground">{voteSummary}</span>
                    </div>
                  ) : (
                    <div className="text-xs text-muted-foreground">
                      You have not voted on this proposal yet.
                    </div>
                  )}
                </>
              ) : (
                <div className="text-sm text-muted-foreground">
                  Voting is closed for this proposal.
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Current results</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <VoteTally tally={tally ?? undefined} bondedTokens={bondedTokens} />
              <QuorumProgress
                tally={tally ?? undefined}
                bondedTokens={bondedTokens}
                tallyParams={tallyParams}
              />
              <div className="space-y-2 text-xs text-muted-foreground">
                <div className="flex items-center justify-between">
                  <span>Threshold</span>
                  <span>
                    {tallyParams?.threshold
                      ? `${Math.round(Number(tallyParams.threshold) * 100)}%`
                      : 'N/A'}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span>Veto threshold</span>
                  <span>
                    {tallyParams?.veto_threshold
                      ? `${Math.round(Number(tallyParams.veto_threshold) * 100)}%`
                      : 'N/A'}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span>Voting period</span>
                  <span>{votingParams?.voting_period ?? 'N/A'}</span>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Timeline</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4 text-sm">
              <TimelineItem
                label="Proposal submitted"
                date={proposal.submit_time}
                status="completed"
              />
              <TimelineItem
                label="Deposit period ends"
                date={proposal.deposit_end_time}
                status={
                  proposal.status === 'PROPOSAL_STATUS_DEPOSIT_PERIOD' ? 'upcoming' : 'completed'
                }
              />
              <TimelineItem
                label="Voting starts"
                date={proposal.voting_start_time}
                status={
                  proposal.status === 'PROPOSAL_STATUS_DEPOSIT_PERIOD' ? 'upcoming' : 'completed'
                }
              />
              <TimelineItem
                label="Voting ends"
                date={proposal.voting_end_time}
                status={
                  proposal.status === 'PROPOSAL_STATUS_VOTING_PERIOD' ? 'upcoming' : 'completed'
                }
              />
            </CardContent>
          </Card>

          {votes && votes.length > 0 ? (
            <Card>
              <CardHeader>
                <CardTitle>Recent votes</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-xs">
                {votes.slice(0, 5).map((vote) => (
                  <div
                    key={vote.voter}
                    className="flex items-center justify-between rounded-lg border border-border px-3 py-2"
                  >
                    <span className="font-mono">{truncateAddress(vote.voter)}</span>
                    <Badge variant="outline">
                      {vote.options?.[0]?.option?.replace('VOTE_OPTION_', '') ?? 'Unknown'}
                    </Badge>
                  </div>
                ))}
              </CardContent>
            </Card>
          ) : null}
        </div>
      </div>

      <VoteModal proposalId={proposal.id} open={voteOpen} onClose={() => setVoteOpen(false)} />
    </div>
  );
}

function TimelineItem({
  label,
  date,
  status,
}: {
  label: string;
  date?: string;
  status: 'completed' | 'upcoming';
}) {
  return (
    <div className="flex gap-3">
      <div className="relative flex flex-col items-center">
        <span
          className={`h-2.5 w-2.5 rounded-full ${
            status === 'completed' ? 'bg-success' : 'bg-muted'
          }`}
        />
        <span className="w-px flex-1 bg-border" />
      </div>
      <div className="pb-3">
        <div className="text-sm font-medium">{label}</div>
        <div className="text-xs text-muted-foreground">{date ? formatDateTime(date) : 'N/A'}</div>
      </div>
    </div>
  );
}

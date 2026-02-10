'use client';

import { useParams } from 'next/navigation';
import { ProposalDetail } from '@/components/governance/ProposalDetail';
import { ProposalListSkeleton } from '@/components/governance/ProposalListSkeleton';
import { useGovernanceProposalDetail } from '@/hooks/useGovernance';
import { useWallet } from '@/lib/portal-adapter';

export default function ProposalDetailClient() {
  const params = useParams();
  const id = params.id as string;
  const wallet = useWallet();
  const voter = wallet.accounts?.[wallet.activeAccountIndex ?? 0]?.address;
  const { data, isLoading, error } = useGovernanceProposalDetail(id, voter);

  if (isLoading || !data?.proposal) {
    return (
      <div className="container py-8">
        {error ? (
          <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-sm text-destructive">
            {error}
          </div>
        ) : (
          <ProposalListSkeleton count={2} />
        )}
      </div>
    );
  }

  return (
    <ProposalDetail
      proposal={data.proposal}
      tally={data.tally ?? undefined}
      votes={data.votes}
      tallyParams={data.tallyParams}
      votingParams={data.votingParams}
      bondedTokens={data.bondedTokens}
      relatedProposals={data.relatedProposals}
      voterVote={data.voterVote ?? null}
    />
  );
}

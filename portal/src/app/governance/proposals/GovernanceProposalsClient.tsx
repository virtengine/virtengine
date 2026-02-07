/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useMemo, useState } from 'react';
import { ProposalFilters } from '@/components/governance/ProposalFilters';
import { ProposalList } from '@/components/governance/ProposalList';
import { DelegationPanel } from '@/components/governance/DelegationPanel';
import { Card, CardContent } from '@/components/ui/Card';
import { useGovernanceProposals } from '@/hooks/useGovernance';
import { useWallet } from '@/lib/portal-adapter';
import { formatTokenBalance } from '@/lib/governance';
import { getChainInfo } from '@/config/chains';
import type { GovernanceProposal } from '@/types/governance';

export function GovernanceProposalsClient() {
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState('PROPOSAL_STATUS_VOTING_PERIOD');
  const [typeFilter, setTypeFilter] = useState('all');
  const wallet = useWallet();
  const chainInfo = getChainInfo();

  const {
    proposals,
    tallyParams,
    bondedTokens,
    isLoading,
    isLoadingMore,
    error,
    hasMore,
    loadMore,
  } = useGovernanceProposals({ status });

  const activeCount = proposals.filter(
    (proposal) => proposal.status === 'PROPOSAL_STATUS_VOTING_PERIOD'
  ).length;

  const quorumTarget = tallyParams?.quorum
    ? `${Math.round(Number(tallyParams.quorum) * 100)}%`
    : 'N/A';

  const votingPower = formatTokenBalance(
    wallet.balance ?? undefined,
    chainInfo.stakeCurrency.coinDecimals
  );

  const upcomingProposals = useMemo<GovernanceProposal[]>(() => {
    const now = Date.now();
    return proposals.filter((proposal) => {
      if (proposal.status === 'PROPOSAL_STATUS_DEPOSIT_PERIOD') return true;
      if (proposal.voting_end_time) {
        const endTime = new Date(proposal.voting_end_time).getTime();
        return endTime - now < 1000 * 60 * 60 * 48;
      }
      return false;
    });
  }, [proposals]);

  return (
    <div className="container py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Governance</h1>
        <p className="mt-1 text-muted-foreground">View and vote on protocol governance proposals</p>
      </div>

      <div className="mb-8 grid gap-4 sm:grid-cols-3">
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Active proposals</div>
            <div className="mt-1 text-2xl font-bold">{activeCount}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Your voting power</div>
            <div className="mt-1 text-2xl font-bold">
              {votingPower} {chainInfo.stakeCurrency.coinDenom}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Quorum target</div>
            <div className="mt-1 text-2xl font-bold">{quorumTarget}</div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <ProposalFilters
            search={search}
            status={status}
            type={typeFilter}
            onSearchChange={setSearch}
            onStatusChange={setStatus}
            onTypeChange={setTypeFilter}
            onReset={() => {
              setSearch('');
              setStatus('PROPOSAL_STATUS_VOTING_PERIOD');
              setTypeFilter('all');
            }}
          />

          <ProposalList
            proposals={proposals}
            tallyParams={tallyParams}
            bondedTokens={bondedTokens}
            isLoading={isLoading}
            isLoadingMore={isLoadingMore}
            error={error}
            hasMore={hasMore}
            onLoadMore={loadMore}
            searchQuery={search}
            typeFilter={typeFilter}
          />
        </div>

        <DelegationPanel
          address={wallet.accounts?.[wallet.activeAccountIndex ?? 0]?.address}
          upcomingProposals={upcomingProposals}
        />
      </div>
    </div>
  );
}

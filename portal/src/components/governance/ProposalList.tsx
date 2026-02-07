/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, useMemo, useRef } from 'react';
import { ProposalCard } from '@/components/governance/ProposalCard';
import { ProposalListSkeleton } from '@/components/governance/ProposalListSkeleton';
import { Button } from '@/components/ui/Button';
import type { GovernanceProposalWithTally, TallyParams } from '@/types/governance';
import { getProposalSummary, getProposalTitle, getProposalType } from '@/lib/governance';

interface ProposalListProps {
  proposals: GovernanceProposalWithTally[];
  tallyParams?: TallyParams;
  bondedTokens?: string;
  isLoading: boolean;
  isLoadingMore: boolean;
  error?: string | null;
  hasMore: boolean;
  onLoadMore: () => void;
  searchQuery: string;
  typeFilter: string;
}

export function ProposalList({
  proposals,
  tallyParams,
  bondedTokens,
  isLoading,
  isLoadingMore,
  error,
  hasMore,
  onLoadMore,
  searchQuery,
  typeFilter,
}: ProposalListProps) {
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!hasMore || isLoadingMore) return;
    const target = sentinelRef.current;
    if (!target) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (entry.isIntersecting) {
          onLoadMore();
        }
      },
      { rootMargin: '200px' }
    );

    observer.observe(target);
    return () => observer.disconnect();
  }, [hasMore, isLoadingMore, onLoadMore]);

  const filteredProposals = useMemo(() => {
    const normalizedSearch = searchQuery.trim().toLowerCase();
    return proposals.filter((proposal) => {
      const matchesSearch =
        normalizedSearch.length === 0 ||
        getProposalTitle(proposal).toLowerCase().includes(normalizedSearch) ||
        getProposalSummary(proposal).toLowerCase().includes(normalizedSearch) ||
        proposal.id.includes(normalizedSearch);

      const typeLabel = getProposalType(proposal.messages);
      const matchesType = typeFilter === 'all' || typeLabel === typeFilter;

      return matchesSearch && matchesType;
    });
  }, [proposals, searchQuery, typeFilter]);

  if (isLoading) {
    return <ProposalListSkeleton />;
  }

  if (error) {
    return (
      <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-sm text-destructive">
        {error}
      </div>
    );
  }

  if (filteredProposals.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-card p-6 text-sm text-muted-foreground">
        No proposals match these filters yet. Try widening the search or checking another status.
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {filteredProposals.map((proposal) => (
        <ProposalCard
          key={proposal.id}
          proposal={proposal}
          tallyParams={tallyParams}
          bondedTokens={bondedTokens}
        />
      ))}
      <div ref={sentinelRef} />
      {hasMore && (
        <Button variant="outline" className="w-full" onClick={onLoadMore} loading={isLoadingMore}>
          {isLoadingMore ? 'Loading more proposals...' : 'Load more proposals'}
        </Button>
      )}
    </div>
  );
}

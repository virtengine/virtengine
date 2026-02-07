/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type {
  GovernanceDelegationsResponse,
  GovernanceProposalDetailResponse,
  GovernanceProposalsResponse,
  GovernanceProposalWithTally,
  ValidatorVoteHistoryResponse,
} from '@/types/governance';

const DEFAULT_PAGE_SIZE = 12;

function buildQuery(params: Record<string, string | number | boolean | undefined>) {
  const searchParams = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== '') {
      searchParams.set(key, String(value));
    }
  });
  return searchParams.toString();
}

export interface GovernanceProposalFilters {
  status?: string;
  limit?: number;
}

export function useGovernanceProposals(filters: GovernanceProposalFilters = {}) {
  const { status, limit = DEFAULT_PAGE_SIZE } = filters;
  const [proposals, setProposals] = useState<GovernanceProposalWithTally[]>([]);
  const [pagination, setPagination] = useState<GovernanceProposalsResponse['pagination']>();
  const [tallyParams, setTallyParams] = useState<GovernanceProposalsResponse['tallyParams']>();
  const [bondedTokens, setBondedTokens] = useState<string | undefined>();
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const hasMore = useMemo(() => {
    if (!pagination?.total) return Boolean(pagination?.next_key);
    const total = parseInt(pagination.total, 10);
    return Number.isNaN(total) ? Boolean(pagination?.next_key) : proposals.length < total;
  }, [pagination, proposals.length]);

  const activeRequest = useRef<AbortController | null>(null);

  const fetchPage = useCallback(
    async (nextPage: number, replace = false) => {
      if (activeRequest.current) {
        activeRequest.current.abort();
      }
      const controller = new AbortController();
      activeRequest.current = controller;
      const isFirstPage = nextPage === 1;
      setError(null);
      if (isFirstPage) {
        setIsLoading(true);
      } else {
        setIsLoadingMore(true);
      }

      try {
        const query = buildQuery({
          status,
          page: nextPage,
          limit,
          includeTally: true,
        });
        const response = await fetch(`/api/governance/proposals?${query}`, {
          signal: controller.signal,
        });
        if (!response.ok) {
          throw new Error(`Failed to load proposals (${response.status})`);
        }
        const data = (await response.json()) as GovernanceProposalsResponse;
        setProposals((prev) => (replace ? data.proposals : [...prev, ...data.proposals]));
        setPagination(data.pagination);
        setTallyParams(data.tallyParams);
        setBondedTokens(data.bondedTokens);
      } catch (err) {
        if ((err as { name?: string }).name !== 'AbortError') {
          setError(err instanceof Error ? err.message : 'Failed to load proposals');
        }
      } finally {
        setIsLoading(false);
        setIsLoadingMore(false);
      }
    },
    [limit, status]
  );

  useEffect(() => {
    setPage(1);
    void fetchPage(1, true);
    return () => {
      activeRequest.current?.abort();
    };
  }, [fetchPage, status, limit]);

  const loadMore = useCallback(() => {
    if (isLoadingMore || !hasMore) return;
    const nextPage = page + 1;
    setPage(nextPage);
    void fetchPage(nextPage, false);
  }, [fetchPage, hasMore, isLoadingMore, page]);

  const refresh = useCallback(() => {
    setPage(1);
    void fetchPage(1, true);
  }, [fetchPage]);

  return {
    proposals,
    pagination,
    tallyParams,
    bondedTokens,
    isLoading,
    isLoadingMore,
    error,
    hasMore,
    loadMore,
    refresh,
  };
}

export function useGovernanceProposalDetail(proposalId?: string, voter?: string) {
  const [data, setData] = useState<GovernanceProposalDetailResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!proposalId) return;
    const controller = new AbortController();
    const query = buildQuery({ voter });
    setIsLoading(true);
    setError(null);
    fetch(`/api/governance/proposals/${proposalId}${query ? `?${query}` : ''}`, {
      signal: controller.signal,
    })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error(`Failed to load proposal (${response.status})`);
        }
        const payload = (await response.json()) as GovernanceProposalDetailResponse;
        setData(payload);
      })
      .catch((err) => {
        if ((err as { name?: string }).name !== 'AbortError') {
          setError(err instanceof Error ? err.message : 'Failed to load proposal');
        }
      })
      .finally(() => {
        setIsLoading(false);
      });

    return () => controller.abort();
  }, [proposalId, voter]);

  return { data, isLoading, error };
}

export function useGovernanceDelegations(address?: string) {
  const [data, setData] = useState<GovernanceDelegationsResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!address) return;
    const controller = new AbortController();
    setIsLoading(true);
    setError(null);
    const query = buildQuery({ address });

    fetch(`/api/governance/delegations?${query}`, { signal: controller.signal })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error(`Failed to load delegations (${response.status})`);
        }
        const payload = (await response.json()) as GovernanceDelegationsResponse;
        setData(payload);
      })
      .catch((err) => {
        if ((err as { name?: string }).name !== 'AbortError') {
          setError(err instanceof Error ? err.message : 'Failed to load delegations');
        }
      })
      .finally(() => setIsLoading(false));

    return () => controller.abort();
  }, [address]);

  return { data, isLoading, error };
}

export function useValidatorVoteHistory(validator?: string) {
  const [data, setData] = useState<ValidatorVoteHistoryResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!validator) return;
    const controller = new AbortController();
    setIsLoading(true);
    setError(null);
    const query = buildQuery({ validator });

    fetch(`/api/governance/validators/votes?${query}`, { signal: controller.signal })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error(`Failed to load validator history (${response.status})`);
        }
        const payload = (await response.json()) as ValidatorVoteHistoryResponse;
        setData(payload);
      })
      .catch((err) => {
        if ((err as { name?: string }).name !== 'AbortError') {
          setError(err instanceof Error ? err.message : 'Failed to load validator history');
        }
      })
      .finally(() => setIsLoading(false));

    return () => controller.abort();
  }, [validator]);

  return { data, isLoading, error };
}

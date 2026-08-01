import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { HPCClientProvider } from '../context/HPCClientProvider';
import {
  createHPCClient,
  HPCClientUnavailableError,
  type HPCQueryAdapter,
  type HPCSignerAdapter,
} from '../lib/hpc-client';
import { useJobCancellation, useJobs, useJobSubmission, useWorkloadTemplates } from './index';

function createQueryAdapter(): HPCQueryAdapter {
  return {
    listWorkloadTemplates: vi.fn().mockResolvedValue([]),
    getWorkloadTemplate: vi.fn().mockResolvedValue(null),
    listOfferings: vi.fn().mockResolvedValue([]),
    getOffering: vi.fn().mockResolvedValue(null),
    listJobs: vi.fn().mockResolvedValue([]),
    getJob: vi.fn().mockResolvedValue(null),
    estimateJobCost: vi.fn().mockResolvedValue({
      estimatedTotal: '0',
      pricePerHour: '0',
      breakdown: { compute: '0', storage: '0', network: '0' },
      denom: 'uve',
    }),
  };
}

function providerWrapper(client: ReturnType<typeof createHPCClient>) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <HPCClientProvider client={client}>{children}</HPCClientProvider>;
  };
}

describe('useWorkloadTemplates', () => {
  it('starts loading and settles unavailable without fixture data', async () => {
    const { result } = renderHook(() => useWorkloadTemplates());

    expect(result.current).toEqual({ templates: [], isLoading: true, error: null });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.templates).toEqual([]);
    expect(result.current.error).toBeInstanceOf(HPCClientUnavailableError);
    expect(result.current.error).toMatchObject({ capability: 'query' });
  });

  it('shares the injected client across query hooks', async () => {
    const listWorkloadTemplates = vi.fn().mockResolvedValue([]);
    const listJobs = vi.fn().mockResolvedValue([]);
    const query = { ...createQueryAdapter(), listWorkloadTemplates, listJobs };
    const client = createHPCClient({ query });
    const { result } = renderHook(() => ({ templates: useWorkloadTemplates(), jobs: useJobs() }), {
      wrapper: providerWrapper(client),
    });

    await waitFor(() => {
      expect(result.current.templates.isLoading).toBe(false);
      expect(result.current.jobs.isLoading).toBe(false);
    });

    expect(listWorkloadTemplates).toHaveBeenCalledOnce();
    expect(listJobs).toHaveBeenCalledOnce();
  });

  it('surfaces failures from the injected authoritative dependency', async () => {
    const dependencyFailure = new Error('authoritative query failed');
    const listWorkloadTemplates = vi.fn().mockRejectedValue(dependencyFailure);
    const query = { ...createQueryAdapter(), listWorkloadTemplates };
    const { result } = renderHook(() => useWorkloadTemplates(), {
      wrapper: providerWrapper(createHPCClient({ query })),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.error).toBe(dependencyFailure);
  });

  it('uses the injected client for submit and cancel mutations', async () => {
    const committed = {
      committed: true as const,
      jobId: 'chain-job-42',
      txHash: 'ABCDEF012345',
      code: 0 as const,
      blockHeight: 42,
    };
    const submitJob = vi.fn().mockResolvedValue(committed);
    const cancelJob = vi.fn().mockResolvedValue(committed);
    const signer: HPCSignerAdapter = {
      state: 'signing-ready',
      submitJob,
      cancelJob,
    };
    const { result } = renderHook(
      () => ({ submission: useJobSubmission(), cancellation: useJobCancellation() }),
      { wrapper: providerWrapper(createHPCClient({ signer })) }
    );

    await act(async () => {
      await expect(
        result.current.submission.submitJob({
          offeringId: 'offering-1',
          name: 'injected job',
          resources: {
            nodes: 1,
            cpusPerNode: 4,
            memoryGBPerNode: 16,
            maxRuntimeSeconds: 3600,
            storageGB: 20,
          },
        })
      ).resolves.toEqual(committed);
      await expect(result.current.cancellation.cancelJob(committed.jobId)).resolves.toEqual(
        committed
      );
    });

    expect(submitJob).toHaveBeenCalledOnce();
    expect(cancelJob).toHaveBeenCalledWith(committed.jobId);
  });
});

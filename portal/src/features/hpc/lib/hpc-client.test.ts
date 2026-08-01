import { describe, expect, it, vi } from 'vitest';
import {
  HPCMutationNotCommittedError,
  createHPCClient,
  type HPCQueryAdapter,
  type HPCSignerAdapter,
  type SubmitJobParams,
} from './hpc-client';

const submitParams: SubmitJobParams = {
  offeringId: 'offering-authoritative',
  name: 'authoritative job',
  resources: {
    nodes: 1,
    cpusPerNode: 4,
    memoryGBPerNode: 16,
    maxRuntimeSeconds: 3600,
    storageGB: 20,
  },
};

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

describe('HPCClient', () => {
  it('is query-unavailable by default instead of selecting mock data', async () => {
    await expect(createHPCClient().listJobs()).rejects.toEqual(
      expect.objectContaining({
        code: 'hpc_client_unavailable',
        capability: 'query',
      })
    );
  });

  it('propagates authoritative query dependency failures', async () => {
    const dependencyFailure = new Error('chain query failed');
    const listJobs = vi.fn().mockRejectedValue(dependencyFailure);
    const query = { ...createQueryAdapter(), listJobs };

    await expect(createHPCClient({ query }).listJobs()).rejects.toBe(dependencyFailure);
  });

  it('requires signing-ready state for mutations', async () => {
    const submitJob = vi.fn();
    const signer: HPCSignerAdapter = {
      state: 'query-only',
      submitJob,
      cancelJob: vi.fn(),
    };

    await expect(createHPCClient({ signer }).submitJob(submitParams)).rejects.toEqual(
      expect.objectContaining({ capability: 'signer' })
    );
    expect(submitJob).not.toHaveBeenCalled();
  });

  it('rejects a signer preview or broadcast response as mutation success', async () => {
    const signer: HPCSignerAdapter = {
      state: 'signing-ready',
      submitJob: vi.fn().mockResolvedValue({ txHash: 'preview-hash' }),
      cancelJob: vi.fn(),
    };

    await expect(createHPCClient({ signer }).submitJob(submitParams)).rejects.toBeInstanceOf(
      HPCMutationNotCommittedError
    );
  });

  it.each([
    { committed: true, jobId: '', txHash: 'hash', code: 0, blockHeight: 12 },
    { committed: true, jobId: 'job-12', txHash: '', code: 0, blockHeight: 12 },
    { committed: true, jobId: 'job-12', txHash: 'hash', code: 7, blockHeight: 12 },
    { committed: true, jobId: 'job-12', txHash: 'hash', code: 0, blockHeight: 0 },
    { committed: true, jobId: 'job-12', txHash: 'hash', code: 0, blockHeight: 1.5 },
  ])('rejects malformed or unsuccessful committed receipt %#', async (receipt) => {
    const signer: HPCSignerAdapter = {
      state: 'signing-ready',
      submitJob: vi.fn().mockResolvedValue(receipt),
      cancelJob: vi.fn(),
    };

    await expect(createHPCClient({ signer }).submitJob(submitParams)).rejects.toBeInstanceOf(
      HPCMutationNotCommittedError
    );
  });

  it('returns authoritative committed transaction and server job identity', async () => {
    const committed = {
      committed: true as const,
      jobId: 'chain-job-42',
      txHash: 'ABCDEF012345',
      code: 0 as const,
      blockHeight: 42,
    };
    const signer: HPCSignerAdapter = {
      state: 'signing-ready',
      submitJob: vi.fn().mockResolvedValue(committed),
      cancelJob: vi.fn().mockResolvedValue(committed),
    };
    const client = createHPCClient({ signer });

    await expect(client.submitJob(submitParams)).resolves.toEqual(committed);
    await expect(client.cancelJob(committed.jobId)).resolves.toEqual(committed);
  });

  it('rejects committed cancellation state for a different job identity', async () => {
    const signer: HPCSignerAdapter = {
      state: 'signing-ready',
      submitJob: vi.fn(),
      cancelJob: vi.fn().mockResolvedValue({
        committed: true,
        jobId: 'different-job',
        txHash: 'ABCDEF012345',
        code: 0,
        blockHeight: 42,
      }),
    };

    await expect(createHPCClient({ signer }).cancelJob('requested-job')).rejects.toBeInstanceOf(
      HPCMutationNotCommittedError
    );
  });
});

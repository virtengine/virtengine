import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createHPCClient, HPCClient } from '@/features/hpc/lib/hpc-client';
import { getChainClient } from '@/lib/chain-sdk';

const mockChainOfferings = [
  {
    offeringId: 'offering-1',
    providerAddress: 'virtengine1provider1abc',
    active: true,
    name: 'NVIDIA A100 Cluster',
    description: 'High-performance GPU cluster',
    pricing: {
      currency: 'uve',
      baseNodeHourPrice: '2500000',
      cpuCoreHourPrice: '15000',
      memoryGbHourPrice: '6000',
      gpuHourPrice: '2500000',
      storageGbPrice: '1000',
      networkGbPrice: '500',
    },
    requiredIdentityThreshold: 0,
    preconfiguredWorkloads: [
      {
        workloadId: 'pytorch-training',
        name: 'PyTorch Training',
        description: 'Train deep learning models with PyTorch.',
        category: 'ml_training',
        requiredResources: {
          nodes: 1,
          cpuCoresPerNode: 8,
          memoryGbPerNode: 64,
          gpusPerNode: 2,
          gpuType: 'nvidia-a100',
          storageGb: 100,
        },
        isPreconfigured: true,
        version: '1.0.0',
      },
    ],
    createdAt: new Date('2024-01-15T10:00:00Z'),
    updatedAt: new Date('2024-02-01T15:30:00Z'),
    maxRuntimeSeconds: 86400,
    queueOptions: [],
    supportsCustomWorkloads: true,
  },
];

const mockChainJobs = [
  {
    jobId: 'job-401',
    workloadSpec: {
      isPreconfigured: true,
      preconfiguredWorkloadId: 'ML Training - ResNet50',
    },
    state: 3,
    resources: {
      nodes: 1,
      cpuCoresPerNode: 8,
      memoryGbPerNode: 64,
      gpusPerNode: 2,
      gpuType: 'nvidia-a100',
      storageGb: 100,
    },
    maxRuntimeSeconds: 86400,
    customerAddress: 'virtengine1customer1',
    providerAddress: 'virtengine1provider1abc',
    offeringId: 'offering-1',
    createdAt: new Date(Date.now() - 3600 * 1000),
    startedAt: new Date(Date.now() - 3000 * 1000),
    agreedPrice: [{ denom: 'uve', amount: '5500000' }],
  },
  {
    jobId: 'job-402',
    workloadSpec: {
      isPreconfigured: true,
      preconfiguredWorkloadId: 'Batch Simulation',
    },
    state: 2,
    resources: {
      nodes: 1,
      cpuCoresPerNode: 4,
      memoryGbPerNode: 32,
      gpusPerNode: 0,
      gpuType: '',
      storageGb: 50,
    },
    maxRuntimeSeconds: 43200,
    customerAddress: 'virtengine1customer1',
    providerAddress: 'virtengine1provider1abc',
    offeringId: 'offering-1',
    createdAt: new Date(Date.now() - 7200 * 1000),
    agreedPrice: [{ denom: 'uve', amount: '1200000' }],
  },
  {
    jobId: 'job-403',
    workloadSpec: {
      isPreconfigured: true,
      preconfiguredWorkloadId: 'Render Batch',
    },
    state: 4,
    resources: {
      nodes: 1,
      cpuCoresPerNode: 16,
      memoryGbPerNode: 32,
      gpusPerNode: 0,
      gpuType: '',
      storageGb: 200,
    },
    maxRuntimeSeconds: 43200,
    customerAddress: 'virtengine1customer1',
    providerAddress: 'virtengine1provider1abc',
    offeringId: 'offering-1',
    createdAt: new Date(Date.now() - 86400 * 1000),
    startedAt: new Date(Date.now() - 82800 * 1000),
    completedAt: new Date(Date.now() - 72000 * 1000),
    agreedPrice: [{ denom: 'uve', amount: '10500000' }],
  },
];

const mockedGetChainClient = vi.mocked(getChainClient);

beforeEach(() => {
  mockedGetChainClient.mockResolvedValue({
    hpc: {
      listOfferings: vi.fn().mockResolvedValue(mockChainOfferings),
      listJobs: vi.fn().mockResolvedValue(mockChainJobs),
      getJob: vi.fn((jobId: string) => mockChainJobs.find((job) => job.jobId === jobId) ?? null),
      getOffering: vi.fn(
        (offeringId: string) =>
          mockChainOfferings.find((offering) => offering.offeringId === offeringId) ?? null
      ),
    },
    provider: {
      getProvider: vi.fn().mockResolvedValue(null),
    },
    market: {
      listOrders: vi.fn().mockResolvedValue([]),
    },
  } as any);
});

describe('HPCClient', () => {
  it('creates client instance via factory', () => {
    const client = createHPCClient();
    expect(client).toBeInstanceOf(HPCClient);
  });

  it('lists workload templates', async () => {
    const client = createHPCClient();
    const templates = await client.listWorkloadTemplates();

    expect(templates).toBeInstanceOf(Array);
    expect(templates.length).toBeGreaterThan(0);

    const template = templates[0];
    expect(template).toHaveProperty('id');
    expect(template).toHaveProperty('name');
    expect(template).toHaveProperty('category');
    expect(template).toHaveProperty('defaultResources');
  });

  it('gets a specific template by ID', async () => {
    const client = createHPCClient();
    const template = await client.getWorkloadTemplate('pytorch-training');

    expect(template).not.toBeNull();
    expect(template!.name).toBe('PyTorch Training');
    expect(template!.category).toBe('ml_training');
  });

  it('returns null for unknown template ID', async () => {
    const client = createHPCClient();
    const template = await client.getWorkloadTemplate('nonexistent');

    expect(template).toBeNull();
  });

  it('lists offerings', async () => {
    const client = createHPCClient();
    const offerings = await client.listOfferings();

    expect(offerings).toBeInstanceOf(Array);
    expect(offerings.length).toBeGreaterThan(0);
    expect(offerings[0]).toHaveProperty('offeringId');
    expect(offerings[0]).toHaveProperty('pricing');
  });

  it('lists jobs', async () => {
    const client = createHPCClient();
    const jobs = await client.listJobs();

    expect(jobs).toBeInstanceOf(Array);
    expect(jobs.length).toBeGreaterThan(0);

    const job = jobs[0];
    expect(job).toHaveProperty('id');
    expect(job).toHaveProperty('name');
    expect(job).toHaveProperty('status');
    expect(job).toHaveProperty('resources');
  });

  it('filters jobs by status', async () => {
    const client = createHPCClient();
    const runningJobs = await client.listJobs({ status: ['running'] });

    expect(runningJobs.every((j) => j.status === 'running')).toBe(true);
  });

  it('gets a specific job by ID', async () => {
    const client = createHPCClient();
    const job = await client.getJob('job-401');

    expect(job).not.toBeNull();
    expect(job!.name).toBe('ML Training - ResNet50');
    expect(job!.status).toBe('running');
  });

  it('returns null for unknown job ID', async () => {
    const client = createHPCClient();
    const job = await client.getJob('nonexistent');

    expect(job).toBeNull();
  });

  it('submits a job and returns jobId and txHash', async () => {
    const client = createHPCClient();
    const result = await client.submitJob({
      offeringId: 'offering-1',
      name: 'Test Job',
      resources: {
        nodes: 1,
        cpusPerNode: 4,
        memoryGBPerNode: 32,
        maxRuntimeSeconds: 3600,
        storageGB: 50,
      },
    });

    expect(result).toHaveProperty('jobId');
    expect(result).toHaveProperty('txHash');
    expect(result.jobId).toBeTruthy();
    expect(result.txHash).toBeTruthy();
  });

  it('cancels a job and returns txHash', async () => {
    const client = createHPCClient();
    const result = await client.cancelJob('job-401');

    expect(result).toHaveProperty('txHash');
    expect(result.txHash).toBeTruthy();
  });

  it('estimates job cost', async () => {
    const client = createHPCClient();
    const estimate = await client.estimateJobCost('offering-1', {
      nodes: 2,
      cpusPerNode: 8,
      memoryGBPerNode: 64,
      gpusPerNode: 1,
      maxRuntimeSeconds: 7200,
      storageGB: 100,
    });

    expect(estimate).toHaveProperty('estimatedTotal');
    expect(estimate).toHaveProperty('pricePerHour');
    expect(estimate).toHaveProperty('breakdown');
    expect(estimate.breakdown).toHaveProperty('compute');
    expect(estimate.breakdown).toHaveProperty('storage');
    expect(estimate).toHaveProperty('denom');
    expect(parseFloat(estimate.estimatedTotal)).toBeGreaterThan(0);
  });

  it('gets job logs', async () => {
    const client = createHPCClient();
    const logs = await client.getJobLogs('job-401');

    expect(logs).toHaveProperty('lines');
    expect(logs).toHaveProperty('hasMore');
    expect(logs.lines).toBeInstanceOf(Array);
    expect(logs.lines.length).toBeGreaterThan(0);
  });

  it('respects tail parameter for logs', async () => {
    const client = createHPCClient();
    const logsAll = await client.getJobLogs('job-401');
    const logsTail = await client.getJobLogs('job-401', { tail: 5 });

    expect(logsTail.lines.length).toBeLessThanOrEqual(5);
    expect(logsTail.lines.length).toBeLessThanOrEqual(logsAll.lines.length);
  });

  it('gets job outputs for completed jobs', async () => {
    const client = createHPCClient();
    const outputs = await client.getJobOutputs('job-403');

    expect(outputs).toBeInstanceOf(Array);
    expect(outputs.length).toBeGreaterThan(0);
    expect(outputs[0]).toHaveProperty('refId');
    expect(outputs[0]).toHaveProperty('name');
    expect(outputs[0]).toHaveProperty('type');
    expect(outputs[0]).toHaveProperty('accessUrl');
    expect(outputs[0]).toHaveProperty('sizeBytes');
  });

  it('returns empty outputs for non-completed jobs', async () => {
    const client = createHPCClient();
    const outputs = await client.getJobOutputs('job-401'); // running

    expect(outputs).toEqual([]);
  });

  it('gets job resource usage for running jobs', async () => {
    const client = createHPCClient();
    const usage = await client.getJobUsage('job-401');

    expect(usage).toHaveProperty('cpuPercent');
    expect(usage).toHaveProperty('memoryPercent');
    expect(usage).toHaveProperty('elapsedSeconds');
    expect(usage.cpuPercent).toBeGreaterThan(0);
    expect(usage.memoryPercent).toBeGreaterThan(0);
  });

  it('returns zero usage for non-running jobs', async () => {
    const client = createHPCClient();
    const usage = await client.getJobUsage('job-402'); // queued

    expect(usage.cpuPercent).toBe(0);
    expect(usage.memoryPercent).toBe(0);
    expect(usage.elapsedSeconds).toBe(0);
  });
});

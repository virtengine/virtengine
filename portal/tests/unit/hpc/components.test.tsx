import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import type { Job, WorkloadTemplate } from '@/features/hpc';
import { getChainClient } from '@/lib/chain-sdk';

const mockJobs: Job[] = [
  {
    id: 'job-1',
    name: 'ML Training Run',
    customerAddress: 'virtengine1cust...',
    providerAddress: 'virtengine1prov...',
    offeringId: 'offering-1',
    templateId: 'pytorch-training',
    status: 'running',
    createdAt: Date.now() - 3600000,
    startedAt: Date.now() - 3000000,
    resources: {
      nodes: 1,
      cpusPerNode: 8,
      memoryGBPerNode: 64,
      gpusPerNode: 2,
      gpuType: 'nvidia-a100',
      maxRuntimeSeconds: 86400,
      storageGB: 100,
    },
    statusHistory: [],
    events: [],
    outputRefs: [],
    totalCost: '5.50',
    depositAmount: '132.00',
    depositStatus: 'held',
    txHash: '0xabc...',
  },
  {
    id: 'job-2',
    name: 'Render Job',
    customerAddress: 'virtengine1cust...',
    providerAddress: 'virtengine1prov...',
    offeringId: 'offering-1',
    status: 'completed',
    createdAt: Date.now() - 86400000,
    startedAt: Date.now() - 82800000,
    completedAt: Date.now() - 72000000,
    resources: {
      nodes: 1,
      cpusPerNode: 16,
      memoryGBPerNode: 32,
      maxRuntimeSeconds: 43200,
      storageGB: 200,
    },
    statusHistory: [],
    events: [],
    outputRefs: [],
    totalCost: '10.50',
    depositAmount: '42.00',
    depositStatus: 'released',
    txHash: '0xdef...',
  },
];

const mockTemplates: WorkloadTemplate[] = [
  {
    id: 'pytorch-training',
    name: 'PyTorch Training',
    description: 'Train deep learning models with PyTorch.',
    category: 'ml_training',
    defaultResources: {
      nodes: 1,
      cpusPerNode: 8,
      memoryGBPerNode: 64,
      gpusPerNode: 2,
      gpuType: 'nvidia-a100',
      maxRuntimeSeconds: 86400,
      storageGB: 100,
    },
    defaultParameters: {},
    requiredIdentityScore: 0,
    mfaRequired: false,
    estimatedCostPerHour: '5.50',
    version: '1.0.0',
  },
  {
    id: 'openfoam',
    name: 'OpenFOAM',
    description: 'Computational fluid dynamics simulations.',
    category: 'scientific',
    defaultResources: {
      nodes: 4,
      cpusPerNode: 32,
      memoryGBPerNode: 128,
      maxRuntimeSeconds: 172800,
      storageGB: 500,
    },
    defaultParameters: {},
    requiredIdentityScore: 0,
    mfaRequired: false,
    estimatedCostPerHour: '12.00',
    version: '1.0.0',
  },
];

const mockChainOfferings = [
  {
    offeringId: 'offering-1',
    providerAddress: 'virtengine1prov...',
    active: true,
    name: 'HPC Cluster',
    description: 'Mock offering for tests',
    pricing: {
      currency: 'uve',
      baseNodeHourPrice: '1000000',
      cpuCoreHourPrice: '10000',
      memoryGbHourPrice: '5000',
      gpuHourPrice: '2000000',
      storageGbPrice: '1000',
      networkGbPrice: '500',
    },
    requiredIdentityThreshold: 0,
    preconfiguredWorkloads: mockTemplates.map((template) => ({
      workloadId: template.id,
      name: template.name,
      description: template.description,
      category: template.category,
      estimatedCostPerHour: template.estimatedCostPerHour,
      requiredResources: {
        nodes: template.defaultResources.nodes,
        cpuCoresPerNode: template.defaultResources.cpusPerNode,
        memoryGbPerNode: template.defaultResources.memoryGBPerNode,
        gpusPerNode: template.defaultResources.gpusPerNode ?? 0,
        gpuType: template.defaultResources.gpuType ?? '',
        storageGb: template.defaultResources.storageGB,
      },
      isPreconfigured: true,
      version: template.version,
    })),
    createdAt: new Date(Date.now() - 86400000),
    updatedAt: new Date(Date.now() - 3600000),
    maxRuntimeSeconds: 86400,
    queueOptions: [],
    supportsCustomWorkloads: true,
  },
];

const mockChainJobs = mockJobs.map((job) => ({
  jobId: job.id,
  workloadSpec: {
    isPreconfigured: Boolean(job.templateId),
    preconfiguredWorkloadId: job.name,
  },
  state: job.status === 'running' ? 3 : job.status === 'completed' ? 4 : 2,
  resources: {
    nodes: job.resources.nodes,
    cpuCoresPerNode: job.resources.cpusPerNode,
    memoryGbPerNode: job.resources.memoryGBPerNode,
    gpusPerNode: job.resources.gpusPerNode ?? 0,
    gpuType: job.resources.gpuType ?? '',
    storageGb: job.resources.storageGB,
  },
  maxRuntimeSeconds: job.resources.maxRuntimeSeconds,
  customerAddress: job.customerAddress,
  providerAddress: job.providerAddress,
  offeringId: job.offeringId,
  createdAt: new Date(job.createdAt),
  startedAt: job.startedAt ? new Date(job.startedAt) : undefined,
  completedAt: job.completedAt ? new Date(job.completedAt) : undefined,
  agreedPrice: [{ denom: 'uve', amount: job.totalCost }],
}));

const mockedGetChainClient = vi.mocked(getChainClient);

beforeEach(() => {
  mockedGetChainClient.mockResolvedValue({
    hpc: {
      listOfferings: vi.fn().mockResolvedValue(mockChainOfferings),
      listJobs: vi.fn().mockResolvedValue(mockChainJobs),
      getJob: vi.fn((jobId: string) => mockChainJobs.find((job) => job.jobId === jobId) ?? null),
      getOffering: vi.fn().mockResolvedValue(mockChainOfferings[0]),
    },
    provider: {
      getProvider: vi.fn().mockResolvedValue(null),
    },
    market: {
      listOrders: vi.fn().mockResolvedValue([]),
    },
  } as any);
});

// Mock the HPC client at the source level so hooks get mock data
vi.mock('@/features/hpc/lib/hpc-client', async () => {
  const actual = await vi.importActual<typeof import('@/features/hpc/lib/hpc-client')>(
    '@/features/hpc/lib/hpc-client'
  );
  return {
    ...actual,
    createHPCClient: () => ({
      listJobs: async (filters?: { status?: string[] }) => {
        let filtered = mockJobs;
        if (filters?.status && filters.status.length > 0) {
          filtered = mockJobs.filter((j) => filters.status!.includes(j.status));
        }
        return filtered;
      },
      listWorkloadTemplates: async () => mockTemplates,
      getJobStatistics: async () => ({ running: 1, queued: 0, completed: 1, failed: 0 }),
      getJob: async () => null,
      submitJob: async () => ({ jobId: 'new', txHash: '0x...' }),
      cancelJob: async () => ({ txHash: '0x...' }),
      estimateJobCost: async () => ({
        estimatedTotal: '0',
        pricePerHour: '0',
        breakdown: { compute: '0', storage: '0', network: '0', gpu: '0' },
        denom: 'uakt',
      }),
      getJobLogs: async () => ({ lines: [], hasMore: false }),
      getJobOutputs: async () => [],
      getJobUsage: async () => ({
        cpuPercent: 0,
        memoryPercent: 0,
        gpuPercent: 0,
        elapsedSeconds: 0,
        remainingSeconds: 0,
      }),
      listOfferings: async () => [],
      getWorkloadTemplate: async () => null,
      delay: () => Promise.resolve(),
    }),
  };
});

import { JobList } from '@/features/hpc/components/JobList';
import { JobStatistics } from '@/features/hpc/components/JobStatistics';
import { TemplateBrowser } from '@/features/hpc/components/TemplateBrowser';

describe('JobList', () => {
  it('renders job cards with names', async () => {
    render(<JobList />);
    expect(await screen.findByText('ML Training Run')).toBeInTheDocument();
    expect(screen.getByText('Render Job')).toBeInTheDocument();
  });

  it('shows status badges', async () => {
    render(<JobList />);
    expect(await screen.findByText('Running')).toBeInTheDocument();
    expect(screen.getByText('Completed')).toBeInTheDocument();
  });

  it('shows resource info', async () => {
    render(<JobList />);
    expect(await screen.findByText('8 CPUs')).toBeInTheDocument();
    expect(screen.getByText('2 GPUs')).toBeInTheDocument();
  });

  it('renders View Details links', async () => {
    render(<JobList />);
    await screen.findByText('ML Training Run');
    const links = screen.getAllByText('View Details');
    expect(links).toHaveLength(2);
  });

  it('shows cancel button for running jobs', async () => {
    render(<JobList />);
    await screen.findByText('ML Training Run');
    const cancelButtons = screen.getAllByText('Cancel');
    expect(cancelButtons.length).toBeGreaterThanOrEqual(1);
  });

  it('shows download output button for completed jobs', async () => {
    render(<JobList />);
    expect(await screen.findByText('Download Output')).toBeInTheDocument();
  });

  it('shows view logs button for completed/failed jobs', async () => {
    render(<JobList />);
    expect(await screen.findByText('View Logs')).toBeInTheDocument();
  });
});

describe('JobStatistics', () => {
  it('renders stat cards', async () => {
    render(<JobStatistics />);
    expect(await screen.findByText('Running')).toBeInTheDocument();
    expect(screen.getByText('Queued')).toBeInTheDocument();
    expect(screen.getByText('Completed (24h)')).toBeInTheDocument();
    expect(screen.getByText('Failed (24h)')).toBeInTheDocument();
  });

  it('displays correct stat values', async () => {
    render(<JobStatistics />);
    await screen.findByText('Running');
    const runningCard = screen.getByText('Running').closest('div');
    expect(runningCard).toBeInTheDocument();
  });
});

describe('TemplateBrowser', () => {
  it('renders template cards', async () => {
    render(<TemplateBrowser />);
    expect(await screen.findByText('PyTorch Training')).toBeInTheDocument();
    expect(screen.getByText('OpenFOAM')).toBeInTheDocument();
  });

  it('renders category filters', async () => {
    render(<TemplateBrowser />);
    expect(await screen.findByText('All')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Machine Learning' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Scientific Computing' })).toBeInTheDocument();
  });

  it('filters templates by category', async () => {
    render(<TemplateBrowser />);
    await screen.findByText('PyTorch Training');

    const sciBtn = screen.getByRole('button', { name: 'Scientific Computing' });
    fireEvent.click(sciBtn);

    await waitFor(() => {
      expect(screen.getByText('OpenFOAM')).toBeInTheDocument();
      expect(screen.queryByText('PyTorch Training')).not.toBeInTheDocument();
    });
  });

  it('shows GPU badge for GPU templates', async () => {
    render(<TemplateBrowser />);
    expect(await screen.findByText('GPU')).toBeInTheDocument();
  });

  it('shows estimated cost per hour', async () => {
    render(<TemplateBrowser />);
    await screen.findByText('PyTorch Training');
    // Cost is rendered as "$" + value + "/hr" in separate text nodes inside a span
    const costSpans = document.querySelectorAll('.font-medium');
    const costTexts = Array.from(costSpans).map((el) => el.textContent);
    expect(costTexts).toContain('$5.50/hr');
    expect(costTexts).toContain('$12.00/hr');
  });

  it('renders Use Template links', async () => {
    render(<TemplateBrowser />);
    await screen.findByText('PyTorch Training');
    const links = screen.getAllByText('Use Template');
    expect(links).toHaveLength(2);
  });
});

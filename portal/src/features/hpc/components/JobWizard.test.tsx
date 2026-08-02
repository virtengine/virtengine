import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useWizardStore } from '../stores/wizardStore';
import { JobWizard } from './JobWizard';

const mocks = vi.hoisted(() => ({
  push: vi.fn(),
  submitJob: vi.fn(),
  reset: vi.fn(),
  estimateCost: vi.fn().mockResolvedValue({
    estimatedTotal: '1',
    pricePerHour: '1',
    breakdown: { compute: '1', storage: '0', network: '0' },
    denom: 'uve',
  }),
}));

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: mocks.push }),
  useSearchParams: () => new URLSearchParams(),
}));

vi.mock('@/features/hpc', () => ({
  useCostEstimation: () => ({ estimateCost: mocks.estimateCost, isEstimating: false }),
  useJobSubmission: () => ({
    submitJob: mocks.submitJob,
    isSubmitting: false,
    error: null,
  }),
  useWorkloadTemplate: () => ({ template: null, isLoading: false, error: null }),
  useWorkloadTemplates: () => ({ templates: [], isLoading: false, error: null }),
}));

describe('JobWizard submission receipt', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useWizardStore.setState({
      currentStep: 'review',
      selectedTemplate: null,
      manifest: {
        name: 'authoritative job',
        resources: {
          nodes: 1,
          cpusPerNode: 4,
          memoryGBPerNode: 16,
          maxRuntimeSeconds: 3600,
          storageGB: 20,
        },
      },
      estimatedCost: null,
      reset: mocks.reset,
    });
  });

  it('does not reset or navigate for a nonzero transaction code', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => undefined);
    mocks.submitJob.mockResolvedValue({
      committed: true,
      jobId: 'chain-job-42',
      txHash: 'ABCDEF012345',
      code: 9,
      blockHeight: 42,
    });

    render(<JobWizard />);
    fireEvent.click(screen.getByRole('button', { name: 'Submit Job' }));

    await waitFor(() => expect(consoleError).toHaveBeenCalled());
    expect(mocks.reset).not.toHaveBeenCalled();
    expect(mocks.push).not.toHaveBeenCalled();
    consoleError.mockRestore();
  });

  it('resets and navigates after an authoritative committed receipt', async () => {
    mocks.submitJob.mockResolvedValue({
      committed: true,
      jobId: 'chain-job-42',
      txHash: 'ABCDEF012345',
      code: 0,
      blockHeight: 42,
    });

    render(<JobWizard />);
    fireEvent.click(screen.getByRole('button', { name: 'Submit Job' }));

    await waitFor(() => expect(mocks.reset).toHaveBeenCalledOnce());
    expect(mocks.push).toHaveBeenCalledWith('/hpc/jobs/chain-job-42');
  });
});

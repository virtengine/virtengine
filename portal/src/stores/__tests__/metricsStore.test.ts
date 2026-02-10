import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useMetricsStore } from '@/stores/metricsStore';

class MockMultiProviderClient {
  initialize = vi.fn().mockResolvedValue(undefined);
  getProviders = vi.fn(() => [{ address: 've1provider', name: 'Provider One' }]);
  listAllDeployments = vi.fn().mockResolvedValue([{ id: 'dep-1', providerId: 've1provider' }]);
  getClient = vi.fn(() => ({
    getDeploymentMetrics: vi.fn().mockResolvedValue({
      cpu: { usage: 1, limit: 2 },
      memory: { usage: 2, limit: 4 },
      storage: { usage: 3, limit: 6 },
      network: { rxBytes: 10, txBytes: 20 },
    }),
  }));
}

vi.mock('@/lib/portal-adapter', () => ({
  MultiProviderClient: MockMultiProviderClient,
}));

const initialState = useMetricsStore.getState();

describe('metricsStore', () => {
  beforeEach(() => {
    useMetricsStore.setState(initialState, true);
  });

  it('aggregates provider metrics from daemon clients', async () => {
    await useMetricsStore.getState().fetchMetrics();

    const state = useMetricsStore.getState();
    expect(state.summary).not.toBeNull();
    expect(state.deploymentMetrics).toHaveLength(1);
    expect(state.summary?.totalProviders).toBe(1);
  });
});

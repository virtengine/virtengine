import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useDeploymentStore } from '@/stores/deploymentStore';

const { MockMultiProviderClient } = vi.hoisted(() => {
  class MockMultiProviderClient {
    initialize = vi.fn().mockResolvedValue(undefined);
    getDeployment = vi.fn().mockResolvedValue({
      id: 'deploy-1',
      providerId: 've1provider',
      owner: 've1owner',
      state: 'running',
      createdAt: '2024-01-01T00:00:00Z',
    });
    getClient = vi.fn(() => ({
      getDeploymentMetrics: vi.fn().mockResolvedValue({
        cpu: { usage: 1, limit: 2 },
        memory: { usage: 2, limit: 4 },
        storage: { usage: 3, limit: 6 },
      }),
      getDeploymentLogs: vi.fn().mockResolvedValue(['log line']),
    }));
    performAction = vi.fn().mockResolvedValue(undefined);
  }
  return { MockMultiProviderClient };
});

vi.mock('@/lib/portal-adapter', () => ({
  MultiProviderClient: MockMultiProviderClient,
}));

const initialState = useDeploymentStore.getState();

describe('deploymentStore', () => {
  beforeEach(() => {
    useDeploymentStore.setState(initialState, true);
  });

  it('fetches deployment details from provider daemon', async () => {
    await useDeploymentStore.getState().fetchDeployment('deploy-1');

    const state = useDeploymentStore.getState();
    expect(state.deployments).toHaveLength(1);
    expect(state.deployments[0].id).toBe('deploy-1');
    expect(state.deployments[0].logs).toHaveLength(1);
  });
});

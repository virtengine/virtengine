import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useDeploymentStore } from '@/stores/deploymentStore';

const { MockMultiProviderClient, mockGetDeployment, mockPerformAction } = vi.hoisted(() => {
  const mockGetDeployment = vi.fn();
  const mockPerformAction = vi.fn();
  class MockMultiProviderClient {
    initialize = vi.fn().mockResolvedValue(undefined);
    getDeployment = mockGetDeployment;
    getClient = vi.fn(() => ({
      getDeploymentMetrics: vi.fn().mockResolvedValue({
        cpu: { usage: 1, limit: 2 },
        memory: { usage: 2, limit: 4 },
        storage: { usage: 3, limit: 6 },
      }),
      getDeploymentLogs: vi.fn().mockResolvedValue(['log line']),
    }));
    performAction = mockPerformAction;
  }
  return { MockMultiProviderClient, mockGetDeployment, mockPerformAction };
});

vi.mock('@/lib/portal-adapter', () => ({
  MultiProviderClient: MockMultiProviderClient,
  ProviderDeploymentActionError: class ProviderDeploymentActionError extends Error {
    constructor(
      public code: string,
      message: string
    ) {
      super(message);
    }
  },
}));

const initialState = useDeploymentStore.getState();

const deployment = (overrides: Record<string, unknown> = {}) => ({
  id: 'deploy-1',
  providerId: 've1provider',
  owner: 've1owner',
  state: 'running',
  version: '2',
  revision: '7',
  createdAt: '2024-01-01T00:00:00Z',
  ...overrides,
});

const receipt = (overrides: Record<string, unknown> = {}) => ({
  operationId: 'operation-1',
  action: 'restart',
  deploymentId: 'deploy-1',
  providerId: 've1provider',
  status: 'committed',
  issuedAt: new Date('2026-08-01T12:00:00Z'),
  completedAt: new Date('2026-08-01T12:00:01Z'),
  state: 'running',
  version: '2',
  revision: '7',
  ...overrides,
});

describe('deploymentStore', () => {
  beforeEach(() => {
    useDeploymentStore.setState(initialState, true);
    mockGetDeployment.mockReset().mockResolvedValue(deployment());
    mockPerformAction.mockReset().mockResolvedValue(receipt());
  });

  it('fetches deployment details from provider daemon', async () => {
    await useDeploymentStore.getState().fetchDeployment('deploy-1');

    const state = useDeploymentStore.getState();
    expect(state.deployments).toHaveLength(1);
    expect(state.deployments[0].id).toBe('deploy-1');
    expect(state.deployments[0].logs).toHaveLength(1);
  });

  it('returns a receipt only after the refreshed deployment matches it', async () => {
    const result = await useDeploymentStore.getState().restartDeployment('deploy-1');

    expect(result.operationId).toBe('operation-1');
    expect(mockPerformAction).toHaveBeenCalledWith('deploy-1', 'restart');
    expect(mockGetDeployment).toHaveBeenCalled();
    expect(useDeploymentStore.getState().deployments[0]).toMatchObject({
      providerId: 've1provider',
      providerState: 'running',
      version: '2',
      revision: '7',
    });
  });

  it.each(['start', 'stop', 'restart', 'terminate'] as const)(
    'propagates the %s action receipt',
    async (action) => {
      mockPerformAction.mockResolvedValueOnce(receipt({ action }));
      const actionMethod = {
        start: useDeploymentStore.getState().startDeployment,
        stop: useDeploymentStore.getState().stopDeployment,
        restart: useDeploymentStore.getState().restartDeployment,
        terminate: useDeploymentStore.getState().terminateDeployment,
      }[action];

      await expect(actionMethod('deploy-1')).resolves.toMatchObject({ action });
    }
  );

  it('propagates the update action receipt', async () => {
    mockPerformAction.mockResolvedValueOnce(receipt({ action: 'update' }));

    await expect(
      useDeploymentStore.getState().updateDeployment('deploy-1', {
        resources: { cpu: 1, memory: 1, storage: 1 },
        containers: [],
        env: [],
        ports: [],
      })
    ).resolves.toMatchObject({ action: 'update' });
  });

  it.each([
    ['provider', { providerId: 'other-provider' }],
    ['state', { state: 'stopped' }],
    ['version', { version: '3' }],
    ['revision', { revision: '8' }],
  ])('rejects refreshed %s drift', async (_field, receiptOverride) => {
    mockPerformAction.mockResolvedValueOnce(receipt(receiptOverride));

    await expect(useDeploymentStore.getState().restartDeployment('deploy-1')).rejects.toMatchObject(
      { code: 'deployment_drift' }
    );
  });

  it('reports refresh failure after provider acceptance', async () => {
    mockGetDeployment.mockRejectedValueOnce(new Error('provider offline'));

    await expect(useDeploymentStore.getState().restartDeployment('deploy-1')).rejects.toMatchObject(
      { code: 'refresh_failed' }
    );
  });

  it('propagates typed feature_unavailable without refreshing', async () => {
    mockPerformAction.mockRejectedValueOnce(
      Object.assign(new Error('receipt capability missing'), {
        code: 'feature_unavailable',
      })
    );

    await expect(useDeploymentStore.getState().restartDeployment('deploy-1')).rejects.toMatchObject(
      { code: 'feature_unavailable' }
    );
    expect(mockGetDeployment).not.toHaveBeenCalled();
  });

  it('rejects a duplicate concurrent action for the same deployment', async () => {
    let resolveAction!: (value: ReturnType<typeof receipt>) => void;
    mockPerformAction.mockImplementationOnce(
      () => new Promise((resolve) => (resolveAction = resolve))
    );

    const first = useDeploymentStore.getState().restartDeployment('deploy-1');
    await expect(useDeploymentStore.getState().stopDeployment('deploy-1')).rejects.toMatchObject({
      code: 'duplicate_action',
    });
    resolveAction(receipt());
    await expect(first).resolves.toMatchObject({ operationId: 'operation-1' });
  });
});

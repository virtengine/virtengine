import { describe, expect, it, vi } from 'vitest';
import type { ProviderDeploymentActionReceipt } from '@/lib/portal-adapter';
import { executeProviderDeploymentAction } from './deploymentActionUI';

const receipt: ProviderDeploymentActionReceipt = {
  operationId: 'operation-1',
  action: 'restart',
  deploymentId: 'deployment-1',
  providerId: 'provider-1',
  status: 'committed',
  issuedAt: new Date('2026-08-01T12:00:00Z'),
  completedAt: new Date('2026-08-01T12:00:01Z'),
  state: 'running',
  version: '2',
  revision: '7',
};

const handlers = () => ({
  setPendingAction: vi.fn(),
  setActionError: vi.fn(),
  setLastReceipt: vi.fn(),
  requestWallet: vi.fn(),
});

describe('executeProviderDeploymentAction', () => {
  it('tracks loading and publishes only the returned receipt', async () => {
    const ui = handlers();

    await expect(
      executeProviderDeploymentAction('restart', () => Promise.resolve(receipt), ui)
    ).resolves.toBe(receipt);
    expect(ui.setPendingAction.mock.calls).toEqual([['restart'], [null]]);
    expect(ui.setActionError).toHaveBeenCalledWith(null);
    expect(ui.setLastReceipt).toHaveBeenCalledWith(receipt);
    expect(ui.requestWallet).not.toHaveBeenCalled();
  });

  it.each(['feature_unavailable', 'action_rejected', 'malformed_receipt', 'deployment_drift'])(
    'shows %s failures without requesting a wallet',
    async (code) => {
      const ui = handlers();
      const error = Object.assign(new Error(`failed: ${code}`), { code });

      await expect(
        executeProviderDeploymentAction('stop', async () => Promise.reject(error), ui)
      ).resolves.toBeNull();
      expect(ui.setActionError).toHaveBeenCalledWith(`failed: ${code}`);
      expect(ui.setLastReceipt).not.toHaveBeenCalled();
      expect(ui.requestWallet).not.toHaveBeenCalled();
      expect(ui.setPendingAction).toHaveBeenLastCalledWith(null);
    }
  );

  it('requests a wallet only when the adapter reports chain signing is required', async () => {
    const ui = handlers();
    const error = Object.assign(new Error('wallet required'), {
      code: 'chain_signing_required',
    });

    await executeProviderDeploymentAction('terminate', async () => Promise.reject(error), ui);

    expect(ui.requestWallet).toHaveBeenCalledOnce();
    expect(ui.setActionError).toHaveBeenCalledWith('wallet required');
  });
});

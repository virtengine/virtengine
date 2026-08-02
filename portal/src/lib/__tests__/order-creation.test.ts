import { describe, expect, it, vi } from 'vitest';
import { submitCommittedOrder } from '@/lib/order-creation';
import { OrderFeatureUnavailableError, type CreateOrderResult } from '@/stores/orderStore';

describe('submitCommittedOrder', () => {
  it('uses the identity bound to the committed result without consulting refreshed orders', async () => {
    const committed = {
      txHash: 'HASH',
      blockHeight: 42,
      orderId: 've1owner/7',
    } as CreateOrderResult;
    const refreshOrders = vi.fn().mockResolvedValue(undefined);

    await expect(
      submitCommittedOrder(vi.fn().mockResolvedValue(committed), refreshOrders)
    ).resolves.toEqual({ txHash: 'HASH', blockHeight: 42, orderId: 've1owner/7' });
    expect(refreshOrders).toHaveBeenCalledOnce();
  });

  it('does not produce a confirmed identity or refresh on feature unavailable', async () => {
    const refreshOrders = vi.fn().mockResolvedValue(undefined);

    await expect(
      submitCommittedOrder(
        vi.fn().mockRejectedValue(new OrderFeatureUnavailableError()),
        refreshOrders
      )
    ).rejects.toMatchObject({ code: 'feature_unavailable' });
    expect(refreshOrders).not.toHaveBeenCalled();
  });
});

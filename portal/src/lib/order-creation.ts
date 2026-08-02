import type { CreateOrderResult } from '@/stores/orderStore';

export interface ConfirmedOrderIdentity {
  txHash: string;
  blockHeight: number;
  orderId: string;
}

export async function submitCommittedOrder(
  createOrder: () => Promise<CreateOrderResult>,
  refreshOrders: () => Promise<void>
): Promise<ConfirmedOrderIdentity> {
  const result = await createOrder();
  await refreshOrders();
  return {
    txHash: result.txHash,
    blockHeight: result.blockHeight,
    orderId: result.orderId,
  };
}

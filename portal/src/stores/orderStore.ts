import { create } from 'zustand';
import {
  fetchPaginated,
  fetchChainJsonWithFallback,
  coerceNumber,
  coerceString,
  toDate,
  signAndBroadcastAmino,
  type WalletSigner,
} from '@/lib/api/chain';

export type OrderStatus =
  | 'pending'
  | 'matched'
  | 'deploying'
  | 'running'
  | 'paused'
  | 'stopped'
  | 'completed'
  | 'failed';

export interface Order {
  id: string;
  providerId: string;
  providerName: string;
  resourceType: string;
  status: OrderStatus;
  createdAt: Date;
  updatedAt: Date;
  cost: {
    hourlyRate: number;
    totalCost: number;
    currency: string;
  };
  resources: {
    cpu: number;
    memory: number;
    storage: number;
    gpu?: number;
  };
}

export interface OrderState {
  orders: Order[];
  selectedOrderId: string | null;
  isLoading: boolean;
  error: string | null;
  filters: {
    status: OrderStatus | 'all';
    sortBy: 'createdAt' | 'cost' | 'status';
    sortOrder: 'asc' | 'desc';
  };
}

export interface OrderActions {
  fetchOrders: (ownerAddress: string) => Promise<void>;
  selectOrder: (orderId: string | null) => void;
  updateOrderStatus: (orderId: string, status: OrderStatus) => void;
  setFilter: (filters: Partial<OrderState['filters']>) => void;
  createOrder: (payload: CreateOrderPayload, wallet: WalletSigner) => Promise<string>;
  closeOrder: (orderId: string, owner: string, wallet: WalletSigner) => Promise<void>;
}

export type OrderStore = OrderState & OrderActions;

export interface CreateOrderPayload {
  owner: string;
  offeringId: string;
  resources: Array<{
    resourceType: string;
    unit: string;
    quantity: number;
  }>;
  deposit: { denom: string; amount: string };
}

const ORDER_ENDPOINTS = ['/virtengine/market/v1beta5/orders', '/virtengine/market/v1/orders'];
const PROVIDER_ENDPOINTS = (address: string) => [
  `/virtengine/provider/v1/providers/${address}`,
  `/virtengine/provider/v1beta4/providers/${address}`,
];

const initialState: OrderState = {
  orders: [],
  selectedOrderId: null,
  isLoading: false,
  error: null,
  filters: {
    status: 'all',
    sortBy: 'createdAt',
    sortOrder: 'desc',
  },
};

const parseOrderStatus = (value: unknown): OrderStatus => {
  const normalized = coerceString(value, '').toLowerCase();
  if (normalized.includes('match')) return 'matched';
  if (normalized.includes('deploy')) return 'deploying';
  if (normalized.includes('run') || normalized.includes('active')) return 'running';
  if (normalized.includes('pause')) return 'paused';
  if (normalized.includes('close') || normalized.includes('complete')) return 'completed';
  if (normalized.includes('fail') || normalized.includes('error')) return 'failed';
  if (normalized.includes('stop') || normalized.includes('cancel')) return 'stopped';
  return 'pending';
};

const parseProviderName = (raw: Record<string, unknown>, fallback: string) => {
  const attributes = Array.isArray(raw.attributes) ? raw.attributes : [];
  for (const attr of attributes) {
    if (!attr || typeof attr !== 'object') continue;
    const record = attr as Record<string, unknown>;
    const key = coerceString(record.key, '').toLowerCase();
    if (['name', 'provider_name', 'moniker', 'organization'].includes(key)) {
      const value = coerceString(record.value, '');
      if (value) return value;
    }
  }
  const info = raw.info as Record<string, unknown> | undefined;
  const name = info ? coerceString(info.name, '') : '';
  return name || fallback;
};

export const useOrderStore = create<OrderStore>()((set, get) => ({
  ...initialState,

  fetchOrders: async (ownerAddress: string) => {
    set({ isLoading: true, error: null });

    try {
      if (!ownerAddress) {
        throw new Error('Wallet address is required to load orders.');
      }

      const [ordersResult] = await Promise.all([
        fetchPaginated<Record<string, unknown>>(ORDER_ENDPOINTS, 'orders', {
          params: { owner: ownerAddress },
        }),
      ]);

      const providerNames = new Map<string, string>();
      await Promise.all(
        ordersResult.items.map(async (order) => {
          const provider = coerceString(order.provider ?? order.provider_address, '');
          if (!provider || providerNames.has(provider)) return;
          try {
            const payload = await fetchChainJsonWithFallback<Record<string, unknown>>(
              PROVIDER_ENDPOINTS(provider)
            );
            const rawProvider =
              (payload.provider as Record<string, unknown> | undefined) ?? payload;
            providerNames.set(provider, parseProviderName(rawProvider, provider));
          } catch {
            providerNames.set(provider, provider);
          }
        })
      );

      const orders: Order[] = ordersResult.items.map((record) => {
        const provider = coerceString(
          record.provider ?? record.provider_address ?? record.providerAddress,
          ''
        );
        const createdAt = toDate(record.created_at ?? record.createdAt);
        const updatedAt = toDate(record.updated_at ?? record.updatedAt ?? record.created_at);
        const resources =
          record.resources && typeof record.resources === 'object'
            ? (record.resources as Record<string, unknown>)
            : undefined;

        return {
          id: coerceString(record.id ?? record.order_id ?? record.orderId, ''),
          providerId: provider,
          providerName: providerNames.get(provider) ?? provider,
          resourceType: coerceString(record.resource_type ?? record.resourceType, 'Compute'),
          status: parseOrderStatus(record.state ?? record.status),
          createdAt,
          updatedAt,
          cost: {
            hourlyRate: coerceNumber(record.hourly_rate ?? record.price_per_hour, 0),
            totalCost: coerceNumber(record.total_cost ?? record.cost, 0),
            currency: coerceString(record.currency, 'uve'),
          },
          resources: {
            cpu: coerceNumber(resources?.cpu ?? record.cpu, 0),
            memory: coerceNumber(resources?.memory ?? record.memory, 0),
            storage: coerceNumber(resources?.storage ?? record.storage, 0),
            gpu: coerceNumber(resources?.gpu ?? record.gpu, 0) || undefined,
          },
        };
      });

      set({ orders, isLoading: false });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to fetch orders',
      });
    }
  },

  selectOrder: (orderId: string | null) => {
    set({ selectedOrderId: orderId });
  },

  updateOrderStatus: (orderId: string, status: OrderStatus) => {
    const { orders } = get();
    set({
      orders: orders.map((order) =>
        order.id === orderId ? { ...order, status, updatedAt: new Date() } : order
      ),
    });
  },

  setFilter: (filters) => {
    set((state) => ({
      filters: { ...state.filters, ...filters },
    }));
  },

  createOrder: async (payload: CreateOrderPayload, wallet: WalletSigner) => {
    const msg = {
      typeUrl: '/virtengine.market.v1beta5.MsgCreateOrder',
      value: payload,
    };
    const result = await signAndBroadcastAmino(wallet, [msg], 'Create order');
    return result.txHash;
  },

  closeOrder: async (orderId: string, owner: string, wallet: WalletSigner) => {
    const msg = {
      typeUrl: '/virtengine.market.v1beta5.MsgCloseOrder',
      value: {
        id: orderId,
        owner,
      },
    };
    await signAndBroadcastAmino(wallet, [msg], 'Close order');
    set((state) => ({
      orders: state.orders.map((order) =>
        order.id === orderId ? { ...order, status: 'stopped', updatedAt: new Date() } : order
      ),
    }));
  },
}));

export const selectFilteredOrders = (state: OrderStore) => {
  const { orders, filters } = state;

  const filtered =
    filters.status === 'all' ? [...orders] : orders.filter((o) => o.status === filters.status);

  filtered.sort((a, b) => {
    const aValue = filters.sortBy === 'cost' ? a.cost.totalCost : a[filters.sortBy];
    const bValue = filters.sortBy === 'cost' ? b.cost.totalCost : b[filters.sortBy];

    if (aValue < bValue) return filters.sortOrder === 'asc' ? -1 : 1;
    if (aValue > bValue) return filters.sortOrder === 'asc' ? 1 : -1;
    return 0;
  });

  return filtered;
};

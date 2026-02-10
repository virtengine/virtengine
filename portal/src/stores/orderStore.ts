import { create } from 'zustand';

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
  fetchOrders: () => Promise<void>;
  selectOrder: (orderId: string | null) => void;
  updateOrderStatus: (orderId: string, status: OrderStatus) => void;
  setFilter: (filters: Partial<OrderState['filters']>) => void;
  cancelOrder: (orderId: string) => Promise<void>;
}

export type OrderStore = OrderState & OrderActions;

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

export const useOrderStore = create<OrderStore>()((set, get) => ({
  ...initialState,

  fetchOrders: async () => {
    set({ isLoading: true, error: null });

    try {
      // In production, this would fetch from the API
      await new Promise((resolve) => setTimeout(resolve, 1000));

      const mockOrders: Order[] = [
        {
          id: '1001',
          providerId: 'provider-1',
          providerName: 'CloudCore',
          resourceType: 'GPU Compute',
          status: 'running',
          createdAt: new Date(Date.now() - 86400000 * 3),
          updatedAt: new Date(),
          cost: {
            hourlyRate: 2.5,
            totalCost: 180,
            currency: 'USD',
          },
          resources: {
            cpu: 16,
            memory: 64,
            storage: 500,
            gpu: 2,
          },
        },
        {
          id: '1002',
          providerId: 'provider-2',
          providerName: 'DataNexus',
          resourceType: 'CPU Cluster',
          status: 'running',
          createdAt: new Date(Date.now() - 86400000),
          updatedAt: new Date(),
          cost: {
            hourlyRate: 0.45,
            totalCost: 10.8,
            currency: 'USD',
          },
          resources: {
            cpu: 32,
            memory: 128,
            storage: 1000,
          },
        },
      ];

      set({ orders: mockOrders, isLoading: false });
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

  cancelOrder: async (orderId: string) => {
    try {
      // In production, this would call the API
      await new Promise((resolve) => setTimeout(resolve, 500));

      const { orders } = get();
      set({
        orders: orders.map((order) =>
          order.id === orderId
            ? { ...order, status: 'stopped' as OrderStatus, updatedAt: new Date() }
            : order
        ),
      });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to cancel order',
      });
    }
  },
}));

// Selectors
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

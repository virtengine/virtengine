import { create } from 'zustand';
import type {
  Offering,
  OfferingFilters,
  Provider,
} from '@/types/offerings';

export interface OfferingStoreState {
  offerings: Offering[];
  providers: Map<string, Provider>;
  selectedOffering: Offering | null;
  isLoading: boolean;
  isLoadingDetail: boolean;
  error: string | null;
  filters: OfferingFilters;
  pagination: {
    page: number;
    pageSize: number;
    total: number;
    nextKey: string | null;
  };
}

export interface OfferingStoreActions {
  fetchOfferings: () => Promise<void>;
  fetchOffering: (providerAddress: string, sequence: number) => Promise<void>;
  fetchProvider: (address: string) => Promise<Provider | null>;
  setFilters: (filters: Partial<OfferingFilters>) => void;
  resetFilters: () => void;
  setPage: (page: number) => void;
  clearError: () => void;
}

export type OfferingStore = OfferingStoreState & OfferingStoreActions;

const DEFAULT_FILTERS: OfferingFilters = {
  category: 'all',
  region: 'all',
  priceRange: null,
  minReputation: 0,
  tags: [],
  search: '',
  state: 'active',
};

const initialState: OfferingStoreState = {
  offerings: [],
  providers: new Map(),
  selectedOffering: null,
  isLoading: false,
  isLoadingDetail: false,
  error: null,
  filters: DEFAULT_FILTERS,
  pagination: {
    page: 1,
    pageSize: 12,
    total: 0,
    nextKey: null,
  },
};

// Mock data for development - in production this queries the blockchain
const MOCK_OFFERINGS: Offering[] = [
  {
    id: { providerAddress: 'virtengine1provider1abc', sequence: 1 },
    state: 'active',
    category: 'gpu',
    name: 'NVIDIA A100 Cluster',
    description: 'High-performance GPU cluster with NVIDIA A100 80GB GPUs, ideal for ML training and inference workloads.',
    version: '1.0.0',
    pricing: { model: 'hourly', basePrice: '2500000', currency: 'uve' },
    prices: [
      { resourceType: 'gpu', unit: 'hour', price: { denom: 'uve', amount: '2500000' }, usdReference: '2.50' },
      { resourceType: 'cpu', unit: 'vcpu-hour', price: { denom: 'uve', amount: '10000' }, usdReference: '0.01' },
      { resourceType: 'ram', unit: 'gb-hour', price: { denom: 'uve', amount: '5000' }, usdReference: '0.005' },
    ],
    allowBidding: false,
    identityRequirement: { minScore: 50, requiredStatus: '', requireVerifiedEmail: true, requireVerifiedDomain: false, requireMFA: false },
    requireMFAForOrders: false,
    specifications: { gpu: 'NVIDIA A100 80GB', vram: '80 GB', cpu: '32 vCPU', memory: '128 GB', storage: '1 TB NVMe' },
    tags: ['gpu', 'ml', 'training', 'a100'],
    regions: ['us-west', 'us-east', 'eu-west'],
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-02-01T15:30:00Z',
    totalOrderCount: 156,
    activeOrderCount: 12,
  },
  {
    id: { providerAddress: 'virtengine1provider2xyz', sequence: 1 },
    state: 'active',
    category: 'compute',
    name: 'AMD EPYC 7763 Instance',
    description: 'High-core-count CPU instances powered by AMD EPYC 7763 processors, perfect for parallel workloads.',
    version: '1.2.0',
    pricing: { model: 'hourly', basePrice: '450000', currency: 'uve' },
    prices: [
      { resourceType: 'cpu', unit: 'vcpu-hour', price: { denom: 'uve', amount: '15000' }, usdReference: '0.015' },
      { resourceType: 'ram', unit: 'gb-hour', price: { denom: 'uve', amount: '3000' }, usdReference: '0.003' },
      { resourceType: 'storage', unit: 'gb-month', price: { denom: 'uve', amount: '100000' }, usdReference: '0.10' },
    ],
    allowBidding: true,
    minBid: { denom: 'uve', amount: '350000' },
    identityRequirement: { minScore: 0, requiredStatus: '', requireVerifiedEmail: false, requireVerifiedDomain: false, requireMFA: false },
    requireMFAForOrders: false,
    specifications: { cpu: 'AMD EPYC 7763 (64 cores)', memory: '256 GB DDR4', storage: '2 TB NVMe' },
    tags: ['cpu', 'amd', 'epyc', 'compute'],
    regions: ['us-east', 'eu-central'],
    createdAt: '2024-01-20T08:00:00Z',
    updatedAt: '2024-02-05T12:00:00Z',
    totalOrderCount: 89,
    activeOrderCount: 23,
  },
  {
    id: { providerAddress: 'virtengine1provider3def', sequence: 1 },
    state: 'active',
    category: 'hpc',
    name: 'HPC Compute Node',
    description: 'Enterprise HPC node with InfiniBand networking for scientific computing and simulations.',
    version: '2.0.0',
    pricing: { model: 'hourly', basePrice: '8000000', currency: 'uve' },
    prices: [
      { resourceType: 'cpu', unit: 'node-hour', price: { denom: 'uve', amount: '8000000' }, usdReference: '8.00' },
    ],
    allowBidding: false,
    identityRequirement: { minScore: 75, requiredStatus: 'verified', requireVerifiedEmail: true, requireVerifiedDomain: true, requireMFA: true },
    requireMFAForOrders: true,
    specifications: { cpu: 'Intel Xeon Platinum 8380 (40 cores x 2)', memory: '512 GB', storage: '4 TB NVMe', network: 'InfiniBand HDR 200Gbps' },
    tags: ['hpc', 'scientific', 'simulation', 'infiniband'],
    regions: ['us-central'],
    createdAt: '2024-02-01T00:00:00Z',
    updatedAt: '2024-02-10T10:00:00Z',
    totalOrderCount: 34,
    activeOrderCount: 8,
  },
  {
    id: { providerAddress: 'virtengine1provider1abc', sequence: 2 },
    state: 'active',
    category: 'storage',
    name: 'NVMe Block Storage',
    description: 'High-performance NVMe block storage with consistent low-latency access.',
    version: '1.1.0',
    pricing: { model: 'monthly', basePrice: '100000', currency: 'uve' },
    prices: [
      { resourceType: 'storage', unit: 'gb-month', price: { denom: 'uve', amount: '100000' }, usdReference: '0.10' },
    ],
    allowBidding: false,
    identityRequirement: { minScore: 0, requiredStatus: '', requireVerifiedEmail: false, requireVerifiedDomain: false, requireMFA: false },
    requireMFAForOrders: false,
    specifications: { type: 'NVMe SSD', iops: '100,000+', latency: '<1ms' },
    tags: ['storage', 'nvme', 'block'],
    regions: ['us-west', 'us-east', 'eu-west', 'asia-pacific'],
    createdAt: '2024-01-25T14:00:00Z',
    updatedAt: '2024-02-08T09:00:00Z',
    totalOrderCount: 245,
    activeOrderCount: 67,
  },
  {
    id: { providerAddress: 'virtengine1provider4ghi', sequence: 1 },
    state: 'active',
    category: 'gpu',
    name: 'RTX 4090 Gaming/AI',
    description: 'Consumer-grade GPU instances ideal for smaller ML workloads, rendering, and gaming.',
    version: '1.0.0',
    pricing: { model: 'hourly', basePrice: '500000', currency: 'uve' },
    prices: [
      { resourceType: 'gpu', unit: 'hour', price: { denom: 'uve', amount: '500000' }, usdReference: '0.50' },
      { resourceType: 'cpu', unit: 'vcpu-hour', price: { denom: 'uve', amount: '8000' }, usdReference: '0.008' },
      { resourceType: 'ram', unit: 'gb-hour', price: { denom: 'uve', amount: '4000' }, usdReference: '0.004' },
    ],
    allowBidding: true,
    minBid: { denom: 'uve', amount: '400000' },
    identityRequirement: { minScore: 25, requiredStatus: '', requireVerifiedEmail: true, requireVerifiedDomain: false, requireMFA: false },
    requireMFAForOrders: false,
    specifications: { gpu: 'NVIDIA RTX 4090', vram: '24 GB', cpu: '16 vCPU', memory: '64 GB' },
    tags: ['gpu', 'rtx4090', 'gaming', 'inference'],
    regions: ['us-west', 'eu-west'],
    createdAt: '2024-02-05T16:00:00Z',
    updatedAt: '2024-02-12T11:00:00Z',
    totalOrderCount: 78,
    activeOrderCount: 15,
  },
  {
    id: { providerAddress: 'virtengine1provider5jkl', sequence: 1 },
    state: 'active',
    category: 'ml',
    name: 'ML Training Platform',
    description: 'Managed ML training platform with auto-scaling, experiment tracking, and model registry.',
    version: '3.0.0',
    pricing: { model: 'usage_based', basePrice: '0', currency: 'uve' },
    prices: [
      { resourceType: 'gpu', unit: 'gpu-second', price: { denom: 'uve', amount: '700' }, usdReference: '0.0007' },
      { resourceType: 'cpu', unit: 'cpu-second', price: { denom: 'uve', amount: '3' }, usdReference: '0.000003' },
      { resourceType: 'storage', unit: 'gb-hour', price: { denom: 'uve', amount: '50' }, usdReference: '0.00005' },
    ],
    allowBidding: false,
    identityRequirement: { minScore: 50, requiredStatus: '', requireVerifiedEmail: true, requireVerifiedDomain: false, requireMFA: false },
    requireMFAForOrders: false,
    specifications: { gpuTypes: 'A100, V100, T4', framework: 'PyTorch, TensorFlow, JAX', features: 'Auto-scaling, Distributed Training' },
    tags: ['ml', 'training', 'platform', 'managed'],
    regions: ['us-west', 'us-east', 'eu-west', 'asia-pacific'],
    createdAt: '2024-01-10T00:00:00Z',
    updatedAt: '2024-02-15T08:00:00Z',
    totalOrderCount: 312,
    activeOrderCount: 45,
  },
];

const MOCK_PROVIDERS: Record<string, Provider> = {
  'virtengine1provider1abc': {
    address: 'virtengine1provider1abc',
    name: 'CloudCore',
    description: 'Enterprise cloud provider specializing in GPU and storage solutions.',
    reputation: 95,
    verified: true,
    totalOfferings: 5,
    totalOrders: 401,
    regions: ['us-west', 'us-east', 'eu-west', 'asia-pacific'],
    website: 'https://cloudcore.example',
    createdAt: '2023-06-15T00:00:00Z',
  },
  'virtengine1provider2xyz': {
    address: 'virtengine1provider2xyz',
    name: 'DataNexus',
    description: 'High-performance computing infrastructure for demanding workloads.',
    reputation: 88,
    verified: true,
    totalOfferings: 3,
    totalOrders: 89,
    regions: ['us-east', 'eu-central'],
    website: 'https://datanexus.example',
    createdAt: '2023-09-01T00:00:00Z',
  },
  'virtengine1provider3def': {
    address: 'virtengine1provider3def',
    name: 'SuperCloud HPC',
    description: 'Specialized HPC infrastructure for scientific research and simulations.',
    reputation: 92,
    verified: true,
    totalOfferings: 2,
    totalOrders: 34,
    regions: ['us-central'],
    website: 'https://supercloud-hpc.example',
    createdAt: '2023-11-20T00:00:00Z',
  },
  'virtengine1provider4ghi': {
    address: 'virtengine1provider4ghi',
    name: 'GPU Labs',
    description: 'Affordable GPU compute for developers and small teams.',
    reputation: 82,
    verified: false,
    totalOfferings: 2,
    totalOrders: 78,
    regions: ['us-west', 'eu-west'],
    createdAt: '2024-01-05T00:00:00Z',
  },
  'virtengine1provider5jkl': {
    address: 'virtengine1provider5jkl',
    name: 'AI Platform Co',
    description: 'Managed ML/AI platforms for enterprise and research.',
    reputation: 90,
    verified: true,
    totalOfferings: 4,
    totalOrders: 312,
    regions: ['us-west', 'us-east', 'eu-west', 'asia-pacific'],
    website: 'https://aiplatform.example',
    createdAt: '2023-07-10T00:00:00Z',
  },
};

export const useOfferingStore = create<OfferingStore>()((set, get) => ({
  ...initialState,

  fetchOfferings: async () => {
    set({ isLoading: true, error: null });

    try {
      const { filters, pagination } = get();

      // In production, query the blockchain:
      // const chainInfo = getChainInfo();
      // const response = await fetch(`${chainInfo.restEndpoint}/virtengine/market/v1/offerings?...`);
      // const data = await response.json();

      // Simulate network delay
      await new Promise((resolve) => setTimeout(resolve, 800));

      // Filter mock data based on filters
      let filtered = [...MOCK_OFFERINGS];

      if (filters.category !== 'all') {
        filtered = filtered.filter((o) => o.category === filters.category);
      }

      if (filters.region !== 'all') {
        filtered = filtered.filter((o) => o.regions?.includes(filters.region));
      }

      if (filters.state !== 'all') {
        filtered = filtered.filter((o) => o.state === filters.state);
      }

      if (filters.search) {
        const searchLower = filters.search.toLowerCase();
        filtered = filtered.filter(
          (o) =>
            o.name.toLowerCase().includes(searchLower) ||
            o.description.toLowerCase().includes(searchLower) ||
            o.tags?.some((t) => t.toLowerCase().includes(searchLower))
        );
      }

      if (filters.minReputation > 0) {
        filtered = filtered.filter((o) => {
          const provider = MOCK_PROVIDERS[o.id.providerAddress];
          return provider && provider.reputation >= filters.minReputation;
        });
      }

      // Pagination
      const startIdx = (pagination.page - 1) * pagination.pageSize;
      const paged = filtered.slice(startIdx, startIdx + pagination.pageSize);

      set({
        offerings: paged,
        isLoading: false,
        pagination: {
          ...pagination,
          total: filtered.length,
          nextKey: startIdx + pagination.pageSize < filtered.length ? 'next' : null,
        },
      });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to fetch offerings',
      });
    }
  },

  fetchOffering: async (providerAddress: string, sequence: number) => {
    set({ isLoadingDetail: true, error: null, selectedOffering: null });

    try {
      // In production:
      // const chainInfo = getChainInfo();
      // const response = await fetch(`${chainInfo.restEndpoint}/virtengine/market/v1/offerings/${providerAddress}/${sequence}`);
      // const data = await response.json();

      await new Promise((resolve) => setTimeout(resolve, 500));

      const offering = MOCK_OFFERINGS.find(
        (o) => o.id.providerAddress === providerAddress && o.id.sequence === sequence
      );

      if (!offering) {
        throw new Error('Offering not found');
      }

      set({ selectedOffering: offering, isLoadingDetail: false });
    } catch (error) {
      set({
        isLoadingDetail: false,
        error: error instanceof Error ? error.message : 'Failed to fetch offering',
      });
    }
  },

  fetchProvider: (address: string): Promise<Provider | null> => {
    const { providers } = get();

    if (providers.has(address)) {
      return Promise.resolve(providers.get(address) ?? null);
    }

    try {
      // In production:
      // const chainInfo = getChainInfo();
      // const response = await fetch(`${chainInfo.restEndpoint}/virtengine/market/v1/providers/${address}`);
      // const data = await response.json();

      const provider = MOCK_PROVIDERS[address];

      if (provider) {
        set((state) => ({
          providers: new Map(state.providers).set(address, provider),
        }));
      }

      return Promise.resolve(provider || null);
    } catch {
      return Promise.resolve(null);
    }
  },

  setFilters: (newFilters: Partial<OfferingFilters>) => {
    set((state) => ({
      filters: { ...state.filters, ...newFilters },
      pagination: { ...state.pagination, page: 1 },
    }));
  },

  resetFilters: () => {
    set({
      filters: DEFAULT_FILTERS,
      pagination: { ...initialState.pagination, page: 1 },
    });
  },

  setPage: (page: number) => {
    set((state) => ({
      pagination: { ...state.pagination, page },
    }));
  },

  clearError: () => {
    set({ error: null });
  },
}));

// Selectors
export const selectOfferingById = (
  state: OfferingStore,
  providerAddress: string,
  sequence: number
) => {
  return state.offerings.find(
    (o) => o.id.providerAddress === providerAddress && o.id.sequence === sequence
  );
};

export const selectProviderByAddress = (state: OfferingStore, address: string) => {
  return state.providers.get(address);
};

// Utility functions
export function formatPrice(amount: string, decimals: number = 6): string {
  const value = parseInt(amount, 10) / Math.pow(10, decimals);
  return value.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export function formatPriceUSD(usdReference: string | undefined): string {
  if (!usdReference) return '—';
  const value = parseFloat(usdReference);
  return `$${value.toFixed(value < 0.01 ? 4 : 2)}`;
}

export function getOfferingDisplayPrice(offering: Offering): { amount: string; unit: string } {
  if (offering.prices && offering.prices.length > 0) {
    const mainPrice = offering.prices[0];
    return {
      amount: formatPriceUSD(mainPrice.usdReference),
      unit: `/${mainPrice.unit.replace('-', ' ')}`,
    };
  }

  const baseAmount = formatPrice(offering.pricing.basePrice);
  const unitMap: Record<string, string> = {
    hourly: '/hour',
    daily: '/day',
    monthly: '/month',
    usage_based: '',
    fixed: '',
  };

  return {
    amount: `${baseAmount} VE`,
    unit: unitMap[offering.pricing.model] || '',
  };
}

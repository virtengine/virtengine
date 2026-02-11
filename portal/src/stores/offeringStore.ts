import { create } from 'zustand';
import type { HPCOffering, Provider as ChainProvider } from '@virtengine/chain-sdk';
import type {
  Offering,
  OfferingCategory,
  OfferingFilters,
  PriceComponent,
  Provider,
} from '@/types/offerings';
import { getChainClient } from '@/lib/chain-sdk';

export interface OfferingStoreState {
  offerings: Offering[];
  providers: Map<string, Provider>;
  selectedOffering: Offering | null;
  isLoading: boolean;
  isLoadingDetail: boolean;
  error: string | null;
  filters: OfferingFilters;
  viewMode: 'grid' | 'list';
  compareIds: string[];
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
  setViewMode: (mode: 'grid' | 'list') => void;
  toggleCompare: (offeringKey: string) => void;
  clearCompare: () => void;
  clearError: () => void;
}

export type OfferingStore = OfferingStoreState & OfferingStoreActions;

const DEFAULT_FILTERS: OfferingFilters = {
  category: 'all',
  region: 'all',
  priceRange: null,
  minReputation: 0,
  minCpuCores: 0,
  minMemoryGB: 0,
  minGpuCount: 0,
  tags: [],
  search: '',
  state: 'active',
  providerSearch: '',
  sortBy: 'name',
  sortOrder: 'asc',
};

const initialState: OfferingStoreState = {
  offerings: [],
  providers: new Map(),
  selectedOffering: null,
  isLoading: false,
  isLoadingDetail: false,
  error: null,
  filters: DEFAULT_FILTERS,
  viewMode: 'grid',
  compareIds: [],
  pagination: {
    page: 1,
    pageSize: 12,
    total: 0,
    nextKey: null,
  },
};

function coerceString(value: unknown, fallback = ''): string {
  if (typeof value === 'string') return value;
  if (typeof value === 'number') return value.toString();
  return fallback;
}

function deriveOfferingSequence(offeringId: string): number {
  const parts = offeringId.split(/[/_-]/);
  const last = parts[parts.length - 1] ?? offeringId;
  const parsed = Number.parseInt(last, 10);
  if (!Number.isNaN(parsed) && parsed > 0) return parsed;

  let hash = 0;
  for (const char of offeringId) {
    hash = (hash * 31 + char.charCodeAt(0)) % 100000;
  }
  return Math.max(1, hash);
}

function buildPriceComponents(offering: HPCOffering): PriceComponent[] | undefined {
  const pricing = offering.pricing;
  if (!pricing) return undefined;

  const denom = pricing.currency || 'uvirt';
  const components: PriceComponent[] = [
    {
      resourceType: 'cpu',
      unit: 'core-hour',
      price: { denom, amount: pricing.cpuCoreHourPrice || '0' },
    },
    {
      resourceType: 'ram',
      unit: 'gb-hour',
      price: { denom, amount: pricing.memoryGbHourPrice || '0' },
    },
    {
      resourceType: 'gpu',
      unit: 'gpu-hour',
      price: { denom, amount: pricing.gpuHourPrice || '0' },
    },
    {
      resourceType: 'storage',
      unit: 'gb',
      price: { denom, amount: pricing.storageGbPrice || '0' },
    },
    {
      resourceType: 'network',
      unit: 'gb',
      price: { denom, amount: pricing.networkGbPrice || '0' },
    },
  ];

  return components.filter((entry) => entry.price.amount !== '0');
}

function mapOffering(offering: HPCOffering): Offering {
  const sequence = deriveOfferingSequence(offering.offeringId);
  const pricing = offering.pricing;
  const category = (offering as { category?: OfferingCategory }).category ?? 'hpc';
  const regions = (offering as { regions?: string[] }).regions ?? [];

  return {
    id: { providerAddress: offering.providerAddress, sequence },
    state: offering.active ? 'active' : 'paused',
    category,
    name: offering.name || `HPC Offering ${sequence}`,
    description: offering.description ?? '',
    version: offering.updatedAt?.toISOString() ?? '',
    pricing: {
      model: 'hourly',
      basePrice: pricing?.baseNodeHourPrice ?? '0',
      currency: pricing?.currency ?? 'uvirt',
      usageRates: pricing
        ? {
            cpu: pricing.cpuCoreHourPrice ?? '0',
            memory: pricing.memoryGbHourPrice ?? '0',
            gpu: pricing.gpuHourPrice ?? '0',
            storage: pricing.storageGbPrice ?? '0',
            network: pricing.networkGbPrice ?? '0',
          }
        : undefined,
    },
    prices: buildPriceComponents(offering),
    allowBidding: false,
    identityRequirement: {
      minScore: offering.requiredIdentityThreshold ?? 0,
      requiredStatus: 'verified',
      requireVerifiedEmail: false,
      requireVerifiedDomain: false,
      requireMFA: false,
    },
    requireMFAForOrders: false,
    publicMetadata: {
      offeringId: offering.offeringId,
      clusterId: offering.clusterId,
      supportsCustomWorkloads: String(Boolean(offering.supportsCustomWorkloads)),
    },
    specifications: {
      max_runtime_seconds: String(offering.maxRuntimeSeconds ?? ''),
      queue_options: String(offering.queueOptions?.length ?? 0),
    },
    tags: offering.preconfiguredWorkloads?.map((workload) => workload.name) ?? [],
    regions,
    createdAt: offering.createdAt?.toISOString() ?? new Date(0).toISOString(),
    updatedAt: offering.updatedAt?.toISOString() ?? new Date(0).toISOString(),
    totalOrderCount: 0,
    activeOrderCount: 0,
  };
}

function parseAttributeMap(attributes: unknown) {
  const map = new Map<string, string[]>();
  if (!Array.isArray(attributes)) return map;
  attributes.forEach((attr) => {
    if (!attr || typeof attr !== 'object') return;
    const record = attr as Record<string, unknown>;
    const key = coerceString(record.key, '').toLowerCase();
    const value = coerceString(record.value, '');
    if (!key) return;
    if (!map.has(key)) {
      map.set(key, []);
    }
    map.get(key)?.push(value);
  });
  return map;
}

function parseProviderName(attributes: Map<string, string[]>, fallback: string) {
  const candidates = ['name', 'provider_name', 'moniker', 'organization', 'org', 'company'];
  for (const key of candidates) {
    const value = attributes.get(key)?.[0];
    if (value) return value;
  }
  return fallback;
}

function parseProviderDescription(attributes: Map<string, string[]>) {
  const candidates = ['description', 'about', 'summary', 'profile'];
  for (const key of candidates) {
    const value = attributes.get(key)?.[0];
    if (value) return value;
  }
  return undefined;
}

function parseProviderReputation(attributes: Map<string, string[]>) {
  const candidates = ['reputation', 'rating', 'score'];
  for (const key of candidates) {
    const value = attributes.get(key)?.[0];
    if (!value) continue;
    const parsed = Number.parseFloat(value);
    if (!Number.isNaN(parsed)) {
      return Math.max(0, Math.min(100, parsed));
    }
  }
  return 0;
}

function parseProviderVerified(attributes: Map<string, string[]>) {
  const candidates = ['verified', 'kyc', 'audited', 'certified'];
  for (const key of candidates) {
    const value = attributes.get(key)?.[0];
    if (!value) continue;
    const normalized = value.toLowerCase();
    if (normalized === 'true' || normalized === '1' || normalized === 'yes') {
      return true;
    }
  }
  return false;
}

function parseProviderRegions(attributes: Map<string, string[]>) {
  const candidates = ['region', 'regions', 'location'];
  const regions: string[] = [];
  for (const key of candidates) {
    const values = attributes.get(key);
    if (!values) continue;
    values.forEach((entry) => {
      entry
        .split(',')
        .map((part) => part.trim())
        .filter(Boolean)
        .forEach((region) => regions.push(region));
    });
  }
  return regions.length ? Array.from(new Set(regions)) : undefined;
}

function buildProviderStats(offerings: Offering[]) {
  const stats = new Map<
    string,
    { totalOfferings: number; totalOrders: number; regions: Set<string> }
  >();
  offerings.forEach((offering) => {
    const key = offering.id.providerAddress;
    if (!stats.has(key)) {
      stats.set(key, { totalOfferings: 0, totalOrders: 0, regions: new Set() });
    }
    const entry = stats.get(key);
    if (!entry) return;
    entry.totalOfferings += 1;
    entry.totalOrders += offering.totalOrderCount;
    offering.regions?.forEach((region) => entry.regions.add(region));
  });
  return stats;
}

function parseProvider(
  raw: ChainProvider,
  stats?: { totalOfferings: number; totalOrders: number; regions: Set<string> }
): Provider {
  const address = coerceString(raw.owner ?? '', '');
  const attributes = parseAttributeMap(raw.attributes);
  const name = parseProviderName(attributes, address ? `${address.slice(0, 10)}...` : 'Unknown');
  const description = parseProviderDescription(attributes);
  const reputation = parseProviderReputation(attributes);
  const verified = parseProviderVerified(attributes);
  const regions =
    parseProviderRegions(attributes) ?? (stats ? Array.from(stats.regions) : undefined);

  const info = (raw.info ?? {}) as Record<string, unknown>;

  return {
    address,
    name,
    description,
    reputation,
    verified,
    totalOfferings: stats?.totalOfferings ?? 0,
    totalOrders: stats?.totalOrders ?? 0,
    regions,
    website: coerceString(info.website, '') || undefined,
    createdAt: undefined,
  };
}

function parseSpecNumber(raw: unknown): number | null {
  if (raw === null || raw === undefined) return null;
  if (typeof raw === 'number') return raw;
  const value = String(raw);
  const countMatch = value.match(/(\\d+)\\s*x/i);
  if (countMatch) {
    return Number.parseFloat(countMatch[1]);
  }
  const match = value.match(/(\\d+(?:\\.\\d+)?)/);
  if (!match) return null;
  const parsed = Number.parseFloat(match[1]);
  return Number.isNaN(parsed) ? null : parsed;
}

function findSpecValue(
  specifications: Record<string, string> | undefined,
  keys: string[]
): number | null {
  if (!specifications) return null;
  const entries = Object.entries(specifications).map(
    ([key, value]) => [key.toLowerCase(), value] as const
  );
  for (const key of keys) {
    const entry = entries.find(([specKey]) => specKey === key || specKey.includes(key));
    if (entry) {
      const parsed = parseSpecNumber(entry[1]);
      if (parsed !== null) return parsed;
    }
  }
  return null;
}

function getOfferingPriceValue(offering: Offering): number | null {
  if (offering.prices && offering.prices.length > 0) {
    const ref = offering.prices[0].usdReference;
    const parsed = ref ? Number.parseFloat(ref) : Number.NaN;
    if (!Number.isNaN(parsed)) return parsed;
  }
  const base = Number.parseFloat(offering.pricing.basePrice);
  if (!Number.isNaN(base)) {
    return base / 1_000_000;
  }
  return null;
}

function matchesFilters(
  offering: Offering,
  filters: OfferingFilters,
  providers: Map<string, Provider>
) {
  if (filters.category !== 'all' && offering.category !== filters.category) return false;
  if (filters.region !== 'all') {
    const regions = offering.regions ?? providers.get(offering.id.providerAddress)?.regions;
    if (!regions?.includes(filters.region)) return false;
  }
  if (filters.state !== 'all' && offering.state !== filters.state) return false;

  if (filters.search) {
    const searchLower = filters.search.toLowerCase();
    const matchesText =
      offering.name.toLowerCase().includes(searchLower) ||
      offering.description.toLowerCase().includes(searchLower) ||
      offering.tags?.some((tag) => tag.toLowerCase().includes(searchLower));
    if (!matchesText) return false;
  }

  if (filters.priceRange) {
    const priceValue = getOfferingPriceValue(offering);
    if (priceValue === null) return false;
    if (priceValue < filters.priceRange.min || priceValue > filters.priceRange.max) return false;
  }

  if (filters.minCpuCores > 0) {
    const cpu = findSpecValue(offering.specifications, ['cpu', 'vcpu', 'core']);
    if (cpu === null || cpu < filters.minCpuCores) return false;
  }

  if (filters.minMemoryGB > 0) {
    const memory = findSpecValue(offering.specifications, ['memory', 'ram', 'mem']);
    if (memory === null || memory < filters.minMemoryGB) return false;
  }

  if (filters.minGpuCount > 0) {
    const gpu = findSpecValue(offering.specifications, ['gpu_count', 'gpus', 'gpu']);
    if (gpu === null || gpu < filters.minGpuCount) return false;
  }

  if (filters.minReputation > 0 || filters.providerSearch) {
    const provider = providers.get(offering.id.providerAddress);
    if (!provider) return false;
    if (filters.minReputation > 0 && provider.reputation < filters.minReputation) return false;
    if (filters.providerSearch) {
      const providerLower = filters.providerSearch.toLowerCase();
      if (!provider.name.toLowerCase().includes(providerLower)) return false;
    }
  }

  return true;
}

export const useOfferingStore = create<OfferingStore>()((set, get) => ({
  ...initialState,

  fetchOfferings: async () => {
    set({ isLoading: true, error: null });

    try {
      const { filters, pagination } = get();
      const client = await getChainClient();
      const offerings = (
        await client.hpc.listOfferings({ activeOnly: filters.state === 'active' })
      ).map((offering) => mapOffering(offering));

      const providerStats = buildProviderStats(offerings);

      const providers = new Map(get().providers);
      providerStats.forEach((stats, address) => {
        const existing = providers.get(address);
        if (existing) {
          providers.set(address, {
            ...existing,
            totalOfferings: stats.totalOfferings,
            totalOrders: stats.totalOrders,
            regions: existing.regions?.length ? existing.regions : Array.from(stats.regions),
          });
        }
      });

      if (filters.minReputation > 0 || filters.providerSearch) {
        await Promise.all(
          Array.from(providerStats.keys()).map(async (address) => {
            if (!providers.has(address)) {
              const provider = await get().fetchProvider(address);
              if (provider) {
                providers.set(address, provider);
              }
            }
          })
        );
      }

      const filtered = offerings.filter((offering) => matchesFilters(offering, filters, providers));

      filtered.sort((a, b) => {
        const dir = filters.sortOrder === 'asc' ? 1 : -1;
        switch (filters.sortBy) {
          case 'price': {
            const priceA = getOfferingPriceValue(a) ?? 0;
            const priceB = getOfferingPriceValue(b) ?? 0;
            return (priceA - priceB) * dir;
          }
          case 'reputation': {
            const repA = providers.get(a.id.providerAddress)?.reputation ?? 0;
            const repB = providers.get(b.id.providerAddress)?.reputation ?? 0;
            return (repA - repB) * dir;
          }
          case 'orders':
            return (a.totalOrderCount - b.totalOrderCount) * dir;
          case 'created':
            return (new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()) * dir;
          default:
            return a.name.localeCompare(b.name) * dir;
        }
      });

      const totalCount = filtered.length;
      const startIdx = (pagination.page - 1) * pagination.pageSize;
      const paged = filtered.slice(startIdx, startIdx + pagination.pageSize);

      set({
        offerings: paged,
        isLoading: false,
        providers,
        pagination: {
          ...pagination,
          total: totalCount,
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
      const client = await getChainClient();
      const offerings = await client.hpc.listOfferings({ activeOnly: false });
      const mapped = offerings.map((offering) => mapOffering(offering));
      const offering = mapped.find(
        (entry) => entry.id.providerAddress === providerAddress && entry.id.sequence === sequence
      );

      if (!offering) {
        throw new Error('Offering not found');
      }

      set({ selectedOffering: offering, isLoadingDetail: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to fetch offering';
      set({
        isLoadingDetail: false,
        error: message,
      });
    }
  },

  fetchProvider: async (address: string): Promise<Provider | null> => {
    const { providers, offerings } = get();

    if (providers.has(address)) {
      return providers.get(address) ?? null;
    }

    try {
      const providerStats = buildProviderStats(offerings).get(address);
      const client = await getChainClient();
      const rawProvider = await client.provider.getProvider(address);
      if (!rawProvider) return null;
      const provider = parseProvider(rawProvider, providerStats);

      set((state) => ({
        providers: new Map(state.providers).set(address, provider),
      }));

      return provider;
    } catch {
      return null;
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

  setViewMode: (mode: 'grid' | 'list') => {
    set({ viewMode: mode });
  },

  toggleCompare: (offeringKey: string) => {
    set((state) => {
      const exists = state.compareIds.includes(offeringKey);
      if (exists) {
        return { compareIds: state.compareIds.filter((id) => id !== offeringKey) };
      }
      if (state.compareIds.length >= 4) return state;
      return { compareIds: [...state.compareIds, offeringKey] };
    });
  },

  clearCompare: () => {
    set({ compareIds: [] });
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
export function offeringKey(offering: Offering): string {
  return `${offering.id.providerAddress}-${offering.id.sequence}`;
}

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

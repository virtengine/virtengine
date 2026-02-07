import { create } from 'zustand';
import type {
  Offering,
  OfferingCategory,
  OfferingFilters,
  OfferingID,
  OfferingState,
  PriceComponent,
  PricingModel,
  Provider,
} from '@/types/offerings';
import { getChainInfo } from '@/config/chains';

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

const OFFERING_STATE_MAP: Record<number, OfferingState> = {
  0: 'unspecified',
  1: 'active',
  2: 'paused',
  3: 'suspended',
  4: 'deprecated',
  5: 'terminated',
};

const OFFERING_CATEGORY_MAP: Record<number, OfferingCategory> = {
  0: 'other',
  1: 'compute',
  2: 'storage',
  3: 'network',
  4: 'hpc',
  5: 'gpu',
  6: 'ml',
  7: 'other',
};

const PRICING_MODEL_MAP: Record<number, PricingModel> = {
  0: 'hourly',
  1: 'hourly',
  2: 'daily',
  3: 'monthly',
  4: 'usage_based',
  5: 'fixed',
};

const OFFERINGS_ENDPOINTS = ['/virtengine/market/v1/offerings', '/marketplace/offerings'];
const OFFERING_ENDPOINTS = (offeringId: string) => [
  `/virtengine/market/v1/offerings/${offeringId}`,
  `/marketplace/offerings/${offeringId}`,
];
const PROVIDER_ENDPOINTS = (address: string) => [
  `/virtengine/provider/v1beta4/providers/${address}`,
  `/virtengine/provider/v1/providers/${address}`,
];

class ChainRequestError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ChainRequestError';
    this.status = status;
  }
}

type ChainOffering = Record<string, unknown>;
type ChainProvider = Record<string, unknown>;

interface ChainOfferingsResponse {
  offerings?: ChainOffering[];
  pagination?: {
    next_key?: string | null;
    nextKey?: string | null;
    total?: string | number;
  };
  next_key?: string | null;
  total?: string | number;
}

const MAX_OFFERING_PAGES = 10;
const OFFERING_PAGE_LIMIT = 200;

function parseOfferingState(value: unknown): OfferingState {
  if (typeof value === 'number') {
    return OFFERING_STATE_MAP[value] ?? 'unspecified';
  }
  if (typeof value === 'string') {
    const normalized = value.toLowerCase();
    if (normalized.startsWith('offering_state_')) {
      const name = value.replace(/offering_state_/i, '').toLowerCase();
      return (name as OfferingState) || 'unspecified';
    }
    if (
      normalized === 'unspecified' ||
      normalized === 'active' ||
      normalized === 'paused' ||
      normalized === 'suspended' ||
      normalized === 'deprecated' ||
      normalized === 'terminated'
    ) {
      return normalized as OfferingState;
    }
  }
  return 'unspecified';
}

function parseOfferingCategory(value: unknown): OfferingCategory {
  if (typeof value === 'number') {
    return OFFERING_CATEGORY_MAP[value] ?? 'other';
  }
  if (typeof value === 'string') {
    const normalized = value.toLowerCase();
    if (normalized.startsWith('offering_category_')) {
      const name = value.replace(/offering_category_/i, '').toLowerCase();
      return (name as OfferingCategory) || 'other';
    }
    if (
      normalized === 'compute' ||
      normalized === 'storage' ||
      normalized === 'network' ||
      normalized === 'hpc' ||
      normalized === 'gpu' ||
      normalized === 'ml' ||
      normalized === 'other'
    ) {
      return normalized as OfferingCategory;
    }
  }
  return 'other';
}

function parsePricingModel(value: unknown): PricingModel {
  if (typeof value === 'number') {
    return PRICING_MODEL_MAP[value] ?? 'hourly';
  }
  if (typeof value === 'string') {
    const normalized = value.toLowerCase();
    if (normalized.startsWith('pricing_model_')) {
      const name = value.replace(/pricing_model_/i, '').toLowerCase();
      return (name as PricingModel) || 'hourly';
    }
    if (
      normalized === 'hourly' ||
      normalized === 'daily' ||
      normalized === 'monthly' ||
      normalized === 'usage_based' ||
      normalized === 'fixed'
    ) {
      return normalized as PricingModel;
    }
  }
  return 'hourly';
}

function coerceInt(value: unknown, fallback = 0): number {
  if (typeof value === 'number') return Math.floor(value);
  if (typeof value === 'string') {
    const parsed = Number.parseInt(value, 10);
    return Number.isNaN(parsed) ? fallback : parsed;
  }
  return fallback;
}

function coerceString(value: unknown, fallback = ''): string {
  if (typeof value === 'string') return value;
  if (typeof value === 'number') return value.toString();
  return fallback;
}

function coerceDateString(value: unknown): string {
  if (!value) return new Date(0).toISOString();
  if (value instanceof Date) return value.toISOString();
  if (typeof value === 'string') return value;
  if (typeof value === 'number') {
    const ms = value > 1e12 ? value : value * 1000;
    return new Date(ms).toISOString();
  }
  return new Date(0).toISOString();
}

function parseOfferingId(rawId: unknown): OfferingID | null {
  if (!rawId) return null;
  if (typeof rawId === 'string') {
    const parts = rawId.split('/');
    if (parts.length === 2) {
      const sequence = Number.parseInt(parts[1], 10);
      if (!Number.isNaN(sequence)) {
        return { providerAddress: parts[0], sequence };
      }
    }
  }
  if (typeof rawId === 'object') {
    const record = rawId as Record<string, unknown>;
    const provider = coerceString(record.provider_address ?? record.providerAddress, '');
    const sequence = coerceInt(record.sequence, -1);
    if (provider && sequence >= 0) {
      return { providerAddress: provider, sequence };
    }
  }
  return null;
}

function parsePriceComponents(rawPrices: unknown): PriceComponent[] | undefined {
  if (!Array.isArray(rawPrices)) return undefined;
  const mapped = rawPrices
    .map((entry) => {
      if (!entry || typeof entry !== 'object') return null;
      const record = entry as Record<string, unknown>;
      const price = record.price as Record<string, unknown> | undefined;
      return {
        resourceType: coerceString(
          record.resource_type ?? record.resourceType,
          'cpu'
        ) as PriceComponent['resourceType'],
        unit: coerceString(record.unit, ''),
        price: {
          denom: coerceString(price?.denom, 'uve'),
          amount: coerceString(price?.amount, '0'),
        },
        usdReference: coerceString(record.usd_reference ?? record.usdReference, ''),
      } as PriceComponent;
    })
    .filter(Boolean) as PriceComponent[];

  return mapped.length > 0 ? mapped : undefined;
}

function parseUsageRates(raw: unknown): Record<string, string> | undefined {
  if (!raw || typeof raw !== 'object') return undefined;
  const result: Record<string, string> = {};
  Object.entries(raw as Record<string, unknown>).forEach(([key, value]) => {
    result[key] = coerceString(value, '0');
  });
  return Object.keys(result).length > 0 ? result : undefined;
}

function parseStringMap(raw: unknown): Record<string, string> | undefined {
  if (!raw || typeof raw !== 'object') return undefined;
  const result: Record<string, string> = {};
  Object.entries(raw as Record<string, unknown>).forEach(([key, value]) => {
    if (value === null || value === undefined) return;
    result[key] = coerceString(value, '');
  });
  return Object.keys(result).length > 0 ? result : undefined;
}

function mapOffering(raw: ChainOffering): Offering | null {
  const id = parseOfferingId(raw.id ?? raw.offering_id ?? raw.offeringId);
  const providerAddress =
    id?.providerAddress ?? coerceString(raw.provider_address ?? raw.providerAddress, '');
  const sequence = id?.sequence ?? coerceInt(raw.sequence, -1);
  if (!providerAddress || sequence < 0) return null;

  const pricing = (raw.pricing ?? {}) as Record<string, unknown>;
  const identityRequirement = (raw.identity_requirement ?? raw.identityRequirement ?? {}) as Record<
    string,
    unknown
  >;

  return {
    id: { providerAddress, sequence },
    state: parseOfferingState(raw.state),
    category: parseOfferingCategory(raw.category),
    name: coerceString(raw.name, 'Unnamed offering'),
    description: coerceString(raw.description, ''),
    version: coerceString(raw.version, ''),
    pricing: {
      model: parsePricingModel(pricing.model),
      basePrice: coerceString(pricing.base_price ?? pricing.basePrice, '0'),
      currency: coerceString(pricing.currency, 'uve'),
      usageRates: parseUsageRates(pricing.usage_rates ?? pricing.usageRates),
      minimumCommitment:
        typeof pricing.minimum_commitment === 'number'
          ? pricing.minimum_commitment
          : typeof pricing.minimumCommitment === 'number'
            ? pricing.minimumCommitment
            : undefined,
    },
    prices: parsePriceComponents(raw.prices),
    allowBidding: Boolean(raw.allow_bidding ?? raw.allowBidding),
    minBid:
      raw.min_bid || raw.minBid
        ? {
            denom: coerceString(
              ((raw.min_bid ?? raw.minBid) as Record<string, unknown>).denom,
              'uve'
            ),
            amount: coerceString(
              ((raw.min_bid ?? raw.minBid) as Record<string, unknown>).amount,
              '0'
            ),
          }
        : undefined,
    identityRequirement: {
      minScore: coerceInt(identityRequirement.min_score ?? identityRequirement.minScore, 0),
      requiredStatus: coerceString(
        identityRequirement.required_status ?? identityRequirement.requiredStatus,
        ''
      ),
      requireVerifiedEmail: Boolean(
        identityRequirement.require_verified_email ?? identityRequirement.requireVerifiedEmail
      ),
      requireVerifiedDomain: Boolean(
        identityRequirement.require_verified_domain ?? identityRequirement.requireVerifiedDomain
      ),
      requireMFA: Boolean(identityRequirement.require_mfa ?? identityRequirement.requireMFA),
    },
    requireMFAForOrders: Boolean(raw.require_mfa_for_orders ?? raw.requireMFAForOrders),
    publicMetadata: parseStringMap(raw.public_metadata ?? raw.publicMetadata),
    specifications: parseStringMap(raw.specifications),
    tags: Array.isArray(raw.tags) ? (raw.tags as string[]) : [],
    regions: Array.isArray(raw.regions) ? (raw.regions as string[]) : [],
    createdAt: coerceDateString(raw.created_at ?? raw.createdAt),
    updatedAt: coerceDateString(raw.updated_at ?? raw.updatedAt),
    activatedAt: raw.activated_at ? coerceDateString(raw.activated_at) : undefined,
    terminatedAt: raw.terminated_at ? coerceDateString(raw.terminated_at) : undefined,
    maxConcurrentOrders: raw.max_concurrent_orders
      ? coerceInt(raw.max_concurrent_orders, 0)
      : undefined,
    totalOrderCount: coerceInt(raw.total_order_count ?? raw.totalOrderCount, 0),
    activeOrderCount: coerceInt(raw.active_order_count ?? raw.activeOrderCount, 0),
  };
}

function extractOfferingsResponse(payload: unknown): {
  offerings: ChainOffering[];
  nextKey: string | null;
  total: number | null;
} {
  if (!payload || typeof payload !== 'object') {
    return { offerings: [], nextKey: null, total: null };
  }

  if (Array.isArray(payload)) {
    return { offerings: payload as ChainOffering[], nextKey: null, total: null };
  }

  const record = payload as ChainOfferingsResponse & {
    data?: ChainOfferingsResponse;
    result?: ChainOfferingsResponse;
  };

  const response = record.offerings
    ? record
    : record.data?.offerings
      ? record.data
      : record.result?.offerings
        ? record.result
        : record;

  const offerings = Array.isArray(response.offerings) ? response.offerings : [];
  const pagination = response.pagination ?? {};
  const nextKey =
    pagination.next_key ??
    pagination.nextKey ??
    response.next_key ??
    (response as unknown as { nextKey?: string }).nextKey ??
    null;
  const totalRaw = pagination.total ?? response.total ?? null;
  const total = totalRaw !== null && totalRaw !== undefined ? coerceInt(totalRaw, 0) : null;

  return { offerings, nextKey, total };
}

async function fetchChainJson<T>(path: string, params?: Record<string, string>) {
  const { restEndpoint } = getChainInfo();
  const url = new URL(path, restEndpoint);
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== '') {
        url.searchParams.set(key, value);
      }
    });
  }

  const response = await fetch(url.toString(), {
    headers: {
      Accept: 'application/json',
    },
  });

  if (!response.ok) {
    const text = await response.text();
    throw new ChainRequestError(
      response.status,
      text || `Chain request failed with status ${response.status}`
    );
  }

  return (await response.json()) as T;
}

async function fetchChainJsonWithFallback<T>(paths: string[], params?: Record<string, string>) {
  let lastError: Error | null = null;

  for (const path of paths) {
    try {
      return await fetchChainJson<T>(path, params);
    } catch (error) {
      lastError = error as Error;
      if (error instanceof ChainRequestError && (error.status === 404 || error.status === 501)) {
        continue;
      }
      break;
    }
  }

  throw lastError ?? new Error('Chain request failed');
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
  const address = coerceString(raw.owner ?? raw.address, '');
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
    createdAt: coerceString(raw.created_at ?? raw.createdAt, ''),
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
  if (filters.region !== 'all' && !offering.regions?.includes(filters.region)) return false;
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
      const offerings: Offering[] = [];

      let nextKey: string | null = null;
      let pageCount = 0;

      do {
        const params: Record<string, string> = {
          'pagination.limit': OFFERING_PAGE_LIMIT.toString(),
        };

        if (nextKey) {
          params['pagination.key'] = nextKey;
        }

        const payload = await fetchChainJsonWithFallback<unknown>(OFFERINGS_ENDPOINTS, params);
        const extracted = extractOfferingsResponse(payload);

        offerings.push(
          ...extracted.offerings
            .map((raw) => mapOffering(raw))
            .filter((item): item is Offering => item !== null)
        );

        nextKey = extracted.nextKey ?? null;
        pageCount += 1;
      } while (nextKey && pageCount < MAX_OFFERING_PAGES);

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
      const offeringId = `${providerAddress}/${sequence}`;
      const payload = await fetchChainJsonWithFallback<unknown>(OFFERING_ENDPOINTS(offeringId));
      const rawOffering =
        (payload as { offering?: ChainOffering }).offering ?? (payload as ChainOffering);
      const offering = mapOffering(rawOffering);

      if (!offering) {
        throw new Error('Offering not found');
      }

      set({ selectedOffering: offering, isLoadingDetail: false });
    } catch (error) {
      const message =
        error instanceof ChainRequestError && error.status === 404
          ? 'Offering not found'
          : error instanceof Error
            ? error.message
            : 'Failed to fetch offering';
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
      const payload = await fetchChainJsonWithFallback<unknown>(PROVIDER_ENDPOINTS(address));
      const rawProvider =
        (payload as { provider?: ChainProvider }).provider ?? (payload as ChainProvider);

      if (!rawProvider) {
        return null;
      }

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

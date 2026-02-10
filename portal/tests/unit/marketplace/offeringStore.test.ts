import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
  useOfferingStore,
  formatPrice,
  formatPriceUSD,
  getOfferingDisplayPrice,
  offeringKey,
} from '@/stores/offeringStore';

const mockOfferings = [
  {
    id: { provider_address: 'virtengine1provider1abc', sequence: 1 },
    state: 1,
    category: 'gpu',
    name: 'NVIDIA A100 Cluster',
    description: 'High-performance GPU cluster',
    version: '1.0.0',
    pricing: { model: 'hourly', base_price: '2500000', currency: 'uve' },
    prices: [
      {
        resource_type: 'gpu',
        unit: 'hour',
        price: { denom: 'uve', amount: '2500000' },
        usd_reference: '2.50',
      },
    ],
    allow_bidding: false,
    identity_requirement: {
      min_score: 50,
      required_status: '',
      require_verified_email: true,
      require_verified_domain: false,
      require_mfa: false,
    },
    require_mfa_for_orders: false,
    specifications: { cpu: '32 vCPU', memory: '128 GB', gpu_count: '8' },
    tags: ['gpu', 'nvidia'],
    regions: ['us-west', 'us-east'],
    created_at: '2024-01-15T10:00:00Z',
    updated_at: '2024-02-01T15:30:00Z',
    total_order_count: 156,
    active_order_count: 12,
  },
  {
    id: { provider_address: 'virtengine1provider2xyz', sequence: 1 },
    state: 1,
    category: 'compute',
    name: 'AMD EPYC 7763 Instance',
    description: 'CPU instance for parallel workloads',
    version: '1.2.0',
    pricing: { model: 'hourly', base_price: '450000', currency: 'uve' },
    prices: [
      {
        resource_type: 'cpu',
        unit: 'vcpu-hour',
        price: { denom: 'uve', amount: '15000' },
        usd_reference: '0.015',
      },
    ],
    allow_bidding: true,
    identity_requirement: {
      min_score: 0,
      required_status: '',
      require_verified_email: false,
      require_verified_domain: false,
      require_mfa: false,
    },
    require_mfa_for_orders: false,
    specifications: { cpu: '64 vCPU', memory: '256 GB' },
    tags: ['cpu', 'amd'],
    regions: ['eu-central'],
    created_at: '2024-01-20T08:00:00Z',
    updated_at: '2024-02-05T12:00:00Z',
    total_order_count: 89,
    active_order_count: 23,
  },
  {
    id: { provider_address: 'virtengine1provider3def', sequence: 1 },
    state: 1,
    category: 'storage',
    name: 'NVMe Block Storage',
    description: 'High-performance NVMe storage',
    version: '1.1.0',
    pricing: { model: 'monthly', base_price: '100000', currency: 'uve' },
    prices: [
      {
        resource_type: 'storage',
        unit: 'gb-month',
        price: { denom: 'uve', amount: '100000' },
        usd_reference: '0.10',
      },
    ],
    allow_bidding: false,
    identity_requirement: {
      min_score: 0,
      required_status: '',
      require_verified_email: false,
      require_verified_domain: false,
      require_mfa: false,
    },
    require_mfa_for_orders: false,
    specifications: { storage: '1000 GB' },
    tags: ['storage'],
    regions: ['us-central'],
    created_at: '2024-01-25T14:00:00Z',
    updated_at: '2024-02-08T09:00:00Z',
    total_order_count: 245,
    active_order_count: 67,
  },
];

const providerPayloads: Record<string, Record<string, unknown>> = {
  virtengine1provider1abc: {
    owner: 'virtengine1provider1abc',
    attributes: [
      { key: 'name', value: 'CloudCore' },
      { key: 'reputation', value: '95' },
      { key: 'region', value: 'us-west' },
      { key: 'description', value: 'Enterprise provider' },
    ],
    info: { website: 'https://cloudcore.example' },
  },
  virtengine1provider2xyz: {
    owner: 'virtengine1provider2xyz',
    attributes: [
      { key: 'name', value: 'DataNexus' },
      { key: 'reputation', value: '88' },
      { key: 'region', value: 'eu-central' },
    ],
    info: { website: 'https://datanexus.example' },
  },
  virtengine1provider3def: {
    owner: 'virtengine1provider3def',
    attributes: [
      { key: 'name', value: 'StorageWorks' },
      { key: 'reputation', value: '70' },
      { key: 'region', value: 'us-central' },
    ],
    info: { website: 'https://storageworks.example' },
  },
};

function createResponse(payload: unknown, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => payload,
    text: async () => JSON.stringify(payload),
  } as Response;
}

function setupFetchMock() {
  const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
    const url = new URL(typeof input === 'string' ? input : input.toString());
    const { pathname } = url;

    if (pathname.endsWith('/virtengine/market/v1/offerings')) {
      return createResponse({ offerings: mockOfferings });
    }

    if (pathname.includes('/virtengine/market/v1/offerings/')) {
      const segments = pathname.split('/');
      const offeringsIndex = segments.findIndex((segment) => segment === 'offerings');
      const offeringId = offeringsIndex >= 0 ? segments.slice(offeringsIndex + 1).join('/') : '';
      const offering = mockOfferings.find(
        (entry) => `${entry.id.provider_address}/${entry.id.sequence}` === offeringId
      );
      if (!offering) {
        return createResponse({ message: 'not found' }, 404);
      }
      return createResponse({ offering });
    }

    if (pathname.includes('/virtengine/provider/v1beta4/providers/')) {
      const address = pathname.split('/').pop() ?? '';
      const provider = providerPayloads[address];
      if (!provider) {
        return createResponse({ message: 'not found' }, 404);
      }
      return createResponse({ provider });
    }

    return createResponse({ message: 'not found' }, 404);
  });

  vi.stubGlobal('fetch', fetchMock);
}

describe('offeringStore', () => {
  beforeEach(() => {
    setupFetchMock();
    const { resetFilters, clearCompare, clearError } = useOfferingStore.getState();
    resetFilters();
    clearCompare();
    clearError();
    useOfferingStore.setState({ viewMode: 'grid', offerings: [], selectedOffering: null });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  describe('fetchOfferings', () => {
    it('loads offerings from chain response', async () => {
      await useOfferingStore.getState().fetchOfferings();

      const { offerings, isLoading, pagination } = useOfferingStore.getState();
      expect(isLoading).toBe(false);
      expect(offerings.length).toBeGreaterThan(0);
      expect(pagination.total).toBeGreaterThan(0);
    });

    it('filters by category', async () => {
      useOfferingStore.getState().setFilters({ category: 'gpu' });
      await useOfferingStore.getState().fetchOfferings();

      const { offerings } = useOfferingStore.getState();
      expect(offerings.length).toBeGreaterThan(0);
      offerings.forEach((o) => expect(o.category).toBe('gpu'));
    });

    it('filters by region', async () => {
      useOfferingStore.getState().setFilters({ region: 'us-central' });
      await useOfferingStore.getState().fetchOfferings();

      const { offerings } = useOfferingStore.getState();
      offerings.forEach((o) => expect(o.regions).toContain('us-central'));
    });

    it('filters by search term', async () => {
      useOfferingStore.getState().setFilters({ search: 'nvidia' });
      await useOfferingStore.getState().fetchOfferings();

      const { offerings } = useOfferingStore.getState();
      expect(offerings.length).toBeGreaterThan(0);
      offerings.forEach((o) => {
        const match =
          o.name.toLowerCase().includes('nvidia') ||
          o.description.toLowerCase().includes('nvidia') ||
          o.tags?.some((t) => t.toLowerCase().includes('nvidia'));
        expect(match).toBe(true);
      });
    });

    it('filters by minimum reputation', async () => {
      useOfferingStore.getState().setFilters({ minReputation: 90 });
      await useOfferingStore.getState().fetchOfferings();

      const { offerings } = useOfferingStore.getState();
      expect(offerings.length).toBeGreaterThan(0);
      offerings.forEach((o) => expect(o.id.providerAddress).toBe('virtengine1provider1abc'));
    });

    it('filters by provider search', async () => {
      useOfferingStore.getState().setFilters({ providerSearch: 'CloudCore' });
      await useOfferingStore.getState().fetchOfferings();

      const { offerings } = useOfferingStore.getState();
      expect(offerings.length).toBeGreaterThan(0);
      offerings.forEach((o) => expect(o.id.providerAddress).toBe('virtengine1provider1abc'));
    });

    it('sorts by price ascending', async () => {
      useOfferingStore.getState().setFilters({ sortBy: 'price', sortOrder: 'asc' });
      await useOfferingStore.getState().fetchOfferings();

      const { offerings } = useOfferingStore.getState();
      const priceValue = (amount?: string, basePrice?: string) => {
        if (amount) {
          const parsed = Number.parseFloat(amount);
          if (!Number.isNaN(parsed)) return parsed;
        }
        if (basePrice) {
          const parsed = Number.parseFloat(basePrice);
          if (!Number.isNaN(parsed)) return parsed / 1_000_000;
        }
        return 0;
      };
      for (let i = 1; i < offerings.length; i++) {
        const prev = priceValue(
          offerings[i - 1].prices?.[0]?.usdReference,
          offerings[i - 1].pricing.basePrice
        );
        const curr = priceValue(
          offerings[i].prices?.[0]?.usdReference,
          offerings[i].pricing.basePrice
        );
        expect(curr).toBeGreaterThanOrEqual(prev);
      }
    });

    it('sorts by price descending', async () => {
      useOfferingStore.getState().setFilters({ sortBy: 'price', sortOrder: 'desc' });
      await useOfferingStore.getState().fetchOfferings();

      const { offerings } = useOfferingStore.getState();
      const priceValue = (amount?: string, basePrice?: string) => {
        if (amount) {
          const parsed = Number.parseFloat(amount);
          if (!Number.isNaN(parsed)) return parsed;
        }
        if (basePrice) {
          const parsed = Number.parseFloat(basePrice);
          if (!Number.isNaN(parsed)) return parsed / 1_000_000;
        }
        return 0;
      };
      for (let i = 1; i < offerings.length; i++) {
        const prev = priceValue(
          offerings[i - 1].prices?.[0]?.usdReference,
          offerings[i - 1].pricing.basePrice
        );
        const curr = priceValue(
          offerings[i].prices?.[0]?.usdReference,
          offerings[i].pricing.basePrice
        );
        expect(curr).toBeLessThanOrEqual(prev);
      }
    });

    it('returns empty result for unmatched search', async () => {
      useOfferingStore.getState().setFilters({ search: 'xyznonexistent123' });
      await useOfferingStore.getState().fetchOfferings();

      const { offerings, pagination } = useOfferingStore.getState();
      expect(offerings.length).toBe(0);
      expect(pagination.total).toBe(0);
    });
  });

  describe('fetchOffering', () => {
    it('loads a single offering by provider and sequence', async () => {
      await useOfferingStore.getState().fetchOffering('virtengine1provider1abc', 1);

      const { selectedOffering, isLoadingDetail } = useOfferingStore.getState();
      expect(isLoadingDetail).toBe(false);
      expect(selectedOffering).not.toBeNull();
      expect(selectedOffering?.name).toBe('NVIDIA A100 Cluster');
    });

    it('sets error for non-existent offering', async () => {
      await useOfferingStore.getState().fetchOffering('nonexistent', 999);

      const { selectedOffering, error } = useOfferingStore.getState();
      expect(selectedOffering).toBeNull();
      expect(error).toBe('Offering not found');
    });
  });

  describe('fetchProvider', () => {
    it('returns provider info', async () => {
      const provider = await useOfferingStore.getState().fetchProvider('virtengine1provider1abc');
      expect(provider).not.toBeNull();
      expect(provider?.name).toBe('CloudCore');
    });

    it('returns null for unknown provider', async () => {
      const provider = await useOfferingStore.getState().fetchProvider('unknownaddress');
      expect(provider).toBeNull();
    });

    it('caches provider in store', async () => {
      await useOfferingStore.getState().fetchProvider('virtengine1provider2xyz');
      const { providers } = useOfferingStore.getState();
      expect(providers.has('virtengine1provider2xyz')).toBe(true);
    });
  });

  describe('filters', () => {
    it('sets partial filters and resets page', () => {
      useOfferingStore.getState().setPage(3);
      useOfferingStore.getState().setFilters({ category: 'hpc' });

      const { filters, pagination } = useOfferingStore.getState();
      expect(filters.category).toBe('hpc');
      expect(pagination.page).toBe(1);
    });

    it('resets filters to defaults', () => {
      useOfferingStore.getState().setFilters({ category: 'gpu', search: 'test' });
      useOfferingStore.getState().resetFilters();

      const { filters } = useOfferingStore.getState();
      expect(filters.category).toBe('all');
      expect(filters.search).toBe('');
      expect(filters.sortBy).toBe('name');
      expect(filters.sortOrder).toBe('asc');
    });
  });

  describe('comparison', () => {
    it('toggles offering into comparison', () => {
      useOfferingStore.getState().toggleCompare('provider1-1');

      const { compareIds } = useOfferingStore.getState();
      expect(compareIds).toContain('provider1-1');
    });

    it('toggles offering out of comparison', () => {
      useOfferingStore.getState().toggleCompare('provider1-1');
      useOfferingStore.getState().toggleCompare('provider1-1');

      const { compareIds } = useOfferingStore.getState();
      expect(compareIds).not.toContain('provider1-1');
    });

    it('limits comparison to 4 items', () => {
      useOfferingStore.getState().toggleCompare('a');
      useOfferingStore.getState().toggleCompare('b');
      useOfferingStore.getState().toggleCompare('c');
      useOfferingStore.getState().toggleCompare('d');
      useOfferingStore.getState().toggleCompare('e');

      const { compareIds } = useOfferingStore.getState();
      expect(compareIds.length).toBe(4);
      expect(compareIds).not.toContain('e');
    });

    it('clears comparison', () => {
      useOfferingStore.getState().toggleCompare('a');
      useOfferingStore.getState().toggleCompare('b');
      useOfferingStore.getState().clearCompare();

      const { compareIds } = useOfferingStore.getState();
      expect(compareIds.length).toBe(0);
    });
  });

  describe('view mode', () => {
    it('defaults to grid', () => {
      const { viewMode } = useOfferingStore.getState();
      expect(viewMode).toBe('grid');
    });

    it('switches to list', () => {
      useOfferingStore.getState().setViewMode('list');
      const { viewMode } = useOfferingStore.getState();
      expect(viewMode).toBe('list');
    });
  });

  describe('utils', () => {
    it('formats base price', () => {
      expect(formatPrice('2500000')).toBe('2.50');
    });

    it('formats USD reference', () => {
      expect(formatPriceUSD('2.5')).toBe('$2.50');
    });

    it('builds offering key', () => {
      const offering = useOfferingStore.getState().offerings[0];
      if (offering) {
        expect(offeringKey(offering)).toContain(offering.id.providerAddress);
      }
    });

    it('derives display price', () => {
      const offering = useOfferingStore.getState().offerings[0];
      if (offering) {
        const display = getOfferingDisplayPrice(offering);
        expect(display.amount).toBeDefined();
      }
    });
  });
});

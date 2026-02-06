import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
  useOfferingStore,
  formatPrice,
  formatPriceUSD,
  getOfferingDisplayPrice,
  offeringKey,
} from '@/stores/offeringStore';

describe('offeringStore', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    const { resetFilters, clearCompare, clearError } = useOfferingStore.getState();
    resetFilters();
    clearCompare();
    clearError();
    useOfferingStore.setState({ viewMode: 'grid', offerings: [], selectedOffering: null });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  // Helper to call an async store action and flush fake timers
  async function runWithTimers<T>(fn: () => Promise<T>): Promise<T> {
    const promise = fn();
    await vi.advanceTimersByTimeAsync(1000);
    return promise;
  }

  describe('fetchOfferings', () => {
    it('loads offerings from mock data', async () => {
      await runWithTimers(() => useOfferingStore.getState().fetchOfferings());

      const { offerings, isLoading, pagination } = useOfferingStore.getState();
      expect(isLoading).toBe(false);
      expect(offerings.length).toBeGreaterThan(0);
      expect(pagination.total).toBeGreaterThan(0);
    });

    it('filters by category', async () => {
      useOfferingStore.getState().setFilters({ category: 'gpu' });
      await runWithTimers(() => useOfferingStore.getState().fetchOfferings());

      const { offerings } = useOfferingStore.getState();
      expect(offerings.length).toBeGreaterThan(0);
      offerings.forEach((o) => expect(o.category).toBe('gpu'));
    });

    it('filters by region', async () => {
      useOfferingStore.getState().setFilters({ region: 'us-central' });
      await runWithTimers(() => useOfferingStore.getState().fetchOfferings());

      const { offerings } = useOfferingStore.getState();
      offerings.forEach((o) => expect(o.regions).toContain('us-central'));
    });

    it('filters by search term', async () => {
      useOfferingStore.getState().setFilters({ search: 'nvidia' });
      await runWithTimers(() => useOfferingStore.getState().fetchOfferings());

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
      await runWithTimers(() => useOfferingStore.getState().fetchOfferings());

      const { offerings } = useOfferingStore.getState();
      expect(offerings.length).toBeGreaterThan(0);
    });

    it('filters by provider search', async () => {
      useOfferingStore.getState().setFilters({ providerSearch: 'CloudCore' });
      await runWithTimers(() => useOfferingStore.getState().fetchOfferings());

      const { offerings } = useOfferingStore.getState();
      expect(offerings.length).toBeGreaterThan(0);
    });

    it('sorts by price ascending', async () => {
      useOfferingStore.getState().setFilters({ sortBy: 'price', sortOrder: 'asc' });
      await runWithTimers(() => useOfferingStore.getState().fetchOfferings());

      const { offerings } = useOfferingStore.getState();
      for (let i = 1; i < offerings.length; i++) {
        const prev = parseInt(offerings[i - 1].pricing.basePrice, 10);
        const curr = parseInt(offerings[i].pricing.basePrice, 10);
        expect(curr).toBeGreaterThanOrEqual(prev);
      }
    });

    it('sorts by price descending', async () => {
      useOfferingStore.getState().setFilters({ sortBy: 'price', sortOrder: 'desc' });
      await runWithTimers(() => useOfferingStore.getState().fetchOfferings());

      const { offerings } = useOfferingStore.getState();
      for (let i = 1; i < offerings.length; i++) {
        const prev = parseInt(offerings[i - 1].pricing.basePrice, 10);
        const curr = parseInt(offerings[i].pricing.basePrice, 10);
        expect(curr).toBeLessThanOrEqual(prev);
      }
    });

    it('returns empty result for unmatched search', async () => {
      useOfferingStore.getState().setFilters({ search: 'xyznonexistent123' });
      await runWithTimers(() => useOfferingStore.getState().fetchOfferings());

      const { offerings, pagination } = useOfferingStore.getState();
      expect(offerings.length).toBe(0);
      expect(pagination.total).toBe(0);
    });
  });

  describe('fetchOffering', () => {
    it('loads a single offering by provider and sequence', async () => {
      await runWithTimers(() =>
        useOfferingStore.getState().fetchOffering('virtengine1provider1abc', 1)
      );

      const { selectedOffering, isLoadingDetail } = useOfferingStore.getState();
      expect(isLoadingDetail).toBe(false);
      expect(selectedOffering).not.toBeNull();
      expect(selectedOffering?.name).toBe('NVIDIA A100 Cluster');
    });

    it('sets error for non-existent offering', async () => {
      await runWithTimers(() => useOfferingStore.getState().fetchOffering('nonexistent', 999));

      const { selectedOffering, error } = useOfferingStore.getState();
      expect(selectedOffering).toBeNull();
      expect(error).toBe('Offering not found');
    });
  });

  describe('fetchProvider', () => {
    it('returns provider info', async () => {
      const provider = await runWithTimers(() =>
        useOfferingStore.getState().fetchProvider('virtengine1provider1abc')
      );
      expect(provider).not.toBeNull();
      expect(provider?.name).toBe('CloudCore');
    });

    it('returns null for unknown provider', async () => {
      const provider = await runWithTimers(() =>
        useOfferingStore.getState().fetchProvider('unknownaddress')
      );
      expect(provider).toBeNull();
    });

    it('caches provider in store', async () => {
      await runWithTimers(() =>
        useOfferingStore.getState().fetchProvider('virtengine1provider2xyz')
      );
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

  describe('pagination', () => {
    it('sets page number', () => {
      useOfferingStore.getState().setPage(5);
      const { pagination } = useOfferingStore.getState();
      expect(pagination.page).toBe(5);
    });
  });

  describe('utility functions', () => {
    it('formats price correctly', () => {
      expect(formatPrice('2500000')).toBe('2.50');
    });

    it('formats USD reference', () => {
      expect(formatPriceUSD('2.50')).toBe('$2.50');
      expect(formatPriceUSD('0.005')).toBe('$0.0050');
      expect(formatPriceUSD(undefined)).toBe('—');
    });

    it('generates offering key', () => {
      const key = offeringKey({
        id: { providerAddress: 'addr1', sequence: 3 },
      } as any);
      expect(key).toBe('addr1-3');
    });

    it('gets display price from prices array', () => {
      const offering = {
        pricing: { model: 'hourly' as const, basePrice: '1000000', currency: 'uve' },
        prices: [
          {
            resourceType: 'gpu' as const,
            unit: 'hour',
            price: { denom: 'uve', amount: '2500000' },
            usdReference: '2.50',
          },
        ],
      } as any;

      const { amount, unit } = getOfferingDisplayPrice(offering);
      expect(amount).toBe('$2.50');
      expect(unit).toBe('/hour');
    });

    it('gets display price from base price when no prices array', () => {
      const offering = {
        pricing: { model: 'monthly' as const, basePrice: '1000000', currency: 'uve' },
      } as any;

      const { amount, unit } = getOfferingDisplayPrice(offering);
      expect(amount).toBe('1.00 VE');
      expect(unit).toBe('/month');
    });
  });
});

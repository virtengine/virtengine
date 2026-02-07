'use client';

import { useOfferingStore } from '@/stores/offeringStore';
import type { OfferingCategory, OfferingState } from '@/types/offerings';
import { CATEGORY_LABELS, CATEGORY_ICONS } from '@/types/offerings';

const REGIONS = [
  { value: 'all', label: 'All Regions' },
  { value: 'us-west', label: 'US West' },
  { value: 'us-east', label: 'US East' },
  { value: 'us-central', label: 'US Central' },
  { value: 'eu-west', label: 'EU West' },
  { value: 'eu-central', label: 'EU Central' },
  { value: 'asia-pacific', label: 'Asia Pacific' },
];

const REPUTATION_OPTIONS = [
  { value: 0, label: 'Any' },
  { value: 50, label: '50+' },
  { value: 70, label: '70+' },
  { value: 85, label: '85+' },
  { value: 95, label: '95+' },
];

const STATE_OPTIONS: Array<{ value: OfferingState | 'all'; label: string }> = [
  { value: 'all', label: 'All States' },
  { value: 'active', label: 'Active' },
  { value: 'paused', label: 'Paused' },
  { value: 'suspended', label: 'Suspended' },
  { value: 'deprecated', label: 'Deprecated' },
];

const PRICE_PRESETS = [
  { label: 'Any', min: 0, max: Infinity },
  { label: 'Under $1/hr', min: 0, max: 1 },
  { label: '$1 – $5/hr', min: 1, max: 5 },
  { label: '$5+/hr', min: 5, max: Infinity },
];

interface OfferingFiltersProps {
  className?: string;
}

export function OfferingFilters({ className = '' }: OfferingFiltersProps) {
  const { filters, setFilters, resetFilters, pagination } = useOfferingStore();

  const handleCategoryChange = (category: OfferingCategory | 'all') => {
    setFilters({ category });
  };

  const handleRegionChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setFilters({ region: e.target.value });
  };

  const handleReputationChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setFilters({ minReputation: parseInt(e.target.value, 10) });
  };

  const handleStateChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setFilters({ state: e.target.value as OfferingState | 'all' });
  };

  const handleProviderSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFilters({ providerSearch: e.target.value });
  };

  const handleCpuChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFilters({ minCpuCores: Number(e.target.value) || 0 });
  };

  const handleMemoryChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFilters({ minMemoryGB: Number(e.target.value) || 0 });
  };

  const handleGpuChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFilters({ minGpuCount: Number(e.target.value) || 0 });
  };

  const handlePricePreset = (min: number, max: number) => {
    if (min === 0 && max === Infinity) {
      setFilters({ priceRange: null });
    } else {
      setFilters({ priceRange: { min, max: max === Infinity ? 999999 : max } });
    }
  };

  const handleReset = () => {
    resetFilters();
  };

  const categories: Array<OfferingCategory | 'all'> = [
    'all',
    'compute',
    'gpu',
    'storage',
    'hpc',
    'ml',
    'network',
    'other',
  ];

  const hasActiveFilters =
    filters.category !== 'all' ||
    filters.region !== 'all' ||
    filters.minReputation > 0 ||
    filters.minCpuCores > 0 ||
    filters.minMemoryGB > 0 ||
    filters.minGpuCount > 0 ||
    filters.search !== '' ||
    filters.state !== 'active' ||
    filters.providerSearch !== '' ||
    filters.priceRange !== null;

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Category Filter */}
      <div>
        <h3 className="mb-3 font-medium">Resource Type</h3>
        <div className="space-y-1">
          {categories.map((category) => {
            const isAll = category === 'all';
            const icon = isAll ? '🔍' : CATEGORY_ICONS[category];
            const label = isAll ? 'All Types' : CATEGORY_LABELS[category];
            const isActive = filters.category === category;

            return (
              <button
                key={category}
                type="button"
                onClick={() => handleCategoryChange(category)}
                className={`flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm transition-colors ${
                  isActive ? 'bg-primary text-primary-foreground' : 'hover:bg-accent'
                }`}
              >
                <span>{icon}</span>
                <span>{label}</span>
              </button>
            );
          })}
        </div>
      </div>

      {/* State Filter */}
      <div>
        <h3 className="mb-3 font-medium">Status</h3>
        <select
          value={filters.state}
          onChange={handleStateChange}
          className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
        >
          {STATE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      {/* Region Filter */}
      <div>
        <h3 className="mb-3 font-medium">Region</h3>
        <select
          value={filters.region}
          onChange={handleRegionChange}
          className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
        >
          {REGIONS.map((region) => (
            <option key={region.value} value={region.value}>
              {region.label}
            </option>
          ))}
        </select>
      </div>

      {/* Price Range */}
      <div>
        <h3 className="mb-3 font-medium">Price Range</h3>
        <div className="space-y-1">
          {PRICE_PRESETS.map((preset) => {
            const isActive =
              (preset.min === 0 && preset.max === Infinity && filters.priceRange === null) ||
              (filters.priceRange?.min === preset.min &&
                (preset.max === Infinity
                  ? filters.priceRange?.max === 999999
                  : filters.priceRange?.max === preset.max));

            return (
              <button
                key={preset.label}
                type="button"
                onClick={() => handlePricePreset(preset.min, preset.max)}
                className={`flex w-full items-center rounded-md px-3 py-2 text-left text-sm transition-colors ${
                  isActive ? 'bg-primary text-primary-foreground' : 'hover:bg-accent'
                }`}
              >
                {preset.label}
              </button>
            );
          })}
        </div>
      </div>

      {/* Provider Search */}
      <div>
        <h3 className="mb-3 font-medium">Provider</h3>
        <input
          type="text"
          placeholder="Search providers..."
          value={filters.providerSearch}
          onChange={handleProviderSearch}
          className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground"
        />
      </div>

      {/* Resource Specs */}
      <div>
        <h3 className="mb-3 font-medium">Resource Specs</h3>
        <div className="grid gap-3">
          <label className="text-sm text-muted-foreground">
            Min CPU Cores
            <input
              type="number"
              min={0}
              value={filters.minCpuCores}
              onChange={handleCpuChange}
              className="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
            />
          </label>
          <label className="text-sm text-muted-foreground">
            Min RAM (GB)
            <input
              type="number"
              min={0}
              value={filters.minMemoryGB}
              onChange={handleMemoryChange}
              className="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
            />
          </label>
          <label className="text-sm text-muted-foreground">
            Min GPU Count
            <input
              type="number"
              min={0}
              value={filters.minGpuCount}
              onChange={handleGpuChange}
              className="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
            />
          </label>
        </div>
      </div>

      {/* Provider Reputation */}
      <div>
        <h3 className="mb-3 font-medium">Min. Provider Reputation</h3>
        <select
          value={filters.minReputation}
          onChange={handleReputationChange}
          className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
        >
          {REPUTATION_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      {/* Active Filters Summary */}
      {hasActiveFilters && (
        <div className="border-t border-border pt-4">
          <button
            type="button"
            onClick={handleReset}
            className="w-full rounded-md border border-border px-3 py-2 text-sm text-muted-foreground hover:bg-accent hover:text-foreground"
          >
            Clear all filters
          </button>
        </div>
      )}

      {/* Results count */}
      <div className="border-t border-border pt-4 text-sm text-muted-foreground">
        {pagination.total} offering{pagination.total !== 1 ? 's' : ''} found
      </div>
    </div>
  );
}

export function OfferingFiltersMobile({ onClose }: { onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm lg:hidden">
      <div className="fixed inset-y-0 right-0 w-full max-w-sm bg-background shadow-xl">
        <div className="flex h-full flex-col">
          <div className="flex items-center justify-between border-b border-border p-4">
            <h2 className="text-lg font-semibold">Filters</h2>
            <button type="button" onClick={onClose} className="rounded-md p-2 hover:bg-accent">
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
          <div className="flex-1 overflow-y-auto p-4">
            <OfferingFilters />
          </div>
          <div className="border-t border-border p-4">
            <button
              type="button"
              onClick={onClose}
              className="w-full rounded-md bg-primary px-4 py-2 font-medium text-primary-foreground hover:bg-primary/90"
            >
              Apply Filters
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

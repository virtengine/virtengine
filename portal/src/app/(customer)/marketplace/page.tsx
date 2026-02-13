'use client';

import { useEffect, useState } from 'react';
import { useOfferingStore } from '@/stores/offeringStore';
import {
  OfferingFilters,
  OfferingFiltersMobile,
  OfferingGrid,
  CompareBar,
} from '@/components/marketplace';
import type { OfferingSortField } from '@/types/offerings';

const SORT_OPTIONS: Array<{ value: OfferingSortField; label: string }> = [
  { value: 'name', label: 'Name' },
  { value: 'price', label: 'Price' },
  { value: 'reputation', label: 'Reputation' },
  { value: 'orders', label: 'Popularity' },
  { value: 'created', label: 'Newest' },
];

export default function MarketplacePage() {
  const { fetchOfferings, filters, setFilters, viewMode, setViewMode } = useOfferingStore();
  const [showMobileFilters, setShowMobileFilters] = useState(false);
  const [searchValue, setSearchValue] = useState(filters.search);

  useEffect(() => {
    void fetchOfferings();
  }, [fetchOfferings, filters]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setFilters({ search: searchValue });
  };

  const handleSortChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setFilters({ sortBy: e.target.value as OfferingSortField });
  };

  const handleSortOrderToggle = () => {
    setFilters({ sortOrder: filters.sortOrder === 'asc' ? 'desc' : 'asc' });
  };

  return (
    <div className="container py-8">
      {/* Header */}
      <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold">Marketplace</h1>
          <p className="mt-1 text-muted-foreground">
            Browse and purchase compute resources from providers worldwide
          </p>
        </div>

        {/* Search and mobile filter toggle */}
        <div className="flex gap-2">
          <form onSubmit={handleSearch} className="relative flex-1 sm:w-64">
            <input
              type="search"
              placeholder="Search offerings..."
              value={searchValue}
              onChange={(e) => setSearchValue(e.target.value)}
              className="w-full rounded-lg border border-border bg-background py-2 pl-9 pr-4 text-sm placeholder:text-muted-foreground focus:border-primary focus:outline-none"
            />
            <svg
              className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
          </form>

          <button
            type="button"
            onClick={() => setShowMobileFilters(true)}
            aria-label="Filters"
            title="Filters"
            className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent lg:hidden"
          >
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
              />
            </svg>
          </button>
        </div>
      </div>

      {/* Sort + View controls */}
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <label htmlFor="sort-select" className="text-sm text-muted-foreground">
            Sort by:
          </label>
          <select
            id="sort-select"
            value={filters.sortBy}
            onChange={handleSortChange}
            className="rounded-md border border-border bg-background px-3 py-1.5 text-sm"
          >
            {SORT_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={handleSortOrderToggle}
            className="rounded-md border border-border p-1.5 hover:bg-accent"
            title={filters.sortOrder === 'asc' ? 'Ascending' : 'Descending'}
          >
            <svg
              className={`h-4 w-4 transition-transform ${filters.sortOrder === 'desc' ? 'rotate-180' : ''}`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 15l7-7 7 7"
              />
            </svg>
          </button>
        </div>

        {/* View mode toggle */}
        <div className="flex rounded-lg border border-border">
          <button
            type="button"
            onClick={() => setViewMode('grid')}
            className={`rounded-l-md px-3 py-1.5 text-sm ${
              viewMode === 'grid' ? 'bg-primary text-primary-foreground' : 'hover:bg-accent'
            }`}
            title="Grid view"
          >
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"
              />
            </svg>
          </button>
          <button
            type="button"
            onClick={() => setViewMode('list')}
            className={`rounded-r-md px-3 py-1.5 text-sm ${
              viewMode === 'list' ? 'bg-primary text-primary-foreground' : 'hover:bg-accent'
            }`}
            title="List view"
          >
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 6h16M4 12h16M4 18h16"
              />
            </svg>
          </button>
        </div>
      </div>

      {/* Main content */}
      <div className="grid gap-6 lg:grid-cols-4">
        {/* Desktop Filters Sidebar */}
        <aside className="hidden lg:block">
          <div className="sticky top-4 rounded-lg border border-border p-4">
            <OfferingFilters />
          </div>
        </aside>

        {/* Offerings Grid */}
        <div className="lg:col-span-3">
          <OfferingGrid />
        </div>
      </div>

      {/* Mobile Filters */}
      {showMobileFilters && <OfferingFiltersMobile onClose={() => setShowMobileFilters(false)} />}

      {/* Compare Bar */}
      <CompareBar />
    </div>
  );
}

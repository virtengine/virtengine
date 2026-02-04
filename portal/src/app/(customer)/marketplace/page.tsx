'use client';

import { useEffect, useState } from 'react';
import { useOfferingStore } from '@/stores/offeringStore';
import { OfferingFilters, OfferingFiltersMobile, OfferingGrid } from '@/components/marketplace';

export default function MarketplacePage() {
  const { fetchOfferings, filters, setFilters } = useOfferingStore();
  const [showMobileFilters, setShowMobileFilters] = useState(false);
  const [searchValue, setSearchValue] = useState(filters.search);

  useEffect(() => {
    fetchOfferings();
  }, [fetchOfferings, filters]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setFilters({ search: searchValue });
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
            className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent lg:hidden"
          >
            <svg
              className="h-5 w-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
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
      {showMobileFilters && (
        <OfferingFiltersMobile onClose={() => setShowMobileFilters(false)} />
      )}
    </div>
  );
}

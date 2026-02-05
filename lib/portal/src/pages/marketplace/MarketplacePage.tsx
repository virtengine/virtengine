/**
 * Marketplace Page Component
 * VE-703: Main customer marketplace browse page
 */
import * as React from 'react';
import { useState, useCallback, useEffect } from 'react';
import type { Offering, OfferingFilter, OfferingSort, OfferingType } from '../../../types/marketplace';
import type { QueryClient } from '../../../types/chain';
import { useOfferings, OFFERING_CATEGORIES } from './hooks/useOfferings';
import { SearchBar } from './components/SearchBar';
import { FilterPanel } from './components/FilterPanel';
import { CategoryNav } from './components/CategoryNav';
import { OfferingGrid } from './components/OfferingGrid';
import { OfferingDetailPage } from './components/OfferingDetailPage';

export interface MarketplacePageProps {
  queryClient: QueryClient;
  onCheckout?: (offeringId: string, durationSeconds: number) => void;
  initialCategory?: string;
  initialQuery?: string;
  className?: string;
}

type ViewMode = 'browse' | 'detail';

export function MarketplacePage({
  queryClient,
  onCheckout,
  initialCategory = 'all',
  initialQuery = '',
  className = '',
}: MarketplacePageProps): JSX.Element {
  // View state
  const [viewMode, setViewMode] = useState<ViewMode>('browse');
  const [selectedOffering, setSelectedOffering] = useState<Offering | null>(null);

  // Filter state
  const [searchQuery, setSearchQuery] = useState(initialQuery);
  const [selectedCategory, setSelectedCategory] = useState(initialCategory);
  const [filter, setFilter] = useState<OfferingFilter>({});
  const [showMobileFilters, setShowMobileFilters] = useState(false);

  // Offerings hook
  const { state, actions, sort } = useOfferings({
    queryClient,
    pageSize: 20,
    initialFilter: {},
    initialSort: { field: 'reliability_score', direction: 'desc' },
  });

  // Initial load
  useEffect(() => {
    const category = OFFERING_CATEGORIES.find((c) => c.id === initialCategory);
    const initialFilter: OfferingFilter = {
      query: initialQuery || undefined,
      types: category?.types.length ? category.types : undefined,
    };
    actions.search(initialFilter);
  }, [initialCategory, initialQuery, actions]);

  // Handle search
  const handleSearch = useCallback(
    (query: string) => {
      const newFilter: OfferingFilter = {
        ...filter,
        query: query || undefined,
      };
      setFilter(newFilter);
      actions.search(newFilter, sort);
    },
    [filter, sort, actions]
  );

  // Handle category change
  const handleCategoryChange = useCallback(
    (categoryId: string, types: OfferingType[]) => {
      setSelectedCategory(categoryId);
      const newFilter: OfferingFilter = {
        ...filter,
        types: types.length > 0 ? types : undefined,
      };
      setFilter(newFilter);
      actions.search(newFilter, sort);
    },
    [filter, sort, actions]
  );

  // Handle filter change
  const handleFilterChange = useCallback((newFilter: OfferingFilter) => {
    setFilter(newFilter);
  }, []);

  // Handle filter apply
  const handleFilterApply = useCallback(() => {
    actions.search(filter, sort);
    setShowMobileFilters(false);
  }, [filter, sort, actions]);

  // Handle filter reset
  const handleFilterReset = useCallback(() => {
    const newFilter: OfferingFilter = {
      query: filter.query,
    };
    setFilter(newFilter);
    setSelectedCategory('all');
    actions.search(newFilter, sort);
  }, [filter.query, sort, actions]);

  // Handle sort change
  const handleSortChange = useCallback(
    (newSort: OfferingSort) => {
      actions.search(filter, newSort);
    },
    [filter, actions]
  );

  // Handle page change
  const handlePageChange = useCallback(
    (page: number) => {
      actions.search(filter, sort, page);
    },
    [filter, sort, actions]
  );

  // Handle offering select
  const handleOfferingSelect = useCallback((offering: Offering) => {
    setSelectedOffering(offering);
    setViewMode('detail');
  }, []);

  // Handle back to browse
  const handleBackToBrowse = useCallback(() => {
    setSelectedOffering(null);
    setViewMode('browse');
  }, []);

  // Handle checkout
  const handleCheckout = useCallback(
    (durationSeconds: number) => {
      if (selectedOffering && onCheckout) {
        onCheckout(selectedOffering.id, durationSeconds);
      }
    },
    [selectedOffering, onCheckout]
  );

  // Render detail view
  if (viewMode === 'detail' && selectedOffering) {
    return (
      <OfferingDetailPage
        offering={selectedOffering}
        onBack={handleBackToBrowse}
        onCheckout={handleCheckout}
        className={className}
      />
    );
  }

  // Render browse view
  return (
    <div className={`marketplace-page ${className}`}>
      {/* Header */}
      <header className="marketplace-page__header">
        <div className="marketplace-page__title-row">
          <h1 className="marketplace-page__title">Marketplace</h1>
          <p className="marketplace-page__subtitle">
            Discover and deploy compute, storage, and AI resources from verified providers
          </p>
        </div>

        <div className="marketplace-page__search-row">
          <SearchBar
            value={searchQuery}
            onChange={setSearchQuery}
            onSearch={handleSearch}
            isLoading={state.isLoading}
            placeholder="Search offerings by name, provider, or specs..."
          />

          {/* Mobile filter toggle */}
          <button
            type="button"
            className="marketplace-page__filter-toggle"
            onClick={() => setShowMobileFilters(!showMobileFilters)}
            aria-expanded={showMobileFilters}
            aria-controls="marketplace-filters"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="4" y1="21" x2="4" y2="14" />
              <line x1="4" y1="10" x2="4" y2="3" />
              <line x1="12" y1="21" x2="12" y2="12" />
              <line x1="12" y1="8" x2="12" y2="3" />
              <line x1="20" y1="21" x2="20" y2="16" />
              <line x1="20" y1="12" x2="20" y2="3" />
              <line x1="1" y1="14" x2="7" y2="14" />
              <line x1="9" y1="8" x2="15" y2="8" />
              <line x1="17" y1="16" x2="23" y2="16" />
            </svg>
            Filters
          </button>
        </div>
      </header>

      {/* Main Content */}
      <div className="marketplace-page__content">
        {/* Sidebar */}
        <aside
          id="marketplace-filters"
          className={`marketplace-page__sidebar ${showMobileFilters ? 'marketplace-page__sidebar--open' : ''}`}
        >
          {/* Mobile close button */}
          <button
            type="button"
            className="marketplace-page__sidebar-close"
            onClick={() => setShowMobileFilters(false)}
            aria-label="Close filters"
          >
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>

          <CategoryNav
            selectedCategory={selectedCategory}
            onCategoryChange={handleCategoryChange}
          />

          <FilterPanel
            filter={filter}
            onChange={handleFilterChange}
            onApply={handleFilterApply}
            onReset={handleFilterReset}
          />
        </aside>

        {/* Mobile overlay */}
        {showMobileFilters && (
          <div
            className="marketplace-page__overlay"
            onClick={() => setShowMobileFilters(false)}
            aria-hidden="true"
          />
        )}

        {/* Main grid */}
        <main className="marketplace-page__main">
          {state.error ? (
            <div className="marketplace-page__error" role="alert">
              <p>Failed to load offerings: {state.error}</p>
              <button
                type="button"
                onClick={() => actions.refresh()}
              >
                Try Again
              </button>
            </div>
          ) : (
            <OfferingGrid
              offerings={state.offerings}
              totalCount={state.totalCount}
              page={state.page}
              pageSize={state.pageSize}
              isLoading={state.isLoading}
              hasMore={state.hasMore}
              sort={sort}
              onOfferingSelect={handleOfferingSelect}
              onSortChange={handleSortChange}
              onPageChange={handlePageChange}
              onLoadMore={actions.loadMore}
            />
          )}
        </main>
      </div>

      <style>{marketplacePageStyles}</style>
    </div>
  );
}

const marketplacePageStyles = `
  .marketplace-page {
    min-height: 100vh;
    background: #f9fafb;
  }

  .marketplace-page__header {
    background: white;
    border-bottom: 1px solid #e5e7eb;
    padding: 24px;
  }

  .marketplace-page__title-row {
    max-width: 1400px;
    margin: 0 auto 16px;
  }

  .marketplace-page__title {
    margin: 0;
    font-size: 1.75rem;
    font-weight: 700;
    color: #111827;
  }

  .marketplace-page__subtitle {
    margin: 4px 0 0;
    font-size: 0.875rem;
    color: #6b7280;
  }

  .marketplace-page__search-row {
    display: flex;
    gap: 12px;
    max-width: 1400px;
    margin: 0 auto;
  }

  .marketplace-page__filter-toggle {
    display: none;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
    cursor: pointer;
    white-space: nowrap;
  }

  .marketplace-page__filter-toggle:hover {
    border-color: #d1d5db;
  }

  .marketplace-page__content {
    display: flex;
    gap: 24px;
    max-width: 1400px;
    margin: 0 auto;
    padding: 24px;
  }

  .marketplace-page__sidebar {
    display: flex;
    flex-direction: column;
    gap: 16px;
    width: 280px;
    flex-shrink: 0;
  }

  .marketplace-page__sidebar-close {
    display: none;
  }

  .marketplace-page__overlay {
    display: none;
  }

  .marketplace-page__main {
    flex: 1;
    min-width: 0;
  }

  .marketplace-page__error {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    padding: 48px 24px;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    text-align: center;
  }

  .marketplace-page__error p {
    margin: 0;
    color: #dc2626;
  }

  .marketplace-page__error button {
    padding: 10px 20px;
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
  }

  .marketplace-page__error button:hover {
    background: #2563eb;
  }

  /* Mobile responsive */
  @media (max-width: 1024px) {
    .marketplace-page__filter-toggle {
      display: flex;
    }

    .marketplace-page__sidebar {
      position: fixed;
      top: 0;
      left: 0;
      bottom: 0;
      width: 320px;
      max-width: 100%;
      background: #f9fafb;
      padding: 16px;
      transform: translateX(-100%);
      transition: transform 0.3s ease;
      z-index: 100;
      overflow-y: auto;
    }

    .marketplace-page__sidebar--open {
      transform: translateX(0);
    }

    .marketplace-page__sidebar-close {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 40px;
      height: 40px;
      margin-bottom: 8px;
      background: transparent;
      border: none;
      border-radius: 8px;
      color: #6b7280;
      cursor: pointer;
      align-self: flex-end;
    }

    .marketplace-page__sidebar-close:hover {
      background: #f3f4f6;
      color: #111827;
    }

    .marketplace-page__overlay {
      display: block;
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba(0, 0, 0, 0.3);
      z-index: 99;
    }
  }

  @media (max-width: 640px) {
    .marketplace-page__header {
      padding: 16px;
    }

    .marketplace-page__content {
      padding: 16px;
    }

    .marketplace-page__search-row {
      flex-direction: column;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .marketplace-page__sidebar {
      transition: none;
    }
  }
`;

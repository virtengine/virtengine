'use client';

import { useOfferingStore, offeringKey } from '@/stores/offeringStore';
import { OfferingCard, OfferingCardSkeleton } from './OfferingCard';
import { OfferingListItem, OfferingListItemSkeleton } from './OfferingListView';

export function OfferingGrid() {
  const { offerings, isLoading, error, pagination, setPage, viewMode, compareIds, toggleCompare } =
    useOfferingStore();

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center rounded-lg border border-red-200 bg-red-50 p-8 text-center dark:border-red-800 dark:bg-red-900/20">
        <svg
          className="mb-4 h-12 w-12 text-red-500"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
        <h3 className="text-lg font-semibold text-red-700 dark:text-red-300">
          Failed to load offerings
        </h3>
        <p className="mt-2 text-sm text-red-600 dark:text-red-400">{error}</p>
        <button
          type="button"
          onClick={() => {
            void useOfferingStore.getState().fetchOfferings();
          }}
          className="mt-4 rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
        >
          Try Again
        </button>
      </div>
    );
  }

  if (isLoading) {
    return viewMode === 'list' ? (
      <div className="space-y-3">
        {Array.from({ length: 6 }, (_, idx) => `skeleton-${idx}`).map((key) => (
          <OfferingListItemSkeleton key={key} />
        ))}
      </div>
    ) : (
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {Array.from({ length: 6 }, (_, idx) => `skeleton-${idx}`).map((key) => (
          <OfferingCardSkeleton key={key} />
        ))}
      </div>
    );
  }

  if (offerings.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-lg border border-border p-12 text-center">
        <svg
          className="mb-4 h-16 w-16 text-muted-foreground"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.5}
            d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
          />
        </svg>
        <h3 className="text-lg font-semibold">No offerings found</h3>
        <p className="mt-2 text-sm text-muted-foreground">
          Try adjusting your filters or search terms
        </p>
        <button
          type="button"
          onClick={() => useOfferingStore.getState().resetFilters()}
          className="mt-4 rounded-md border border-border px-4 py-2 text-sm hover:bg-accent"
        >
          Clear Filters
        </button>
      </div>
    );
  }

  const totalPages = Math.ceil(pagination.total / pagination.pageSize);

  return (
    <div>
      {viewMode === 'list' ? (
        <div className="space-y-3">
          {offerings.map((offering) => (
            <OfferingListItem
              key={`${offering.id.providerAddress}-${offering.id.sequence}`}
              offering={offering}
              isComparing={compareIds.includes(offeringKey(offering))}
              onToggleCompare={toggleCompare}
            />
          ))}
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {offerings.map((offering) => (
            <OfferingCard
              key={`${offering.id.providerAddress}-${offering.id.sequence}`}
              offering={offering}
            />
          ))}
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="mt-8 flex items-center justify-center gap-2">
          <button
            type="button"
            onClick={() => setPage(pagination.page - 1)}
            disabled={pagination.page === 1}
            className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
          >
            Previous
          </button>

          {Array.from({ length: totalPages }).map((_, i) => {
            const pageNum = i + 1;
            const isActive = pageNum === pagination.page;

            // Show first, last, and nearby pages
            if (
              pageNum === 1 ||
              pageNum === totalPages ||
              Math.abs(pageNum - pagination.page) <= 1
            ) {
              return (
                <button
                  key={pageNum}
                  type="button"
                  onClick={() => setPage(pageNum)}
                  className={`rounded-lg px-4 py-2 text-sm ${
                    isActive
                      ? 'bg-primary text-primary-foreground'
                      : 'border border-border hover:bg-accent'
                  }`}
                >
                  {pageNum}
                </button>
              );
            }

            // Show ellipsis
            if (pageNum === pagination.page - 2 || pageNum === pagination.page + 2) {
              return (
                <span key={pageNum} className="px-2 text-muted-foreground">
                  ...
                </span>
              );
            }

            return null;
          })}

          <button
            type="button"
            onClick={() => setPage(pagination.page + 1)}
            disabled={pagination.page === totalPages}
            className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}

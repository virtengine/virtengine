/**
 * Offering Grid Component
 * VE-703: Enhanced marketplace offering display with layout toggle and pagination
 */
import * as React from "react";
import { useState, useCallback } from "react";
import type {
  Offering,
  OfferingSort,
  OfferingSortField,
} from "../../../../types/marketplace";
import { MarketplaceOfferingCard } from "./MarketplaceOfferingCard";

export interface OfferingGridProps {
  offerings: Offering[];
  totalCount: number;
  page: number;
  pageSize: number;
  isLoading: boolean;
  hasMore: boolean;
  sort: OfferingSort;
  onOfferingSelect: (offering: Offering) => void;
  onSortChange: (sort: OfferingSort) => void;
  onPageChange: (page: number) => void;
  onLoadMore?: () => void;
  className?: string;
}

type LayoutMode = "grid" | "list";

export function OfferingGrid({
  offerings,
  totalCount,
  page,
  pageSize,
  isLoading,
  hasMore,
  sort,
  onOfferingSelect,
  onSortChange,
  onPageChange,
  onLoadMore,
  className = "",
}: OfferingGridProps): JSX.Element {
  const [layout, setLayout] = useState<LayoutMode>("grid");

  const handleSortFieldChange = useCallback(
    (e: React.ChangeEvent<HTMLSelectElement>) => {
      onSortChange({
        field: e.target.value as OfferingSortField,
        direction: sort.direction,
      });
    },
    [sort.direction, onSortChange],
  );

  const handleSortDirectionToggle = useCallback(() => {
    onSortChange({
      field: sort.field,
      direction: sort.direction === "asc" ? "desc" : "asc",
    });
  }, [sort, onSortChange]);

  const totalPages = Math.ceil(totalCount / pageSize);
  const startItem = (page - 1) * pageSize + 1;
  const endItem = Math.min(page * pageSize, totalCount);

  return (
    <div className={`offering-grid ${className}`}>
      {/* Toolbar */}
      <div className="offering-grid__toolbar">
        <div className="offering-grid__info">
          {totalCount > 0 ? (
            <span>
              Showing{" "}
              <strong>
                {startItem}-{endItem}
              </strong>{" "}
              of <strong>{totalCount}</strong> offerings
            </span>
          ) : (
            <span>No offerings found</span>
          )}
        </div>

        <div className="offering-grid__controls">
          {/* Sort */}
          <div className="offering-grid__sort">
            <label htmlFor="sort-field" className="sr-only">
              Sort by
            </label>
            <select
              id="sort-field"
              className="offering-grid__sort-select"
              value={sort.field}
              onChange={handleSortFieldChange}
            >
              <option value="reliability_score">Reliability</option>
              <option value="price">Price</option>
              <option value="created_at">Newest</option>
              <option value="cpu_score">CPU Score</option>
              <option value="gpu_score">GPU Score</option>
              <option value="provider_name">Provider</option>
            </select>
            <button
              type="button"
              className="offering-grid__sort-direction"
              onClick={handleSortDirectionToggle}
              aria-label={`Sort ${sort.direction === "asc" ? "ascending" : "descending"}`}
              title={sort.direction === "asc" ? "Ascending" : "Descending"}
            >
              {sort.direction === "asc" ? (
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M12 19V5M5 12l7-7 7 7" />
                </svg>
              ) : (
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M12 5v14M19 12l-7 7-7-7" />
                </svg>
              )}
            </button>
          </div>

          {/* Layout Toggle */}
          <div
            className="offering-grid__layout"
            role="group"
            aria-label="Layout"
          >
            <button
              type="button"
              className={`offering-grid__layout-btn ${layout === "grid" ? "offering-grid__layout-btn--active" : ""}`}
              onClick={() => setLayout("grid")}
              aria-pressed={layout === "grid"}
              title="Grid view"
            >
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <rect x="3" y="3" width="7" height="7" />
                <rect x="14" y="3" width="7" height="7" />
                <rect x="14" y="14" width="7" height="7" />
                <rect x="3" y="14" width="7" height="7" />
              </svg>
            </button>
            <button
              type="button"
              className={`offering-grid__layout-btn ${layout === "list" ? "offering-grid__layout-btn--active" : ""}`}
              onClick={() => setLayout("list")}
              aria-pressed={layout === "list"}
              title="List view"
            >
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <line x1="8" y1="6" x2="21" y2="6" />
                <line x1="8" y1="12" x2="21" y2="12" />
                <line x1="8" y1="18" x2="21" y2="18" />
                <line x1="3" y1="6" x2="3.01" y2="6" />
                <line x1="3" y1="12" x2="3.01" y2="12" />
                <line x1="3" y1="18" x2="3.01" y2="18" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      {/* Grid/List Content */}
      {isLoading && offerings.length === 0 ? (
        <div
          className="offering-grid__loading"
          role="status"
          aria-live="polite"
        >
          <LoadingSpinner />
          <span>Loading offerings...</span>
        </div>
      ) : offerings.length === 0 ? (
        <div className="offering-grid__empty" role="status">
          <EmptyState />
        </div>
      ) : (
        <>
          <div
            className={`offering-grid__content offering-grid__content--${layout}`}
            role="list"
            aria-label="Offerings"
          >
            {offerings.map((offering) => (
              <div key={offering.id} role="listitem">
                <MarketplaceOfferingCard
                  offering={offering}
                  onSelect={onOfferingSelect}
                />
              </div>
            ))}
          </div>

          {/* Pagination or Load More */}
          {totalPages > 1 && (
            <div className="offering-grid__pagination">
              {onLoadMore ? (
                <button
                  type="button"
                  className="offering-grid__load-more"
                  onClick={onLoadMore}
                  disabled={!hasMore || isLoading}
                >
                  {isLoading ? (
                    <>
                      <LoadingSpinner size={16} /> Loading...
                    </>
                  ) : hasMore ? (
                    "Load More"
                  ) : (
                    "No more offerings"
                  )}
                </button>
              ) : (
                <Pagination
                  page={page}
                  totalPages={totalPages}
                  onPageChange={onPageChange}
                />
              )}
            </div>
          )}
        </>
      )}

      <style>{offeringGridStyles}</style>
    </div>
  );
}

interface PaginationProps {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

function Pagination({
  page,
  totalPages,
  onPageChange,
}: PaginationProps): JSX.Element {
  const pages = getPageNumbers(page, totalPages);

  return (
    <nav className="pagination" aria-label="Pagination">
      <button
        type="button"
        className="pagination__btn"
        onClick={() => onPageChange(page - 1)}
        disabled={page === 1}
        aria-label="Previous page"
      >
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M15 18l-6-6 6-6" />
        </svg>
      </button>

      {pages.map((p, i) =>
        p === "..." ? (
          <span key={`ellipsis-${i}`} className="pagination__ellipsis">
            ...
          </span>
        ) : (
          <button
            key={p}
            type="button"
            className={`pagination__btn pagination__btn--page ${page === p ? "pagination__btn--active" : ""}`}
            onClick={() => onPageChange(p as number)}
            aria-label={`Page ${p}`}
            aria-current={page === p ? "page" : undefined}
          >
            {p}
          </button>
        ),
      )}

      <button
        type="button"
        className="pagination__btn"
        onClick={() => onPageChange(page + 1)}
        disabled={page === totalPages}
        aria-label="Next page"
      >
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M9 18l6-6-6-6" />
        </svg>
      </button>
    </nav>
  );
}

function getPageNumbers(current: number, total: number): (number | string)[] {
  const pages: (number | string)[] = [];
  const delta = 2;

  for (let i = 1; i <= total; i++) {
    if (
      i === 1 ||
      i === total ||
      (i >= current - delta && i <= current + delta)
    ) {
      pages.push(i);
    } else if (pages[pages.length - 1] !== "...") {
      pages.push("...");
    }
  }

  return pages;
}

function LoadingSpinner({ size = 24 }: { size?: number }): JSX.Element {
  return (
    <svg
      className="offering-grid__spinner"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
    >
      <circle cx="12" cy="12" r="10" opacity="0.25" />
      <path d="M12 2a10 10 0 0 1 10 10" />
    </svg>
  );
}

function EmptyState(): JSX.Element {
  return (
    <div className="empty-state">
      <div className="empty-state__icon" aria-hidden="true">
        <svg
          width="64"
          height="64"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="1"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <circle cx="11" cy="11" r="8" />
          <path d="M21 21l-4.35-4.35" />
        </svg>
      </div>
      <h3 className="empty-state__title">No offerings found</h3>
      <p className="empty-state__description">
        Try adjusting your filters or search query to find what you're looking
        for.
      </p>
    </div>
  );
}

const offeringGridStyles = `
  .offering-grid {
    flex: 1;
    min-width: 0;
  }

  .offering-grid__toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    flex-wrap: wrap;
    gap: 12px;
  }

  .offering-grid__info {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .offering-grid__info strong {
    color: #111827;
  }

  .offering-grid__controls {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .offering-grid__sort {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .offering-grid__sort-select {
    padding: 8px 12px;
    padding-right: 32px;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 0.875rem;
    color: #374151;
    cursor: pointer;
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%236b7280' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 8px center;
  }

  .offering-grid__sort-select:focus {
    outline: none;
    border-color: #3b82f6;
  }

  .offering-grid__sort-direction {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 8px;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    color: #6b7280;
    cursor: pointer;
    transition: border-color 0.2s, color 0.2s;
  }

  .offering-grid__sort-direction:hover {
    border-color: #3b82f6;
    color: #3b82f6;
  }

  .offering-grid__layout {
    display: flex;
    background: #f3f4f6;
    border-radius: 8px;
    padding: 2px;
  }

  .offering-grid__layout-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 6px 10px;
    background: transparent;
    border: none;
    border-radius: 6px;
    color: #6b7280;
    cursor: pointer;
    transition: background-color 0.2s, color 0.2s;
  }

  .offering-grid__layout-btn:hover {
    color: #374151;
  }

  .offering-grid__layout-btn--active {
    background: white;
    color: #111827;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  }

  .offering-grid__content {
    display: grid;
    gap: 16px;
  }

  .offering-grid__content--grid {
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  }

  .offering-grid__content--list {
    grid-template-columns: 1fr;
  }

  .offering-grid__loading,
  .offering-grid__empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 48px 24px;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    text-align: center;
    color: #6b7280;
    gap: 16px;
  }

  .offering-grid__spinner {
    animation: spin 1s linear infinite;
    color: #3b82f6;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  .offering-grid__pagination {
    display: flex;
    justify-content: center;
    margin-top: 24px;
  }

  .offering-grid__load-more {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 24px;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
    cursor: pointer;
    transition: border-color 0.2s, background-color 0.2s;
  }

  .offering-grid__load-more:hover:not(:disabled) {
    border-color: #3b82f6;
    background: #f8fafc;
  }

  .offering-grid__load-more:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Pagination styles */
  .pagination {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .pagination__btn {
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 36px;
    height: 36px;
    padding: 0 8px;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 0.875rem;
    color: #374151;
    cursor: pointer;
    transition: border-color 0.2s, background-color 0.2s;
  }

  .pagination__btn:hover:not(:disabled) {
    border-color: #3b82f6;
    background: #f8fafc;
  }

  .pagination__btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .pagination__btn--active {
    background: #3b82f6;
    border-color: #3b82f6;
    color: white;
  }

  .pagination__btn--active:hover {
    background: #2563eb;
    border-color: #2563eb;
  }

  .pagination__ellipsis {
    padding: 0 8px;
    color: #9ca3af;
  }

  /* Empty state */
  .empty-state__icon {
    color: #d1d5db;
  }

  .empty-state__title {
    margin: 0;
    font-size: 1.125rem;
    font-weight: 600;
    color: #111827;
  }

  .empty-state__description {
    margin: 0;
    font-size: 0.875rem;
    color: #6b7280;
    max-width: 320px;
  }

  /* Screen reader only */
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  @media (prefers-reduced-motion: reduce) {
    .offering-grid__sort-direction,
    .offering-grid__layout-btn,
    .offering-grid__load-more,
    .pagination__btn {
      transition: none;
    }
    .offering-grid__spinner {
      animation: none;
    }
  }

  @media (max-width: 640px) {
    .offering-grid__toolbar {
      flex-direction: column;
      align-items: stretch;
    }

    .offering-grid__controls {
      justify-content: space-between;
    }

    .offering-grid__content--grid {
      grid-template-columns: 1fr;
    }
  }
`;

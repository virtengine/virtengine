/**
 * Category Navigation Component
 * VE-703: Marketplace category sidebar navigation
 */
import * as React from 'react';
import { useCallback } from 'react';
import type { OfferingType } from '../../../../types/marketplace';
import { OFFERING_CATEGORIES, type OfferingCategory } from '../hooks/useOfferings';

export interface CategoryNavProps {
  selectedCategory: string;
  onCategoryChange: (categoryId: string, types: OfferingType[]) => void;
  offeringCounts?: Record<string, number>;
  className?: string;
}

export function CategoryNav({
  selectedCategory,
  onCategoryChange,
  offeringCounts = {},
  className = '',
}: CategoryNavProps): JSX.Element {
  const handleCategoryClick = useCallback(
    (category: OfferingCategory) => {
      onCategoryChange(category.id, category.types);
    },
    [onCategoryChange]
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent, category: OfferingCategory) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        handleCategoryClick(category);
      }
    },
    [handleCategoryClick]
  );

  const featuredCategories = OFFERING_CATEGORIES.filter((c) => c.featured);
  const otherCategories = OFFERING_CATEGORIES.filter((c) => !c.featured && c.id !== 'all');
  const allCategory = OFFERING_CATEGORIES.find((c) => c.id === 'all');

  return (
    <nav
      className={`category-nav ${className}`}
      aria-label="Offering categories"
    >
      {/* All Offerings */}
      {allCategory && (
        <div className="category-nav__section">
          <CategoryItem
            category={allCategory}
            isSelected={selectedCategory === allCategory.id}
            count={offeringCounts['all']}
            onClick={() => handleCategoryClick(allCategory)}
            onKeyDown={(e) => handleKeyDown(e, allCategory)}
          />
        </div>
      )}

      {/* Featured Categories */}
      <div className="category-nav__section">
        <h3 className="category-nav__heading">Featured</h3>
        <ul className="category-nav__list" role="list">
          {featuredCategories.map((category) => (
            <li key={category.id}>
              <CategoryItem
                category={category}
                isSelected={selectedCategory === category.id}
                count={offeringCounts[category.id]}
                onClick={() => handleCategoryClick(category)}
                onKeyDown={(e) => handleKeyDown(e, category)}
              />
            </li>
          ))}
        </ul>
      </div>

      {/* Other Categories */}
      <div className="category-nav__section">
        <h3 className="category-nav__heading">More Categories</h3>
        <ul className="category-nav__list" role="list">
          {otherCategories.map((category) => (
            <li key={category.id}>
              <CategoryItem
                category={category}
                isSelected={selectedCategory === category.id}
                count={offeringCounts[category.id]}
                onClick={() => handleCategoryClick(category)}
                onKeyDown={(e) => handleKeyDown(e, category)}
              />
            </li>
          ))}
        </ul>
      </div>

      <style>{categoryNavStyles}</style>
    </nav>
  );
}

interface CategoryItemProps {
  category: OfferingCategory;
  isSelected: boolean;
  count?: number;
  onClick: () => void;
  onKeyDown: (e: React.KeyboardEvent) => void;
}

function CategoryItem({
  category,
  isSelected,
  count,
  onClick,
  onKeyDown,
}: CategoryItemProps): JSX.Element {
  return (
    <div
      className={`category-item ${isSelected ? 'category-item--selected' : ''}`}
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={onKeyDown}
      aria-pressed={isSelected}
      aria-label={`${category.name}${count !== undefined ? `, ${count} offerings` : ''}`}
    >
      <span className="category-item__icon" aria-hidden="true">
        <CategoryIcon name={category.icon} />
      </span>
      <span className="category-item__content">
        <span className="category-item__name">{category.name}</span>
        <span className="category-item__description">{category.description}</span>
      </span>
      {count !== undefined && (
        <span className="category-item__count">{count}</span>
      )}
    </div>
  );
}

interface CategoryIconProps {
  name: string;
}

function CategoryIcon({ name }: CategoryIconProps): JSX.Element {
  const iconMap: Record<string, JSX.Element> = {
    grid: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="3" width="7" height="7" />
        <rect x="14" y="3" width="7" height="7" />
        <rect x="14" y="14" width="7" height="7" />
        <rect x="3" y="14" width="7" height="7" />
      </svg>
    ),
    cpu: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="4" y="4" width="16" height="16" rx="2" />
        <rect x="9" y="9" width="6" height="6" />
        <path d="M9 1v3M15 1v3M9 20v3M15 20v3M20 9h3M20 14h3M1 9h3M1 14h3" />
      </svg>
    ),
    gpu: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="2" y="6" width="20" height="12" rx="2" />
        <path d="M6 10v4M10 10v4M14 10v4M18 10v4" />
      </svg>
    ),
    database: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <ellipse cx="12" cy="5" rx="9" ry="3" />
        <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
        <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
      </svg>
    ),
    kubernetes: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <path d="M12 2v4M12 18v4M2 12h4M18 12h4" />
        <circle cx="12" cy="12" r="3" />
      </svg>
    ),
    server: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="2" y="2" width="20" height="8" rx="2" ry="2" />
        <rect x="2" y="14" width="20" height="8" rx="2" ry="2" />
        <line x1="6" y1="6" x2="6.01" y2="6" />
        <line x1="6" y1="18" x2="6.01" y2="18" />
      </svg>
    ),
    settings: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
      </svg>
    ),
  };

  return iconMap[name] || iconMap['grid'];
}

const categoryNavStyles = `
  .category-nav {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 16px;
    width: 280px;
    flex-shrink: 0;
  }

  .category-nav__section {
    margin-bottom: 16px;
  }

  .category-nav__section:last-child {
    margin-bottom: 0;
  }

  .category-nav__heading {
    font-size: 0.75rem;
    font-weight: 600;
    color: #6b7280;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin: 0 0 8px;
    padding: 0 8px;
  }

  .category-nav__list {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .category-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    border-radius: 8px;
    cursor: pointer;
    transition: background-color 0.2s;
    user-select: none;
  }

  .category-item:hover {
    background: #f3f4f6;
  }

  .category-item:focus-visible {
    outline: 2px solid #3b82f6;
    outline-offset: -2px;
  }

  .category-item--selected {
    background: #eff6ff;
  }

  .category-item--selected .category-item__name {
    color: #3b82f6;
    font-weight: 500;
  }

  .category-item--selected .category-item__icon {
    color: #3b82f6;
  }

  .category-item__icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    background: #f3f4f6;
    border-radius: 8px;
    color: #6b7280;
    flex-shrink: 0;
  }

  .category-item--selected .category-item__icon {
    background: #dbeafe;
  }

  .category-item__content {
    flex: 1;
    min-width: 0;
  }

  .category-item__name {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: #111827;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .category-item__description {
    display: block;
    font-size: 0.75rem;
    color: #6b7280;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .category-item__count {
    font-size: 0.75rem;
    font-weight: 500;
    color: #6b7280;
    background: #f3f4f6;
    padding: 2px 8px;
    border-radius: 12px;
    flex-shrink: 0;
  }

  .category-item--selected .category-item__count {
    background: #dbeafe;
    color: #3b82f6;
  }

  @media (prefers-reduced-motion: reduce) {
    .category-item {
      transition: none;
    }
  }

  @media (max-width: 768px) {
    .category-nav {
      width: 100%;
    }

    .category-item__description {
      display: none;
    }
  }
`;

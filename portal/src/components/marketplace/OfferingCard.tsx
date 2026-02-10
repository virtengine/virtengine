'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import type { Offering, Provider } from '@/types/offerings';
import { CATEGORY_LABELS, CATEGORY_ICONS, STATE_COLORS } from '@/types/offerings';
import { useOfferingStore, getOfferingDisplayPrice, offeringKey } from '@/stores/offeringStore';

interface OfferingCardProps {
  offering: Offering;
}

export function OfferingCard({ offering }: OfferingCardProps) {
  const [provider, setProvider] = useState<Provider | null>(null);
  const fetchProvider = useOfferingStore((state) => state.fetchProvider);
  const compareIds = useOfferingStore((state) => state.compareIds);
  const toggleCompare = useOfferingStore((state) => state.toggleCompare);
  const key = offeringKey(offering);
  const isComparing = compareIds.includes(key);

  useEffect(() => {
    void fetchProvider(offering.id.providerAddress).then(setProvider);
  }, [offering.id.providerAddress, fetchProvider]);

  const { amount, unit } = getOfferingDisplayPrice(offering);
  const isAvailable = offering.state === 'active';
  const categoryIcon = CATEGORY_ICONS[offering.category] || '📦';
  const categoryLabel = CATEGORY_LABELS[offering.category] || offering.category;

  return (
    <div className="group relative flex flex-col rounded-lg border border-border bg-card transition-all hover:border-primary/50 hover:shadow-md active:scale-[0.99]">
      {/* Compare checkbox */}
      <button
        type="button"
        onClick={() => toggleCompare(key)}
        className={`absolute right-3 top-3 z-10 flex h-6 w-6 items-center justify-center rounded border transition-colors sm:h-5 sm:w-5 ${
          isComparing
            ? 'border-primary bg-primary text-primary-foreground'
            : 'border-border bg-background opacity-0 group-hover:opacity-100'
        }`}
        title={isComparing ? 'Remove from comparison' : 'Add to comparison'}
      >
        {isComparing && (
          <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
          </svg>
        )}
      </button>

      <Link
        href={`/marketplace/${offering.id.providerAddress}/${offering.id.sequence}`}
        className="flex flex-1 flex-col p-4"
      >
        {/* Header */}
        <div className="flex items-start justify-between">
          <span className="inline-flex items-center gap-1 rounded-full bg-primary/10 px-2 py-1 text-xs font-medium text-primary">
            <span>{categoryIcon}</span>
            {categoryLabel}
          </span>
          <span
            className={`flex items-center gap-1.5 text-xs ${isAvailable ? 'text-green-600 dark:text-green-400' : STATE_COLORS[offering.state]}`}
          >
            <span
              className={`h-2 w-2 rounded-full ${isAvailable ? 'bg-green-500' : 'bg-gray-400'}`}
            />
            {isAvailable ? 'Available' : offering.state}
          </span>
        </div>

        {/* Content */}
        <h3 className="mt-3 font-semibold group-hover:text-primary">{offering.name}</h3>

        {/* Provider */}
        <div className="mt-1 flex items-center gap-2">
          <span className="text-sm text-muted-foreground">by {provider?.name || 'Loading...'}</span>
          {provider?.verified && (
            <span className="text-xs text-blue-500" title="Verified Provider">
              ✓
            </span>
          )}
        </div>

        {/* Description */}
        <p className="mt-2 line-clamp-2 flex-1 text-sm text-muted-foreground">
          {offering.description}
        </p>

        {/* Tags */}
        {offering.tags && offering.tags.length > 0 && (
          <div className="mt-3 flex flex-wrap gap-1">
            {offering.tags.slice(0, 3).map((tag) => (
              <span
                key={tag}
                className="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground"
              >
                {tag}
              </span>
            ))}
            {offering.tags.length > 3 && (
              <span className="text-xs text-muted-foreground">+{offering.tags.length - 3}</span>
            )}
          </div>
        )}

        {/* Footer */}
        <div className="mt-4 flex items-end justify-between border-t border-border pt-3">
          <div>
            <span className="text-lg font-bold">{amount}</span>
            <span className="text-sm text-muted-foreground">{unit}</span>
          </div>

          {/* Provider reputation */}
          {provider && (
            <div className="flex items-center gap-1" title="Provider Reputation">
              <svg className="h-4 w-4 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
                <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
              </svg>
              <span className="text-sm font-medium">{provider.reputation}</span>
            </div>
          )}
        </div>

        {/* Bidding indicator */}
        {offering.allowBidding && (
          <div className="mt-2 text-xs text-muted-foreground">💰 Bidding available in Phase 3</div>
        )}
      </Link>
    </div>
  );
}

export function OfferingCardSkeleton() {
  return (
    <div className="animate-pulse rounded-lg border border-border bg-card p-4">
      <div className="flex items-start justify-between">
        <div className="h-6 w-20 rounded-full bg-muted" />
        <div className="h-4 w-16 rounded-full bg-muted" />
      </div>
      <div className="mt-3 h-5 w-3/4 rounded bg-muted" />
      <div className="mt-2 h-4 w-1/2 rounded bg-muted" />
      <div className="mt-2 h-12 rounded bg-muted" />
      <div className="mt-4 flex justify-between border-t border-border pt-3">
        <div className="h-6 w-20 rounded bg-muted" />
        <div className="h-5 w-12 rounded bg-muted" />
      </div>
    </div>
  );
}

'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import type { Offering, Provider } from '@/types/offerings';
import { CATEGORY_LABELS, CATEGORY_ICONS, STATE_COLORS } from '@/types/offerings';
import { useOfferingStore, getOfferingDisplayPrice, offeringKey } from '@/stores/offeringStore';

interface OfferingListItemProps {
  offering: Offering;
  isComparing: boolean;
  onToggleCompare: (key: string) => void;
}

export function OfferingListItem({
  offering,
  isComparing,
  onToggleCompare,
}: OfferingListItemProps) {
  const [provider, setProvider] = useState<Provider | null>(null);
  const fetchProvider = useOfferingStore((state) => state.fetchProvider);

  useEffect(() => {
    void fetchProvider(offering.id.providerAddress).then(setProvider);
  }, [offering.id.providerAddress, fetchProvider]);

  const { amount, unit } = getOfferingDisplayPrice(offering);
  const isAvailable = offering.state === 'active';
  const categoryIcon = CATEGORY_ICONS[offering.category] || '📦';
  const categoryLabel = CATEGORY_LABELS[offering.category] || offering.category;
  const key = offeringKey(offering);

  return (
    <div className="flex items-center gap-4 rounded-lg border border-border bg-card p-4 transition-all hover:border-primary/50 hover:shadow-md">
      {/* Compare checkbox */}
      <button
        type="button"
        onClick={(e) => {
          e.preventDefault();
          onToggleCompare(key);
        }}
        className={`flex h-5 w-5 shrink-0 items-center justify-center rounded border transition-colors ${
          isComparing
            ? 'border-primary bg-primary text-primary-foreground'
            : 'border-border hover:border-primary'
        }`}
        title={isComparing ? 'Remove from comparison' : 'Add to comparison'}
      >
        {isComparing && (
          <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
          </svg>
        )}
      </button>

      {/* Main content link */}
      <Link
        href={`/marketplace/${offering.id.providerAddress}/${offering.id.sequence}`}
        className="flex flex-1 flex-col gap-3 sm:flex-row sm:items-center"
      >
        {/* Category + Name */}
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center gap-1 rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
              <span>{categoryIcon}</span>
              {categoryLabel}
            </span>
            <span
              className={`flex items-center gap-1 text-xs ${isAvailable ? 'text-green-600 dark:text-green-400' : STATE_COLORS[offering.state]}`}
            >
              <span
                className={`h-1.5 w-1.5 rounded-full ${isAvailable ? 'bg-green-500' : 'bg-gray-400'}`}
              />
              {isAvailable ? 'Available' : offering.state}
            </span>
          </div>
          <h3 className="mt-1 font-semibold">{offering.name}</h3>
          <p className="mt-0.5 line-clamp-1 text-sm text-muted-foreground">
            {offering.description}
          </p>
        </div>

        {/* Provider */}
        <div className="flex shrink-0 items-center gap-2 sm:w-36">
          <span className="text-sm text-muted-foreground">{provider?.name || '...'}</span>
          {provider?.verified && (
            <span className="text-xs text-blue-500" title="Verified">
              ✓
            </span>
          )}
        </div>

        {/* Reputation */}
        <div className="flex shrink-0 items-center gap-1 sm:w-20">
          {provider && (
            <>
              <svg className="h-4 w-4 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
                <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
              </svg>
              <span className="text-sm font-medium">{provider.reputation}</span>
            </>
          )}
        </div>

        {/* Regions */}
        <div className="hidden shrink-0 sm:w-32 lg:block">
          <span className="text-xs text-muted-foreground">
            {offering.regions?.slice(0, 2).join(', ')}
            {(offering.regions?.length ?? 0) > 2 && ` +${offering.regions!.length - 2}`}
          </span>
        </div>

        {/* Price */}
        <div className="shrink-0 text-right sm:w-28">
          <span className="font-bold">{amount}</span>
          <span className="text-sm text-muted-foreground">{unit}</span>
        </div>
      </Link>
    </div>
  );
}

export function OfferingListItemSkeleton() {
  return (
    <div className="flex animate-pulse items-center gap-4 rounded-lg border border-border bg-card p-4">
      <div className="h-5 w-5 rounded bg-muted" />
      <div className="flex flex-1 items-center gap-3">
        <div className="flex-1 space-y-2">
          <div className="h-5 w-24 rounded bg-muted" />
          <div className="h-4 w-48 rounded bg-muted" />
        </div>
        <div className="h-4 w-20 rounded bg-muted" />
        <div className="h-4 w-16 rounded bg-muted" />
        <div className="h-5 w-24 rounded bg-muted" />
      </div>
    </div>
  );
}

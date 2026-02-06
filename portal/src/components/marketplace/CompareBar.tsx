'use client';

import Link from 'next/link';
import { useOfferingStore, offeringKey } from '@/stores/offeringStore';

export function CompareBar() {
  const { compareIds, offerings, clearCompare } = useOfferingStore();

  if (compareIds.length === 0) return null;

  const selectedNames = compareIds
    .map((id) => {
      const offering = offerings.find((o) => offeringKey(o) === id);
      return offering?.name ?? id;
    })
    .slice(0, 4);

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 border-t border-border bg-card shadow-lg">
      <div className="container flex items-center justify-between gap-4 py-3">
        <div className="flex flex-1 items-center gap-3 overflow-hidden">
          <span className="shrink-0 text-sm font-medium">Compare ({compareIds.length}/4):</span>
          <div className="flex gap-2 overflow-x-auto">
            {selectedNames.map((name, i) => (
              <span
                key={compareIds[i]}
                className="shrink-0 rounded-full bg-primary/10 px-3 py-1 text-xs font-medium text-primary"
              >
                {name}
              </span>
            ))}
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <button
            type="button"
            onClick={clearCompare}
            className="rounded-md border border-border px-3 py-1.5 text-sm hover:bg-accent"
          >
            Clear
          </button>
          <Link
            href={`/marketplace/compare?ids=${compareIds.join(',')}`}
            className={`rounded-md bg-primary px-4 py-1.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 ${
              compareIds.length < 2 ? 'pointer-events-none opacity-50' : ''
            }`}
            aria-disabled={compareIds.length < 2}
          >
            Compare
          </Link>
        </div>
      </div>
    </div>
  );
}

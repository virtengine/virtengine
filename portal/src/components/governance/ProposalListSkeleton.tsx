/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useMemo } from 'react';
import { Skeleton } from '@/components/ui/Skeleton';

export function ProposalListSkeleton({ count = 4 }: { count?: number }) {
  const skeletonKeys = useMemo(
    () => Array.from({ length: count }, (_, idx) => `proposal-skeleton-${count}-${idx}`),
    [count]
  );

  return (
    <div className="space-y-4">
      {skeletonKeys.map((key) => (
        <div key={key} className="rounded-lg border border-border bg-card p-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Skeleton className="h-4 w-12" />
              <Skeleton className="h-6 w-24" />
              <Skeleton className="h-6 w-28" />
            </div>
            <Skeleton className="h-4 w-20" />
          </div>
          <Skeleton className="mt-4 h-5 w-2/3" />
          <Skeleton className="mt-3 h-4 w-full" />
          <Skeleton className="mt-2 h-4 w-5/6" />
          <div className="mt-5 flex gap-3">
            <Skeleton className="h-2 w-full" />
            <Skeleton className="h-2 w-full" />
          </div>
        </div>
      ))}
    </div>
  );
}

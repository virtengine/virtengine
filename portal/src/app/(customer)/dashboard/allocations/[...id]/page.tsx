/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Metadata } from 'next';
import dynamic from 'next/dynamic';

export const metadata: Metadata = {
  title: 'Allocation Details',
  description: 'View allocation details, usage, and manage lifecycle',
};

const AllocationDetailClient = dynamic(() => import('./AllocationDetailClient'), {
  loading: () => (
    <div className="space-y-6">
      <div className="h-5 w-32 animate-pulse rounded bg-muted" />
      <div className="space-y-2">
        <div className="h-8 w-64 animate-pulse rounded bg-muted" />
        <div className="h-5 w-96 animate-pulse rounded bg-muted" />
      </div>
    </div>
  ),
});

export function generateStaticParams() {
  // Catch-all segments must be returned as an array of segments, not a joined
  // string, or Next throws "A required parameter (id) was not provided as an
  // array received string".
  return [{ id: ['_'] }];
}

export default function AllocationDetailPage() {
  return <AllocationDetailClient />;
}

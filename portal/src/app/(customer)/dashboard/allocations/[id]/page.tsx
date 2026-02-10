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
  ssr: false,
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
  return [{ id: '_' }];
}

export default function AllocationDetailPage() {
  return <AllocationDetailClient />;
}

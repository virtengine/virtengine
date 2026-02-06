/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useRouter } from 'next/navigation';
import { BillingDashboard } from '@/components/billing';

export function BillingPage() {
  const router = useRouter();

  return (
    <BillingDashboard
      onViewInvoice={(id) => router.push(`/billing/invoices/${id}`)}
      onViewAllInvoices={() => router.push('/billing/invoices')}
      onViewUsage={() => router.push('/billing/usage')}
    />
  );
}

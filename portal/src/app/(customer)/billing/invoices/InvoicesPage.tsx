/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useRouter } from 'next/navigation';
import { InvoiceList } from '@/components/billing';

export function InvoicesPage() {
  const router = useRouter();

  return <InvoiceList onViewInvoice={(id) => router.push(`/billing/invoices/${id}`)} />;
}

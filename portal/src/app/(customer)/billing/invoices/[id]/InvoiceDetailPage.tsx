/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useRouter } from 'next/navigation';
import { InvoiceDetail } from '@/components/billing';

interface InvoiceDetailPageProps {
  invoiceId: string;
}

export function InvoiceDetailPage({ invoiceId }: InvoiceDetailPageProps) {
  const router = useRouter();

  return <InvoiceDetail invoiceId={invoiceId} onBack={() => router.push('/billing/invoices')} />;
}

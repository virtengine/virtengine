/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Metadata } from 'next';
import { InvoiceDetailPage } from './InvoiceDetailPage';

export const metadata: Metadata = {
  title: 'Invoice Detail',
  description: 'View invoice details, line items, and payment history',
};

export default function InvoiceDetailRoute({ params }: { params: { id: string } }) {
  return <InvoiceDetailPage invoiceId={params.id} />;
}

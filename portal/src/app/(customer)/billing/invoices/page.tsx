/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Metadata } from 'next';
import { InvoicesPage } from './InvoicesPage';

export const metadata: Metadata = {
  title: 'Invoices',
  description: 'View and manage all invoices',
};

export default function InvoicesRoute() {
  return <InvoicesPage />;
}

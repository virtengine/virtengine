/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Metadata } from 'next';
import { BillingPage } from './BillingPage';

export const metadata: Metadata = {
  title: 'Billing',
  description: 'View billing overview, invoices, and usage analytics',
};

export default function BillingRoute() {
  return <BillingPage />;
}

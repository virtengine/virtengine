/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Metadata } from 'next';
import { EscrowPaymentsPage } from './EscrowPaymentsPage';

export const metadata: Metadata = {
  title: 'Escrow & Payments',
  description: 'Track escrow balances, deposits, settlements, and payouts.',
};

export default function EscrowPaymentsRoute() {
  return <EscrowPaymentsPage />;
}

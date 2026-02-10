/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Metadata } from 'next';
import { CustomerDashboardPage } from './CustomerDashboardPage';

export const metadata: Metadata = {
  title: 'Dashboard',
  description: 'Overview of your allocations, usage, and billing',
};

export default function DashboardRoute() {
  return <CustomerDashboardPage />;
}

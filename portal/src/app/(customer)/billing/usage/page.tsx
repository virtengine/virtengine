/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Metadata } from 'next';
import { UsagePage } from './UsagePage';

export const metadata: Metadata = {
  title: 'Usage Analytics',
  description: 'View resource usage analytics, cost breakdown, and trends',
};

export default function UsageRoute() {
  return <UsagePage />;
}

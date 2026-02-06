/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Metadata } from 'next';
import { MetricsDashboard } from '@/components/metrics/MetricsDashboard';

export const metadata: Metadata = {
  title: 'Metrics',
  description: 'Aggregated resource metrics across all deployments',
};

export default function MetricsPage() {
  return <MetricsDashboard />;
}

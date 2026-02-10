/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { ReactNode } from 'react';
import { CustomerLayout } from '@/layouts';

export default function MetricsLayout({ children }: { children: ReactNode }) {
  return <CustomerLayout>{children}</CustomerLayout>;
}

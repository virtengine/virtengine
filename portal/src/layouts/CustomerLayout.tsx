/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { ReactNode } from 'react';
import { AppLayout } from './AppLayout';

interface CustomerLayoutProps {
  children: ReactNode;
}

export function CustomerLayout({ children }: CustomerLayoutProps) {
  return <AppLayout sidebarVariant="customer">{children}</AppLayout>;
}

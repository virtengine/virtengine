/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { ReactNode } from 'react';
import { Header, Footer } from '@/components/layout';

interface MarketingLayoutProps {
  children: ReactNode;
}

export function MarketingLayout({ children }: MarketingLayoutProps) {
  return (
    <div className="flex min-h-screen flex-col bg-background text-foreground">
      <Header />
      <main id="main-content" className="flex-1">
        {children}
      </main>
      <Footer />
    </div>
  );
}

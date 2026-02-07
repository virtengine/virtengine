/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { ReactNode } from 'react';
import { Header, Footer, Navigation, Sidebar } from '@/components/layout';

interface AppLayoutProps {
  children: ReactNode;
  sidebarVariant: 'customer' | 'provider' | 'admin';
}

export function AppLayout({ children, sidebarVariant }: AppLayoutProps) {
  return (
    <div className="flex min-h-screen flex-col bg-background text-foreground">
      <Header />
      <div className="border-b border-border bg-background lg:hidden">
        <div className="container py-2">
          <Navigation className="flex w-full gap-2 overflow-x-auto" />
        </div>
      </div>
      <div className="flex flex-1">
        <div className="hidden lg:block">
          <Sidebar variant={sidebarVariant} />
        </div>
        <main id="main-content" className="flex-1">
          <div className="container py-8">{children}</div>
        </main>
      </div>
      <Footer />
    </div>
  );
}

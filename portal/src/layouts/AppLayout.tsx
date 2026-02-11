/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { ReactNode } from 'react';
import { Header, Footer, Navigation, Sidebar, MobileBottomNav } from '@/components/layout';
import { ChatPanel } from '@/components/chat/ChatPanel';

interface AppLayoutProps {
  children: ReactNode;
  sidebarVariant: 'customer' | 'provider' | 'admin';
}

export function AppLayout({ children, sidebarVariant }: AppLayoutProps) {
  return (
    <div className="flex min-h-screen flex-col bg-background text-foreground">
      <Header />
      {/* Horizontal nav strip – visible on tablets where sidebar is hidden */}
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
          {/* Extra bottom padding on mobile for bottom nav clearance */}
          <div className="container px-4 py-4 pb-20 sm:px-6 sm:py-6 md:px-8 md:py-8 md:pb-8">
            {children}
          </div>
        </main>
      </div>
      {/* Desktop footer – hidden on mobile where bottom nav replaces it */}
      <div className="hidden md:block">
        <Footer />
      </div>
      <MobileBottomNav />
      <ChatPanel />
    </div>
  );
}

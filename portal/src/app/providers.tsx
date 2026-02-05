'use client';

import { ThemeProvider } from 'next-themes';
import type { ReactNode } from 'react';
import { PortalProvider } from '@/lib/portal-adapter';
import { portalConfig, chainConfig, walletConfig } from '@/config';

/**
 * Root Providers
 *
 * Wraps the application with necessary providers:
 * - ThemeProvider: next-themes for dark/light mode
 * - PortalProvider: VirtEngine portal lib for auth, identity, marketplace, etc.
 */
export function Providers({ children }: { children: ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
      <PortalProvider config={portalConfig} chainConfig={chainConfig} walletConfig={walletConfig}>
        {children}
      </PortalProvider>
    </ThemeProvider>
  );
}

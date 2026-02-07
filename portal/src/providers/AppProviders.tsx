/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import type { ReactNode } from 'react';
import { ThemeProvider } from 'next-themes';
import { PortalProvider } from '@/lib/portal-adapter';
import { portalConfig, chainConfig, walletConfig } from '@/config';
import { Toaster } from '@/components/ui/Toaster';
import { I18nProvider } from '@/i18n/I18nProvider';
import { CosmosKitProvider } from './CosmosKitProvider';
import { ChainEventProvider } from './ChainEventProvider';

interface AppProvidersProps {
  children: ReactNode;
}

/**
 * AppProviders
 *
 * Root-level providers for the Portal app. This centralizes
 * theming, portal context, wallet connectivity and global UI.
 */
export function AppProviders({ children }: AppProvidersProps) {
  return (
    <I18nProvider>
      <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
        <CosmosKitProvider>
          <PortalProvider
            config={portalConfig}
            chainConfig={chainConfig}
            walletConfig={walletConfig}
          >
            <ChainEventProvider>{children}</ChainEventProvider>
            <Toaster />
          </PortalProvider>
        </CosmosKitProvider>
      </ThemeProvider>
    </I18nProvider>
  );
}

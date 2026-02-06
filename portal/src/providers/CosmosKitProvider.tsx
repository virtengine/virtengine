/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import * as React from 'react';
import { SUPPORTED_WALLETS, isWalletInstalled } from '@/config';
import type { WalletType } from '@/config';

export interface CosmosKitWallet {
  id: WalletType;
  name: string;
  description: string;
  recommended: boolean;
  installed: boolean;
}

interface CosmosKitContextValue {
  wallets: CosmosKitWallet[];
  refresh: () => void;
}

const CosmosKitContext = React.createContext<CosmosKitContextValue | null>(null);

function buildWallets(): CosmosKitWallet[] {
  return SUPPORTED_WALLETS.filter((wallet) => wallet.id !== 'walletconnect').map((wallet) => ({
    id: wallet.id,
    name: wallet.name,
    description: wallet.description,
    recommended: wallet.recommended,
    installed: isWalletInstalled(wallet.id),
  }));
}

/**
 * CosmosKitProvider
 *
 * Lightweight provider for Cosmos wallets (Keplr, Leap, Cosmostation).
 * This keeps wallet discovery centralized for the Portal UI and
 * complements the Portal wallet adapters.
 */
export function CosmosKitProvider({ children }: { children: React.ReactNode }) {
  const [wallets, setWallets] = React.useState<CosmosKitWallet[]>(() => buildWallets());

  const refresh = React.useCallback(() => {
    setWallets(buildWallets());
  }, []);

  React.useEffect(() => {
    refresh();
    if (typeof window === 'undefined') return;

    const handleFocus = () => refresh();
    window.addEventListener('focus', handleFocus);

    return () => {
      window.removeEventListener('focus', handleFocus);
    };
  }, [refresh]);

  const value = React.useMemo(() => ({ wallets, refresh }), [wallets, refresh]);

  return <CosmosKitContext.Provider value={value}>{children}</CosmosKitContext.Provider>;
}

export function useCosmosKitWallets(): CosmosKitContextValue {
  const context = React.useContext(CosmosKitContext);
  if (!context) {
    throw new Error('useCosmosKitWallets must be used within a CosmosKitProvider');
  }
  return context;
}
